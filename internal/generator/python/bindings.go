package python

import (
	"fmt"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func buildOperationBindings(cfg *config.Config, doc *openapi.Document) []OperationBinding {
	allOps := doc.ListOperationDetails()
	bindings := make([]OperationBinding, 0)
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
					bindings = append(bindings, OperationBinding{
						PackageName: NormalizePackageName(pkg.Name),
						MethodName:  NormalizeMethodName(methodName),
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
		bindings = append(bindings, OperationBinding{
			PackageName: NormalizePackageName(pkg.Name),
			MethodName:  DefaultMethodName(details.OperationID, details.Path, details.Method),
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
			bindings = append(bindings, OperationBinding{
				PackageName: NormalizePackageName(pkg.Name),
				MethodName:  NormalizeMethodName(methodName),
				Details:     details,
				Mapping:     &mappingCopy,
				Order:       order,
			})
		}
	}

	return DeduplicateBindings(bindings)
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

func DeduplicateBindings(bindings []OperationBinding) []OperationBinding {
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

func groupBindingsByPackage(bindings []OperationBinding) map[string][]OperationBinding {
	pkgOps := map[string][]OperationBinding{}
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

func buildPackageMeta(cfg *config.Config, packages map[string][]OperationBinding) map[string]PackageMeta {
	metas := map[string]PackageMeta{}
	for _, pkg := range cfg.API.Packages {
		pkgCopy := pkg
		name := NormalizePackageName(pkg.Name)
		dir := NormalizePackageDir(pkg.SourceDir, name)
		metas[name] = PackageMeta{
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
		metas[name] = PackageMeta{
			Name:       name,
			ModulePath: name,
			DirPath:    name,
		}
	}
	mergeAutoInferredChildClients(metas)
	return metas
}

func mergeAutoInferredChildClients(metas map[string]PackageMeta) {
	inferred := inferChildClientsByPackageHierarchy(metas)
	for pkgName, children := range inferred {
		meta, ok := metas[pkgName]
		if !ok || meta.Package == nil || len(children) == 0 {
			continue
		}
		existingByAttr := map[string]struct{}{}
		for _, child := range meta.Package.ChildClients {
			attr := NormalizePythonIdentifier(strings.TrimSpace(child.Attribute))
			if attr == "" {
				continue
			}
			existingByAttr[attr] = struct{}{}
		}
		for _, child := range children {
			attr := NormalizePythonIdentifier(strings.TrimSpace(child.Attribute))
			if attr == "" {
				continue
			}
			if _, exists := existingByAttr[attr]; exists {
				continue
			}
			meta.Package.ChildClients = append(meta.Package.ChildClients, child)
			existingByAttr[attr] = struct{}{}
		}
		metas[pkgName] = meta
	}
}

func inferChildClientsByPackageHierarchy(metas map[string]PackageMeta) map[string][]config.ChildClient {
	packageNameByDir := map[string]string{}
	for pkgName, meta := range metas {
		dir := strings.Trim(strings.TrimSpace(meta.DirPath), "/")
		if dir == "" {
			continue
		}
		packageNameByDir[dir] = pkgName
	}

	result := map[string][]config.ChildClient{}
	for _, childMeta := range metas {
		childDir := strings.Trim(strings.TrimSpace(childMeta.DirPath), "/")
		if childDir == "" {
			continue
		}
		slash := strings.LastIndex(childDir, "/")
		if slash <= 0 || slash >= len(childDir)-1 {
			continue
		}
		parentDir := childDir[:slash]
		childLeaf := childDir[slash+1:]
		parentName, ok := packageNameByDir[parentDir]
		if !ok {
			continue
		}
		result[parentName] = append(result[parentName], config.ChildClient{
			Attribute: childLeaf,
			Module:    "." + childLeaf,
			SyncClass: packageClientClassName(childMeta, false),
			AsyncClass: packageClientClassName(
				childMeta,
				true,
			),
		})
	}

	for pkgName := range result {
		sort.Slice(result[pkgName], func(i, j int) bool {
			return result[pkgName][i].Attribute < result[pkgName][j].Attribute
		})
	}
	return result
}
