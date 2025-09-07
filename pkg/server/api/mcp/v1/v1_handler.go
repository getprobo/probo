package v1

import (
	"encoding/json"
	"net/http"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/go-chi/chi/v5"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	resolver struct {
		proboSvc       *probo.TenantService
		organizationID gid.GID
	}
)

func NewMux(proboSvc *probo.Service) *chi.Mux {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "probo",
			Title:   "Probo",
			Version: "1.0.0", // todo retrieve from build info
		},
		&mcp.ServerOptions{},
	)

	tenantID, err := gid.ParseTenantID("lXdXZSh-AAE")
	if err != nil {
		panic(err)
	}

	organizationID, err := gid.ParseGID("lXdXZSh-AAEAAAAAAZfLJi38a0AGbu37")
	if err != nil {
		panic(err)
	}

	resolver := &resolver{proboSvc: proboSvc.WithTenant(tenantID), organizationID: organizationID}

	mcp.AddTool(
		server,
		&mcp.Tool{
			Title:       "List Vendors",
			Description: "List all vendors for the organization",
			Name:        "listVendors",
			Annotations: &mcp.ToolAnnotations{
				Title:        "List Vendors",
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"orderField": {
						Type:    "string",
						Default: json.RawMessage(`"NAME"`),
						Enum: []any{
							"NAME",
							"CREATED_AT",
							"UPDATED_AT",
						},
					},
					"cursor": {
						Type: "string",
					},
					"size": {
						Type:    "integer",
						Minimum: jsonschema.Ptr(float64(1)),
						Maximum: jsonschema.Ptr(float64(1000)),
						Default: json.RawMessage(`100`),
					},
				},
			},
			OutputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"result": {
						Type: "array",
						Items: &jsonschema.Schema{
							Type:     "object",
							Required: []string{"name", "id"},
							Properties: map[string]*jsonschema.Schema{
								"name": {
									Type: "string",
								},
								"id": {
									Type: "string",
								},
							},
						},
					},
				},
			},
		},
		resolver.ListVendors,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "addVendor",
			Description: "Add a vendor",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name": {
						Type:     "string",
						Required: []string{"name"},
					},
					"description": {
						Type: "string",
					},
					"headquarterAddress": {
						Type: "string",
					},
					"legalName": {
						Type: "string",
					},
					"websiteURL": {
						Type:   "string",
						Format: "uri",
					},
					"category": {
						Type: "string",
					},
					"privacyPolicyURL": {
						Type: "string",
					},
					"serviceLevelAgreementURL": {
						Type:   "string",
						Format: "uri",
					},
					"dataProcessingAgreementURL": {
						Type:   "string",
						Format: "uri",
					},
					"businessAssociateAgreementURL": {
						Type:   "string",
						Format: "uri",
					},
					"subprocessorsListURL": {
						Type:   "string",
						Format: "uri",
					},
					"certifications": {
						Type: "array",
						Items: &jsonschema.Schema{
							Type: "string",
						},
					},
					"securityPageURL": {
						Type:   "string",
						Format: "uri",
					},
					"trustPageURL": {
						Type:   "string",
						Format: "uri",
					},
					"termsOfServiceURL": {
						Type:   "string",
						Format: "uri",
					},
					"statusPageURL": {
						Type:   "string",
						Format: "uri",
					},
					"businessOwnerID": {
						Type: "string",
					},
					"securityOwnerID": {
						Type: "string",
					},
				},
			},
			OutputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"result": {
						Type:     "object",
						Required: []string{"name", "id"},
						Properties: map[string]*jsonschema.Schema{
							"name": {
								Type: "string",
							},
							"id": {
								Type: "string",
							},
						},
					},
				},
			},
		},
		resolver.AddVendor,
	)

	getServer := func(r *http.Request) *mcp.Server { return server }

	handler := mcp.NewStreamableHTTPHandler(
		getServer,
		&mcp.StreamableHTTPOptions{Stateless: true},
	)

	r := chi.NewMux()
	r.Handle("/", handler)

	return r
}
