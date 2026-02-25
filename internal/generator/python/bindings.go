package python

import (
	"fmt"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func buildOperationBindings(cfg *config.Config, doc *openapi.Document) []operationBinding {
	allOps := doc.ListOperationDetails()
	bindings := make([]operationBinding, 0)
	existingOps := map[string]struct{}{}

	for _, details := range allOps {
		existingOps[strings.ToLower(details.Method)+" "+details.Path] = struct{}{}
		if cfg.IsIgnored(details.Path, details.Method) {
			continue
		}

		mappings := cfg.FindOperationMappings(details.Path, details.Method)
		if cfg.API.GenerateOnlyMapped && len(mappings) == 0 {
			continue
		}

		if len(mappings) > 0 {
			for _, mapping := range mappings {
				mappingCopy := mapping
				for methodIndex, sdkMethod := range mapping.SDKMethods {
					pkgName, methodName, ok := config.ParseSDKMethod(sdkMethod)
					if !ok {
						continue
					}
					pkg, ok := cfg.ResolvePackage(details.Path, pkgName)
					if !ok {
						continue
					}
					order := len(bindings)
					if mappingCopy.Order > 0 {
						order = mappingCopy.Order + methodIndex
					}
					bindings = append(bindings, operationBinding{
						PackageName: normalizePackageName(pkg.Name),
						MethodName:  normalizeMethodName(methodName),
						Details:     details,
						Mapping:     &mappingCopy,
						Order:       order,
					})
				}
			}
			continue
		}

		pkg, ok := cfg.ResolvePackage(details.Path, "")
		if !ok {
			continue
		}
		bindings = append(bindings, operationBinding{
			PackageName: normalizePackageName(pkg.Name),
			MethodName:  defaultMethodName(details.OperationID, details.Path, details.Method),
			Details:     details,
			Order:       len(bindings),
		})
	}

	for _, mapping := range cfg.API.OperationMappings {
		if !mapping.AllowMissingInSwagger {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(mapping.Method)) + " " + strings.TrimSpace(mapping.Path)
		if _, ok := existingOps[key]; ok {
			continue
		}
		if cfg.IsIgnored(mapping.Path, mapping.Method) {
			continue
		}
		if len(mapping.SDKMethods) == 0 {
			continue
		}
		mappingCopy := mapping
		details := syntheticOperationDetails(mappingCopy)
		for methodIndex, sdkMethod := range mappingCopy.SDKMethods {
			pkgName, methodName, ok := config.ParseSDKMethod(sdkMethod)
			if !ok {
				continue
			}
			pkg, ok := cfg.ResolvePackage(details.Path, pkgName)
			if !ok {
				continue
			}
			order := len(bindings)
			if mappingCopy.Order > 0 {
				order = mappingCopy.Order + methodIndex
			}
			bindings = append(bindings, operationBinding{
				PackageName: normalizePackageName(pkg.Name),
				MethodName:  normalizeMethodName(methodName),
				Details:     details,
				Mapping:     &mappingCopy,
				Order:       order,
			})
		}
	}

	return deduplicateBindings(bindings)
}

func syntheticOperationDetails(mapping config.OperationMapping) openapi.OperationDetails {
	details := openapi.OperationDetails{
		Path:   strings.TrimSpace(mapping.Path),
		Method: strings.ToLower(strings.TrimSpace(mapping.Method)),
	}
	path := details.Path
	for {
		start := strings.Index(path, "{")
		if start < 0 {
			break
		}
		endOffset := strings.Index(path[start:], "}")
		if endOffset <= 1 {
			break
		}
		end := start + endOffset
		paramName := strings.TrimSpace(path[start+1 : end])
		if paramName == "" {
			path = path[end+1:]
			continue
		}
		details.PathParameters = append(details.PathParameters, openapi.ParameterSpec{
			Name:     paramName,
			In:       "path",
			Required: true,
			Schema:   &openapi.Schema{Type: "string"},
		})
		path = path[end+1:]
	}
	return details
}

func deduplicateBindings(bindings []operationBinding) []operationBinding {
	syncSeen := map[string]int{}
	asyncSeen := map[string]int{}
	nextSuffix := map[string]int{}
	for i := range bindings {
		key := bindings[i].PackageName + ":" + bindings[i].MethodName
		conflict := false
		if mappingGeneratesSync(bindings[i].Mapping) {
			syncSeen[key]++
			if syncSeen[key] > 1 {
				conflict = true
			}
		}
		if mappingGeneratesAsync(bindings[i].Mapping) {
			asyncSeen[key]++
			if asyncSeen[key] > 1 {
				conflict = true
			}
		}
		if conflict {
			suffix := nextSuffix[key]
			if suffix == 0 {
				suffix = 2
			}
			bindings[i].MethodName = fmt.Sprintf("%s_%d", bindings[i].MethodName, suffix)
			nextSuffix[key] = suffix + 1
		}
	}
	return bindings
}

func groupBindingsByPackage(bindings []operationBinding) map[string][]operationBinding {
	pkgOps := map[string][]operationBinding{}
	for _, binding := range bindings {
		pkgOps[binding.PackageName] = append(pkgOps[binding.PackageName], binding)
	}
	for pkgName := range pkgOps {
		sort.Slice(pkgOps[pkgName], func(i, j int) bool {
			return pkgOps[pkgName][i].Order < pkgOps[pkgName][j].Order
		})
	}
	return pkgOps
}

func buildPackageMeta(cfg *config.Config, packages map[string][]operationBinding) map[string]packageMeta {
	metas := map[string]packageMeta{}
	for _, pkg := range cfg.API.Packages {
		pkgCopy := pkg
		name := normalizePackageName(pkg.Name)
		dir := normalizePackageDir(pkg.SourceDir, name)
		metas[name] = packageMeta{
			Name:       name,
			ModulePath: strings.ReplaceAll(dir, "/", "."),
			DirPath:    dir,
			Package:    &pkgCopy,
		}
	}
	for name := range packages {
		if _, ok := metas[name]; ok {
			continue
		}
		metas[name] = packageMeta{
			Name:       name,
			ModulePath: name,
			DirPath:    name,
		}
	}
	return metas
}
