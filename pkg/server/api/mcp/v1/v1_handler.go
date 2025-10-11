package v1

import (
	"net/http"

	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/go-chi/chi/v5"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
)

type (
	resolver struct {
		proboSvc  *probo.Service
		usrmgrSvc *usrmgr.Service
		logger    *log.Logger
	}
)

func NewMux(logger *log.Logger, proboSvc *probo.Service, usrmgrSvc *usrmgr.Service, cfg Config) *chi.Mux {
	logger.Info("initializing MCP server",
		log.String("version", cfg.Version),
		log.String("request_timeout", cfg.RequestTimeout.String()),
	)

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "probo",
			Title:   "Probo",
			Version: cfg.Version,
		},
		&mcp.ServerOptions{},
	)

	resolver := &resolver{
		proboSvc:  proboSvc,
		usrmgrSvc: usrmgrSvc,
		logger:    logger,
	}

	mcp.AddTool(
		server,
		&mcp.Tool{
			Title:       "List Organizations",
			Description: "List all organizations the user has access to",
			Name:        "listOrganizations",
			Annotations: &mcp.ToolAnnotations{
				Title:        "List Organizations",
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:       "object",
				Properties: map[string]*jsonschema.Schema{},
			},
			OutputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"result": {
						Type: "array",
						Items: &jsonschema.Schema{
							Type:     "object",
							Required: []string{"name", "id", "tenantID"},
							Properties: map[string]*jsonschema.Schema{
								"name": {
									Type:        "string",
									Description: "The organization name",
								},
								"id": {
									Type:        "string",
									Description: "The organization ID",
								},
								"tenantID": {
									Type:        "string",
									Description: "The tenant ID this organization belongs to",
								},
							},
						},
					},
				},
			},
		},
		resolver.ListOrganizations,
	)

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
				Type:     "object",
				Required: []string{"organizationID", "size", "orderField"},
				Properties: map[string]*jsonschema.Schema{
					"organizationID": {
						Type:        "string",
						Description: "The organization ID to list vendors for",
					},
					"orderField": {
						Type:        "string",
						Description: "Field to order results by",
						Enum: []any{
							"NAME",
							"CREATED_AT",
							"UPDATED_AT",
						},
					},
					"cursor": {
						Type:        "string",
						Description: "Cursor for pagination",
					},
					"size": {
						Type:        "integer",
						Description: "Number of results to return",
						Minimum:     jsonschema.Ptr(float64(1)),
						Maximum:     jsonschema.Ptr(float64(1000)),
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
				Type:     "object",
				Required: []string{"organizationID", "name"},
				Properties: map[string]*jsonschema.Schema{
					"organizationID": {
						Type:        "string",
						Description: "The organization ID to add the vendor to",
					},
					"name": {
						Type:        "string",
						Description: "The vendor name",
						MinLength:   jsonschema.Ptr(1),
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

	// Wrap handler with authentication middleware
	authHandler := WithMCPAuth(logger, usrmgrSvc, cfg.Auth, handler)

	r := chi.NewMux()
	r.Handle("/", authHandler)

	logger.Info("MCP server initialized successfully")

	return r
}
