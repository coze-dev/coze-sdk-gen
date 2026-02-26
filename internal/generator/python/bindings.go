package python

import (
	"fmt"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

var childAttributeLexicon = map[string]string{
	"app":          "apps",
	"audio":        "audio",
	"bot":          "bots",
	"dataset":      "datasets",
	"document":     "documents",
	"event":        "events",
	"execute_node": "execute_nodes",
	"file":         "files",
	"folder":       "folders",
	"image":        "images",
	"member":       "members",
	"message":      "messages",
	"room":         "rooms",
	"run_history":  "run_histories",
	"template":     "templates",
	"user":         "users",
	"variable":     "variables",
	"version":      "versions",
	"voice":        "voices",
	"workflow":     "workflows",
}

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
					pkgCopy := pkg
					bindings = append(bindings, OperationBinding{
						PackageName: NormalizePackageName(pkg.Name),
						Package:     &pkgCopy,
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
		pkgCopy := pkg
		bindings = append(bindings, OperationBinding{
			PackageName: NormalizePackageName(pkg.Name),
			Package:     &pkgCopy,
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
			pkgCopy := pkg
			bindings = append(bindings, OperationBinding{
				PackageName: NormalizePackageName(pkg.Name),
				Package:     &pkgCopy,
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
	inferredDirs := inferPackageDirsFromNames(cfg.API.Packages)
	for _, pkg := range cfg.API.Packages {
		pkgCopy := pkg
		name := NormalizePackageName(pkg.Name)
		fallbackDir := name
		if inferred, ok := inferredDirs[name]; ok && strings.TrimSpace(inferred) != "" {
			fallbackDir = inferred
		}
		dir := NormalizePackageDir(pkg.SourceDir, fallbackDir)
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
	mergeAutoInferredChildClients(metas, packages)
	return metas
}

func inferPackageDirsFromNames(packages []config.Package) map[string]string {
	names := make(map[string]struct{}, len(packages))
	for _, pkg := range packages {
		name := NormalizePackageName(pkg.Name)
		if name == "" {
			continue
		}
		names[name] = struct{}{}
	}
	cache := make(map[string]string, len(names))
	for name := range names {
		cache[name] = inferPackageDirByName(name, names, cache)
	}
	return cache
}

func inferPackageDirByName(name string, names map[string]struct{}, cache map[string]string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if cached, ok := cache[name]; ok && strings.TrimSpace(cached) != "" {
		return cached
	}
	bestParent := ""
	for candidate := range names {
		if candidate == name {
			continue
		}
		prefix := candidate + "_"
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if len(candidate) > len(bestParent) {
			bestParent = candidate
		}
	}
	if bestParent == "" {
		cache[name] = name
		return name
	}
	leaf := strings.TrimPrefix(name, bestParent+"_")
	parentDir := inferPackageDirByName(bestParent, names, cache)
	if parentDir == "" {
		cache[name] = leaf
		return leaf
	}
	if leaf == "" {
		cache[name] = parentDir
		return parentDir
	}
	cache[name] = parentDir + "/" + leaf
	return cache[name]
}

func mergeAutoInferredChildClients(metas map[string]PackageMeta, packageBindings map[string][]OperationBinding) {
	inferred := inferChildClientsByPackageHierarchy(metas)
	for pkgName, children := range inferred {
		meta, ok := metas[pkgName]
		if !ok || meta.Package == nil || len(children) == 0 {
			continue
		}
		blockedNames := collectReservedMemberNames(meta.Package, packageBindings[pkgName])
		for _, child := range children {
			attr := NormalizePythonIdentifier(strings.TrimSpace(child.Attribute))
			if attr == "" {
				continue
			}
			if _, exists := blockedNames[attr]; exists {
				continue
			}
			meta.ChildClients = append(meta.ChildClients, child)
			blockedNames[attr] = struct{}{}
		}
		metas[pkgName] = meta
	}
}

func collectReservedMemberNames(pkg *config.Package, bindings []OperationBinding) map[string]struct{} {
	reserved := map[string]struct{}{}
	if pkg == nil {
		return reserved
	}
	for _, block := range pkg.SyncExtraMethods {
		name := NormalizePythonIdentifier(strings.TrimSpace(DetectMethodBlockName(block)))
		if name == "" {
			continue
		}
		reserved[name] = struct{}{}
	}
	for _, block := range pkg.AsyncExtraMethods {
		name := NormalizePythonIdentifier(strings.TrimSpace(DetectMethodBlockName(block)))
		if name == "" {
			continue
		}
		reserved[name] = struct{}{}
	}
	for _, binding := range bindings {
		name := NormalizePythonIdentifier(strings.TrimSpace(binding.MethodName))
		if name == "" {
			continue
		}
		reserved[name] = struct{}{}
	}
	return reserved
}

func inferChildClientsByPackageHierarchy(metas map[string]PackageMeta) map[string][]childClient {
	packageNameByDir := map[string]string{}
	for pkgName, meta := range metas {
		dir := strings.Trim(strings.TrimSpace(meta.DirPath), "/")
		if dir == "" {
			continue
		}
		packageNameByDir[dir] = pkgName
	}

	result := map[string][]childClient{}
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
		attribute := preferredChildAttribute(childLeaf)
		if attribute == "" {
			continue
		}
		parentName, ok := packageNameByDir[parentDir]
		if !ok {
			continue
		}
		result[parentName] = append(result[parentName], childClient{
			Attribute: attribute,
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

func preferredChildAttribute(leaf string) string {
	key := NormalizePythonIdentifier(strings.TrimSpace(leaf))
	if key == "" {
		return ""
	}
	if preferred, ok := childAttributeLexicon[key]; ok {
		return preferred
	}
	return key
}
