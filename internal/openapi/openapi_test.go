package openapi

import (
	"path/filepath"
	"testing"
)

func TestLoadAndLookupOperations(t *testing.T) {
	doc, err := Load(filepath.Join("testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	op, ok := doc.Operation("post", "/v3/chat")
	if !ok {
		t.Fatalf("expected /v3/chat#post to exist")
	}
	if op.OperationID != "OpenApiChat" {
		t.Fatalf("unexpected operation id: %q", op.OperationID)
	}

	if !doc.HasOperation("GET", "/v3/chat/message/list") {
		t.Fatalf("expected GET /v3/chat/message/list to exist")
	}
	if doc.HasOperation("patch", "/v3/chat") {
		t.Fatalf("did not expect PATCH /v3/chat")
	}
	if _, ok := doc.Operation("invalid", "/v3/chat"); ok {
		t.Fatalf("did not expect invalid method to exist")
	}
}

func TestListOperationsSorted(t *testing.T) {
	doc, err := Load(filepath.Join("testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	ops := doc.ListOperations()
	if len(ops) != 5 {
		t.Fatalf("expected 5 operations, got %d", len(ops))
	}

	for i := 1; i < len(ops); i++ {
		prev := ops[i-1]
		curr := ops[i]
		if prev.Path > curr.Path || (prev.Path == curr.Path && prev.Method > curr.Method) {
			t.Fatalf("operations are not sorted at %d: %#v > %#v", i, prev, curr)
		}
	}
}

func TestPathsWithPrefix(t *testing.T) {
	doc, err := Load(filepath.Join("testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	chatPaths := doc.PathsWithPrefix("/v3/chat")
	if len(chatPaths) != 3 {
		t.Fatalf("expected 3 chat paths, got %d", len(chatPaths))
	}

	missing := doc.PathsWithPrefix("/v2/not-found")
	if len(missing) != 0 {
		t.Fatalf("expected no paths, got %v", missing)
	}
}

func TestOperationDetailsAndSchemaResolution(t *testing.T) {
	doc, err := Load(filepath.Join("testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	details, ok := doc.OperationDetails("/v3/chat", "post")
	if !ok {
		t.Fatal("expected operation details")
	}
	if details.RequestBodySchema == nil {
		t.Fatal("expected request body schema")
	}
	if details.ResponseSchema == nil {
		t.Fatal("expected response schema")
	}
	if details.RequestBodyContentType != "application/json" {
		t.Fatalf("unexpected request content type: %s", details.RequestBodyContentType)
	}
	if details.ResponseContentType != "application/json" {
		t.Fatalf("unexpected response content type: %s", details.ResponseContentType)
	}
	if len(details.QueryParameters) != 1 {
		t.Fatalf("expected 1 query param, got %d", len(details.QueryParameters))
	}

	workspaceDetails, ok := doc.OperationDetails("/v1/workspaces/{workspace_id}", "get")
	if !ok {
		t.Fatal("expected workspace details")
	}
	if len(workspaceDetails.PathParameters) != 1 {
		t.Fatalf("expected 1 path param, got %d", len(workspaceDetails.PathParameters))
	}
	if workspaceDetails.PathParameters[0].Name != "workspace_id" {
		t.Fatalf("unexpected path param name: %s", workspaceDetails.PathParameters[0].Name)
	}

	if _, ok := doc.ResolveSchemaRef("#/components/schemas/OpenApiChatReq"); !ok {
		t.Fatal("expected to resolve OpenApiChatReq")
	}
	if _, ok := doc.ResolveSchemaRef("#/components/schemas/NotExist"); ok {
		t.Fatal("did not expect to resolve missing schema")
	}

	refs := doc.CollectSchemaRefsFromOperation(*details)
	if len(refs) == 0 {
		t.Fatal("expected schema refs from operation")
	}
	foundReq := false
	foundResp := false
	for _, ref := range refs {
		if ref == "OpenApiChatReq" {
			foundReq = true
		}
		if ref == "OpenApiChatResp" {
			foundResp = true
		}
	}
	if !foundReq || !foundResp {
		t.Fatalf("expected req/resp refs, got %v", refs)
	}
}

func TestOperationDetailsKeepsDescriptions(t *testing.T) {
	doc, err := Parse([]byte(`
openapi: 3.0.0
paths:
  /v1/demo/{id}:
    get:
      operationId: GetDemo
      summary: get demo
      description: get demo detail
      parameters:
        - name: id
          in: path
          required: true
          description: demo id
          schema:
            type: string
      responses:
        "200":
          description: demo response
          content:
            application/json:
              schema:
                type: object
components: {}
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	details, ok := doc.OperationDetails("/v1/demo/{id}", "get")
	if !ok {
		t.Fatal("expected operation details")
	}
	if details.Description != "get demo detail" {
		t.Fatalf("unexpected operation description: %q", details.Description)
	}
	if len(details.PathParameters) != 1 {
		t.Fatalf("expected one path parameter, got %d", len(details.PathParameters))
	}
	if details.PathParameters[0].Description != "demo id" {
		t.Fatalf("unexpected path parameter description: %q", details.PathParameters[0].Description)
	}
	if details.Response == nil || details.Response.Description != "demo response" {
		t.Fatalf("unexpected response description: %#v", details.Response)
	}
}

func TestListOperationDetails(t *testing.T) {
	doc, err := Load(filepath.Join("testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	details := doc.ListOperationDetails()
	if len(details) != 5 {
		t.Fatalf("expected 5 operation details, got %d", len(details))
	}
}

func TestParseInvalidYAML(t *testing.T) {
	if _, err := Parse([]byte("paths: [")); err == nil {
		t.Fatal("expected Parse() to fail for invalid yaml")
	}
}

func TestOperationSupportsAllMethodsAndNilDoc(t *testing.T) {
	doc, err := Parse([]byte(`
paths:
  /all:
    get: {operationId: getOp}
    put: {operationId: putOp}
    post: {operationId: postOp}
    delete: {operationId: deleteOp}
    patch: {operationId: patchOp}
    options: {operationId: optionsOp}
    head: {operationId: headOp}
    trace: {operationId: traceOp}
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	methods := []string{"get", "put", "post", "delete", "patch", "options", "head", "trace"}
	for _, method := range methods {
		if _, ok := doc.Operation(method, "/all"); !ok {
			t.Fatalf("expected method %s to exist", method)
		}
	}

	var nilDoc *Document
	if _, ok := nilDoc.Operation("get", "/all"); ok {
		t.Fatal("did not expect operation on nil doc")
	}
	if len(nilDoc.ListOperations()) != 0 {
		t.Fatal("expected nil list operations from nil doc")
	}
	if len(nilDoc.PathsWithPrefix("/")) != 0 {
		t.Fatal("expected nil paths from nil doc")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join("testdata", "missing.yaml")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSelectResponseAndContentTypeHelpers(t *testing.T) {
	response := selectResponse(map[string]*Response{
		"default": {Description: "default"},
		"201":     {Description: "created"},
		"200":     {Description: "ok"},
	})
	if response == nil || response.Description != "ok" {
		t.Fatalf("expected 200 response first, got %#v", response)
	}

	response = selectResponse(map[string]*Response{
		"418": {Description: "teapot"},
		"500": {Description: "server"},
	})
	if response == nil || response.Description != "teapot" {
		t.Fatalf("expected smallest numeric code response, got %#v", response)
	}

	response = selectResponse(map[string]*Response{
		"x-custom": {Description: "custom"},
		"a-custom": {Description: "custom2"},
	})
	if response == nil || response.Description != "custom2" {
		t.Fatalf("expected lexical fallback response, got %#v", response)
	}

	contentType, _ := selectContentType(map[string]*MediaType{
		"text/event-stream": {},
		"application/json":  {},
	})
	if contentType != "application/json" {
		t.Fatalf("expected application/json priority, got %s", contentType)
	}

	contentType, _ = selectContentType(map[string]*MediaType{
		"application/xml": {},
	})
	if contentType != "application/xml" {
		t.Fatalf("expected lexical content-type fallback, got %s", contentType)
	}

	contentType, mediaType := selectContentType(nil)
	if contentType != "" || mediaType != nil {
		t.Fatalf("expected empty content-type and nil media, got %q %#v", contentType, mediaType)
	}
}

func TestResolveHelpersWithRefs(t *testing.T) {
	doc, err := Parse([]byte(`
components:
  parameters:
    q:
      name: q
      in: query
      schema:
        type: string
  requestBodies:
    Body:
      content:
        application/json:
          schema:
            type: object
            properties:
              id:
                type: string
  responses:
    Resp:
      content:
        application/json:
          schema:
            type: object
            properties:
              ok:
                type: boolean
  schemas:
    Item:
      type: object
      properties:
        id:
          type: string
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	param := doc.resolveParameter(&Parameter{Ref: "#/components/parameters/q"})
	if param == nil || param.Name != "q" {
		t.Fatalf("unexpected resolved param: %#v", param)
	}
	if got := doc.resolveParameter(&Parameter{Ref: "#/components/parameters/missing"}); got == nil || got.Ref == "" {
		t.Fatalf("expected unresolved param to be returned as-is, got %#v", got)
	}

	body := doc.resolveRequestBody(&RequestBody{Ref: "#/components/requestBodies/Body"})
	if body == nil || len(body.Content) != 1 {
		t.Fatalf("unexpected resolved body: %#v", body)
	}
	resp := doc.resolveResponse(&Response{Ref: "#/components/responses/Resp"})
	if resp == nil || len(resp.Content) != 1 {
		t.Fatalf("unexpected resolved response: %#v", resp)
	}

	if name, ok := doc.SchemaName(&Schema{Ref: "#/components/schemas/Item"}); !ok || name != "Item" {
		t.Fatalf("unexpected schema name from ref: name=%q ok=%v", name, ok)
	}
	if name, ok := doc.SchemaName(doc.Components.Schemas["Item"]); !ok || name != "Item" {
		t.Fatalf("unexpected schema name from pointer: name=%q ok=%v", name, ok)
	}
	if _, ok := doc.SchemaName(&Schema{}); ok {
		t.Fatal("did not expect schema name for anonymous schema")
	}

	if _, ok := refName("#/components/schemas/Item", "#/components/schemas/"); !ok {
		t.Fatal("expected refName success")
	}
	if _, ok := refName("Item", "#/components/schemas/"); ok {
		t.Fatal("did not expect refName success for invalid ref")
	}
}

func TestSchemaNameDeterministicForAliasedPointers(t *testing.T) {
	shared := &Schema{Type: "object"}
	doc := &Document{
		Components: Components{
			Schemas: map[string]*Schema{
				"z_alias": shared,
				"a_alias": shared,
				"m_alias": shared,
			},
		},
	}

	for i := 0; i < 128; i++ {
		name, ok := doc.SchemaName(shared)
		if !ok {
			t.Fatal("expected schema name for aliased pointer")
		}
		if name != "a_alias" {
			t.Fatalf("expected deterministic lexical alias, got %q", name)
		}
	}
}
