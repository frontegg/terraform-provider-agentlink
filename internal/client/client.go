package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client is the Frontegg API client
type Client struct {
	baseURL    string
	clientID   string
	secret     string
	httpClient *http.Client

	mu          sync.RWMutex
	accessToken string
	tokenExpiry time.Time

	// ApplicationID stores the resolved application ID
	ApplicationID string
	// ApplicationName stores the resolved application name
	ApplicationName string
}

// AuthResponse represents the response from /auth/vendor
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
}

// Application represents a Frontegg application
type Application struct {
	ID                    string                 `json:"id"`
	VendorID              string                 `json:"vendorId"`
	Name                  string                 `json:"name"`
	AppURL                string                 `json:"appURL"`
	LoginURL              string                 `json:"loginURL"`
	LogoURL               string                 `json:"logoURL"`
	AccessType            string                 `json:"accessType"`
	IsDefault             bool                   `json:"isDefault"`
	IsActive              bool                   `json:"isActive"`
	Type                  string                 `json:"type"`
	FrontendStack         string                 `json:"frontendStack"`
	Description           string                 `json:"description"`
	CreatedAt             string                 `json:"createdAt"`
	UpdatedAt             string                 `json:"updatedAt"`
	IntegrationFinishedAt string                 `json:"integrationFinishedAt,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	AppHost               string                 `json:"appHost,omitempty"`
	AllowDcr              bool                   `json:"allowDcr"`
}

// CreateApplicationRequest represents the request to create an application
type CreateApplicationRequest struct {
	Name          string                 `json:"name"`
	AppURL        string                 `json:"appURL"`
	LoginURL      string                 `json:"loginURL"`
	LogoURL       string                 `json:"logoURL,omitempty"`
	AccessType    string                 `json:"accessType,omitempty"`
	IsDefault     *bool                  `json:"isDefault,omitempty"`
	IsActive      *bool                  `json:"isActive,omitempty"`
	Type          string                 `json:"type,omitempty"`
	FrontendStack string                 `json:"frontendStack,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	AllowDcr      *bool                  `json:"allowDcr,omitempty"`
}

// Source represents a Frontegg MCP configuration source
type Source struct {
	ID         string `json:"id"`
	VendorID   string `json:"vendorId"`
	AppID      string `json:"appId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	SourceURL  string `json:"sourceUrl"`
	Secret     string `json:"secret,omitempty"`
	APITimeout int    `json:"apiTimeout"`
	Enabled    bool   `json:"enabled"`
}

// CreateSourceRequest represents the request to create a source
type CreateSourceRequest struct {
	AppID      string `json:"appId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	SourceURL  string `json:"sourceUrl"`
	APITimeout int    `json:"apiTimeout"`
	Enabled    bool   `json:"enabled"`
}

// NewClient creates a new Frontegg API client
func NewClient(baseURL, clientID, secret string) *Client {
	return &Client{
		baseURL:  baseURL,
		clientID: clientID,
		secret:   secret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate authenticates with the Frontegg API and retrieves an access token
func (c *Client) Authenticate(ctx context.Context) error {
	authURL := fmt.Sprintf("%s/auth/vendor", c.baseURL)

	tflog.Info(ctx, "Authenticating with Frontegg API", map[string]interface{}{
		"url":       authURL,
		"client_id": c.clientID,
	})

	payload := map[string]string{
		"clientId": c.clientID,
		"secret":   c.secret,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		tflog.Error(ctx, "Failed to execute authentication request", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to execute auth request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Log the trace ID for debugging
	logTraceID(ctx, resp, "POST /auth/vendor")

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tflog.Error(ctx, "Authentication failed", map[string]interface{}{
			"status_code":       resp.StatusCode,
			"response":          string(bodyBytes),
			"frontegg_trace_id": resp.Header.Get("frontegg-trace-id"),
		})
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.mu.Lock()
	c.accessToken = authResp.Token
	// Set expiry with a buffer of 60 seconds
	c.tokenExpiry = time.Now().Add(time.Duration(authResp.ExpiresIn-60) * time.Second)
	c.mu.Unlock()

	tflog.Info(ctx, "Successfully authenticated with Frontegg API", map[string]interface{}{
		"token_expires_in_seconds": authResp.ExpiresIn,
	})

	return nil
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	token := c.accessToken
	expiry := c.tokenExpiry
	c.mu.RUnlock()

	if token != "" && time.Now().Before(expiry) {
		return token, nil
	}

	if err := c.Authenticate(ctx); err != nil {
		return "", err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken, nil
}

// logTraceID logs the frontegg-trace-id header from the response
func logTraceID(ctx context.Context, resp *http.Response, operation string) {
	traceID := resp.Header.Get("frontegg-trace-id")
	if traceID != "" {
		tflog.Info(ctx, "Frontegg API response", map[string]interface{}{
			"operation":         operation,
			"status_code":       resp.StatusCode,
			"frontegg_trace_id": traceID,
		})
	}
}

// DoRequest executes an authenticated HTTP request
func (c *Client) DoRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Log the trace ID for debugging
	logTraceID(ctx, resp, fmt.Sprintf("%s %s", method, path))

	return resp, nil
}

// GetApplications retrieves all applications
func (c *Client) GetApplications(ctx context.Context) ([]Application, error) {
	tflog.Info(ctx, "Fetching applications from Frontegg API")

	resp, err := c.DoRequest(ctx, http.MethodGet, "/applications/resources/applications/v1", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tflog.Error(ctx, "Failed to get applications", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		})
		return nil, fmt.Errorf("failed to get applications with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var applications []Application
	if err := json.NewDecoder(resp.Body).Decode(&applications); err != nil {
		return nil, fmt.Errorf("failed to decode applications response: %w", err)
	}

	// Collect application names for logging
	appNames := make([]string, len(applications))
	for i, app := range applications {
		appNames[i] = app.Name
	}

	tflog.Info(ctx, "Successfully fetched applications", map[string]interface{}{
		"count": len(applications),
		"names": appNames,
	})

	return applications, nil
}

// FindApplicationByName searches for an application by name
func (c *Client) FindApplicationByName(ctx context.Context, name string) (*Application, error) {
	applications, err := c.GetApplications(ctx)
	if err != nil {
		return nil, err
	}

	for _, app := range applications {
		if app.Name == name {
			tflog.Info(ctx, "Found application by name", map[string]interface{}{
				"name": name,
				"id":   app.ID,
			})
			return &app, nil
		}
	}

	tflog.Info(ctx, "Application not found by name", map[string]interface{}{
		"name": name,
	})
	return nil, nil
}

// CreateApplication creates a new application
func (c *Client) CreateApplication(ctx context.Context, req CreateApplicationRequest) (*Application, error) {
	tflog.Info(ctx, "Creating application", map[string]interface{}{
		"name": req.Name,
	})

	resp, err := c.DoRequest(ctx, http.MethodPost, "/applications/resources/applications/v1", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tflog.Error(ctx, "Failed to create application", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		})
		return nil, fmt.Errorf("failed to create application with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var application Application
	if err := json.NewDecoder(resp.Body).Decode(&application); err != nil {
		return nil, fmt.Errorf("failed to decode application response: %w", err)
	}

	tflog.Info(ctx, "Successfully created application", map[string]interface{}{
		"name": application.Name,
		"id":   application.ID,
	})

	return &application, nil
}

// FindOrCreateApplication finds an application by name or creates it if not found
func (c *Client) FindOrCreateApplication(ctx context.Context, name, appURL, loginURL string) (*Application, error) {
	// First, try to find the application
	app, err := c.FindApplicationByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if app != nil {
		c.ApplicationID = app.ID
		c.ApplicationName = app.Name
		return app, nil
	}

	// Application not found, create it
	tflog.Info(ctx, "Application not found, creating new application", map[string]interface{}{
		"name": name,
	})

	newApp, err := c.CreateApplication(ctx, CreateApplicationRequest{
		Name:     name,
		AppURL:   appURL,
		LoginURL: loginURL,
		Type:     "web",
	})
	if err != nil {
		return nil, err
	}

	c.ApplicationID = newApp.ID
	c.ApplicationName = newApp.Name
	return newApp, nil
}

// GetSources retrieves all sources for an application
func (c *Client) GetSources(ctx context.Context, appID string) ([]Source, error) {
	tflog.Info(ctx, "Fetching sources from Frontegg API", map[string]interface{}{
		"app_id": appID,
	})

	path := fmt.Sprintf("/app-integrations/resources/app-mcp-configuration-sources/v1?appId=%s", appID)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tflog.Error(ctx, "Failed to get sources", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		})
		return nil, fmt.Errorf("failed to get sources with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var sources []Source
	if err := json.NewDecoder(resp.Body).Decode(&sources); err != nil {
		return nil, fmt.Errorf("failed to decode sources response: %w", err)
	}

	// Collect source names for logging
	sourceNames := make([]string, len(sources))
	for i, src := range sources {
		sourceNames[i] = src.Name
	}

	tflog.Info(ctx, "Successfully fetched sources", map[string]interface{}{
		"count": len(sources),
		"names": sourceNames,
	})

	return sources, nil
}

// FindSourceByName searches for a source by name
func (c *Client) FindSourceByName(ctx context.Context, appID, name string) (*Source, error) {
	sources, err := c.GetSources(ctx, appID)
	if err != nil {
		return nil, err
	}

	for _, src := range sources {
		if src.Name == name {
			tflog.Info(ctx, "Found source by name", map[string]interface{}{
				"name": name,
				"id":   src.ID,
			})
			return &src, nil
		}
	}

	tflog.Info(ctx, "Source not found by name", map[string]interface{}{
		"name": name,
	})
	return nil, nil
}

// CreateSource creates a new source
func (c *Client) CreateSource(ctx context.Context, req CreateSourceRequest) (*Source, error) {
	tflog.Info(ctx, "Creating source", map[string]interface{}{
		"name":       req.Name,
		"type":       req.Type,
		"source_url": req.SourceURL,
	})

	resp, err := c.DoRequest(ctx, http.MethodPost, "/app-integrations/resources/app-mcp-configuration-sources/v1", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tflog.Error(ctx, "Failed to create source", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		})
		return nil, fmt.Errorf("failed to create source with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var source Source
	if err := json.NewDecoder(resp.Body).Decode(&source); err != nil {
		return nil, fmt.Errorf("failed to decode source response: %w", err)
	}

	tflog.Info(ctx, "Successfully created source", map[string]interface{}{
		"name": source.Name,
		"id":   source.ID,
	})

	return &source, nil
}

// FindOrCreateSource finds a source by name or creates it if not found
func (c *Client) FindOrCreateSource(ctx context.Context, appID, name, sourceType, sourceURL string, apiTimeout int) (*Source, error) {
	// First, try to find the source
	src, err := c.FindSourceByName(ctx, appID, name)
	if err != nil {
		return nil, err
	}

	if src != nil {
		return src, nil
	}

	// Source not found, create it
	tflog.Info(ctx, "Source not found, creating new source", map[string]interface{}{
		"name": name,
	})

	newSource, err := c.CreateSource(ctx, CreateSourceRequest{
		AppID:      appID,
		Name:       name,
		Type:       sourceType,
		SourceURL:  sourceURL,
		APITimeout: apiTimeout,
		Enabled:    true,
	})
	if err != nil {
		return nil, err
	}

	return newSource, nil
}

// InternalTool represents a tool returned from import or used in upsert
type InternalTool struct {
	ID                 string                 `json:"id,omitempty"`
	VendorID           string                 `json:"vendorId,omitempty"`
	AppID              string                 `json:"appId,omitempty"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	OriginalMethod     string                 `json:"originalMethod,omitempty"`
	OriginalPath       string                 `json:"originalPath,omitempty"`
	IsActive           bool                   `json:"isActive"`
	Schema             map[string]interface{} `json:"schema,omitempty"`
	AuthenticationType string                 `json:"authenticationType,omitempty"`
	SourceID           string                 `json:"sourceId,omitempty"`
}

// UpsertToolsRequest represents the request to upsert tools
type UpsertToolsRequest struct {
	AppID    string         `json:"appId"`
	ToolType string         `json:"toolType,omitempty"`
	Tools    []InternalTool `json:"tools"`
}

// ImportOpenAPISchema imports tools from an OpenAPI specification file
func (c *Client) ImportOpenAPISchema(ctx context.Context, appID string, schemaContent []byte, filename string) ([]InternalTool, error) {
	tflog.Info(ctx, "Importing OpenAPI schema", map[string]interface{}{
		"app_id":   appID,
		"filename": filename,
	})

	return c.importSchema(ctx, appID, schemaContent, filename, "openapi", "/app-integrations/resources/internal-tools/v1/openapi/import")
}

// ImportGraphQLSchema imports tools from a GraphQL schema file
func (c *Client) ImportGraphQLSchema(ctx context.Context, appID string, schemaContent []byte, filename string) ([]InternalTool, error) {
	tflog.Info(ctx, "Importing GraphQL schema", map[string]interface{}{
		"app_id":   appID,
		"filename": filename,
	})

	return c.importSchema(ctx, appID, schemaContent, filename, "graphql", "/app-integrations/resources/internal-tools/v1/graphql/import")
}

// importSchema is a helper function for importing schemas via multipart form
func (c *Client) importSchema(ctx context.Context, appID string, schemaContent []byte, filename, fieldName, endpoint string) ([]InternalTool, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add appId field
	if err := writer.WriteField("appId", appID); err != nil {
		return nil, fmt.Errorf("failed to write appId field: %w", err)
	}

	// Add file field with the appropriate field name (openapi or graphql)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(schemaContent); err != nil {
		return nil, fmt.Errorf("failed to write schema content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create import request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute import request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Log the trace ID for debugging
	logTraceID(ctx, resp, fmt.Sprintf("POST %s", endpoint))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tflog.Error(ctx, "Failed to import schema", map[string]interface{}{
			"status_code":       resp.StatusCode,
			"response":          string(bodyBytes),
			"frontegg_trace_id": resp.Header.Get("frontegg-trace-id"),
		})
		return nil, fmt.Errorf("failed to import schema with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tools []InternalTool
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return nil, fmt.Errorf("failed to decode import response: %w", err)
	}

	tflog.Info(ctx, "Successfully imported schema", map[string]interface{}{
		"tools_count": len(tools),
	})

	return tools, nil
}

// UpsertTools creates or updates multiple tools
func (c *Client) UpsertTools(ctx context.Context, req UpsertToolsRequest) ([]InternalTool, error) {
	tflog.Info(ctx, "Upserting tools", map[string]interface{}{
		"app_id":      req.AppID,
		"tool_type":   req.ToolType,
		"tools_count": len(req.Tools),
	})

	resp, err := c.DoRequest(ctx, http.MethodPost, "/app-integrations/resources/internal-tools/v1/upsert", req)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert tools: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tflog.Error(ctx, "Failed to upsert tools", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		})
		return nil, fmt.Errorf("failed to upsert tools with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tools []InternalTool
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return nil, fmt.Errorf("failed to decode upsert response: %w", err)
	}

	tflog.Info(ctx, "Successfully upserted tools", map[string]interface{}{
		"tools_count": len(tools),
	})

	return tools, nil
}

// ImportAndUpsertSchema imports a schema and then upserts the resulting tools
func (c *Client) ImportAndUpsertSchema(ctx context.Context, appID, sourceID, sourceType string, schemaContent []byte, filename string) error {
	var tools []InternalTool
	var err error

	// Import schema based on source type
	switch sourceType {
	case "REST":
		tools, err = c.ImportOpenAPISchema(ctx, appID, schemaContent, filename)
	case "GRAPHQL":
		tools, err = c.ImportGraphQLSchema(ctx, appID, schemaContent, filename)
	default:
		return fmt.Errorf("schema import not supported for source type: %s", sourceType)
	}

	if err != nil {
		return fmt.Errorf("failed to import schema: %w", err)
	}

	if len(tools) == 0 {
		tflog.Info(ctx, "No tools found in schema, skipping upsert")
		return nil
	}

	// Set sourceId on all tools
	for i := range tools {
		tools[i].SourceID = sourceID
	}

	// Upsert the tools
	_, err = c.UpsertTools(ctx, UpsertToolsRequest{
		AppID:    appID,
		ToolType: sourceType,
		Tools:    tools,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert tools: %w", err)
	}

	return nil
}

// ============================================================================
// Application CRUD Methods
// ============================================================================

// UpdateApplicationRequest represents the request to update an application
type UpdateApplicationRequest struct {
	Name        string `json:"name,omitempty"`
	AppURL      string `json:"appURL,omitempty"`
	LoginURL    string `json:"loginURL,omitempty"`
	LogoURL     string `json:"logoURL,omitempty"`
	AccessType  string `json:"accessType,omitempty"`
	IsDefault   *bool  `json:"isDefault,omitempty"`
	IsActive    *bool  `json:"isActive,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	AllowDcr    *bool  `json:"allowDcr,omitempty"`
}

// GetApplicationByID retrieves an application by ID
func (c *Client) GetApplicationByID(ctx context.Context, id string) (*Application, error) {
	tflog.Info(ctx, "Fetching application by ID", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/applications/resources/applications/v1/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get application with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var application Application
	if err := json.NewDecoder(resp.Body).Decode(&application); err != nil {
		return nil, fmt.Errorf("failed to decode application response: %w", err)
	}

	return &application, nil
}

// UpdateApplication updates an existing application
func (c *Client) UpdateApplication(ctx context.Context, id string, req UpdateApplicationRequest) (*Application, error) {
	tflog.Info(ctx, "Updating application", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/applications/resources/applications/v1/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update application with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Fetch the updated application
	return c.GetApplicationByID(ctx, id)
}

// DeleteApplication deletes an application
func (c *Client) DeleteApplication(ctx context.Context, id string) error {
	tflog.Info(ctx, "Deleting application", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/applications/resources/applications/v1/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete application with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// ============================================================================
// MCP Configuration Methods
// ============================================================================

// McpConfiguration represents MCP configuration for an app
type McpConfiguration struct {
	ID         string `json:"id"`
	VendorID   string `json:"vendorId"`
	AppID      string `json:"appId"`
	BaseURL    string `json:"baseUrl"`
	APITimeout int    `json:"apiTimeout"`
}

// CreateOrUpdateMcpConfigurationRequest represents the request to create/update MCP config
type CreateOrUpdateMcpConfigurationRequest struct {
	AppID      string `json:"appId"`
	BaseURL    string `json:"baseUrl"`
	APITimeout int    `json:"apiTimeout"`
}

// CreateOrUpdateMcpConfiguration creates or updates MCP configuration
func (c *Client) CreateOrUpdateMcpConfiguration(ctx context.Context, req CreateOrUpdateMcpConfigurationRequest) (*McpConfiguration, error) {
	tflog.Info(ctx, "Creating/updating MCP configuration", map[string]interface{}{
		"app_id":   req.AppID,
		"base_url": req.BaseURL,
	})

	resp, err := c.DoRequest(ctx, http.MethodPost, "/app-integrations/resources/app-mcp-configurations/v1", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update MCP configuration: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create/update MCP configuration with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config McpConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode MCP configuration response: %w", err)
	}

	return &config, nil
}

// GetMcpConfiguration retrieves MCP configuration for an app
func (c *Client) GetMcpConfiguration(ctx context.Context, appID string) (*McpConfiguration, error) {
	tflog.Info(ctx, "Fetching MCP configuration", map[string]interface{}{
		"app_id": appID,
	})

	path := fmt.Sprintf("/app-integrations/resources/app-mcp-configurations/v1?appId=%s", appID)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP configuration: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get MCP configuration with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config McpConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode MCP configuration response: %w", err)
	}

	return &config, nil
}

// ============================================================================
// Source CRUD Methods (additional methods)
// ============================================================================

// UpdateSourceRequest represents the request to update a source
type UpdateSourceRequest struct {
	AppID      string `json:"appId"`
	Name       string `json:"name,omitempty"`
	Type       string `json:"type,omitempty"`
	SourceURL  string `json:"sourceUrl,omitempty"`
	APITimeout int    `json:"apiTimeout,omitempty"`
	Enabled    *bool  `json:"enabled,omitempty"`
}

// GetSourceByID retrieves a source by ID
func (c *Client) GetSourceByID(ctx context.Context, appID, sourceID string) (*Source, error) {
	sources, err := c.GetSources(ctx, appID)
	if err != nil {
		return nil, err
	}

	for _, src := range sources {
		if src.ID == sourceID {
			return &src, nil
		}
	}

	return nil, nil
}

// UpdateSource updates an existing source
func (c *Client) UpdateSource(ctx context.Context, sourceID string, req UpdateSourceRequest) (*Source, error) {
	tflog.Info(ctx, "Updating source", map[string]interface{}{
		"id": sourceID,
	})

	path := fmt.Sprintf("/app-integrations/resources/app-mcp-configuration-sources/v1/%s", sourceID)
	resp, err := c.DoRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update source: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update source with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var source Source
	if err := json.NewDecoder(resp.Body).Decode(&source); err != nil {
		return nil, fmt.Errorf("failed to decode source response: %w", err)
	}

	return &source, nil
}

// DeleteSource deletes a source
func (c *Client) DeleteSource(ctx context.Context, appID, sourceID string) error {
	tflog.Info(ctx, "Deleting source", map[string]interface{}{
		"id": sourceID,
	})

	path := fmt.Sprintf("/app-integrations/resources/app-mcp-configuration-sources/v1/%s?appId=%s", sourceID, appID)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete source with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// ============================================================================
// Policy Types and Methods
// ============================================================================

// PolicyCondition represents a condition in policy targeting
type PolicyCondition struct {
	Attribute string                 `json:"attribute"`
	Negate    bool                   `json:"negate"`
	Op        string                 `json:"op"`
	Value     map[string]interface{} `json:"value"`
}

// PolicyIfBlock represents the if block in targeting
type PolicyIfBlock struct {
	Conditions []PolicyCondition `json:"conditions"`
}

// PolicyThenBlock represents the then block in targeting
type PolicyThenBlock struct {
	Result         string `json:"result"`
	ApprovalFlowID string `json:"approvalFlowId,omitempty"`
}

// PolicyTargeting represents policy targeting rules
type PolicyTargeting struct {
	If   PolicyIfBlock   `json:"if"`
	Then PolicyThenBlock `json:"then"`
}

// MaskingPolicyConfiguration represents the configuration for data masking
type MaskingPolicyConfiguration struct {
	CreditCard      bool `json:"creditCard,omitempty"`
	EmailAddress    bool `json:"emailAddress,omitempty"`
	PhoneNumber     bool `json:"phoneNumber,omitempty"`
	IpAddress       bool `json:"ipAddress,omitempty"`
	UsSsn           bool `json:"usSsn,omitempty"`
	UsDriverLicense bool `json:"usDriverLicense,omitempty"`
	UsPassport      bool `json:"usPassport,omitempty"`
	UsItin          bool `json:"usItin,omitempty"`
	UsBankNumber    bool `json:"usBankNumber,omitempty"`
	IbanCode        bool `json:"ibanCode,omitempty"`
	SwiftCode       bool `json:"swiftCode,omitempty"`
	BitcoinAddress  bool `json:"bitcoinAddress,omitempty"`
	EthereumAddress bool `json:"ethereumAddress,omitempty"`
	CvvCvc          bool `json:"cvvCvc,omitempty"`
	Url             bool `json:"url,omitempty"`
}

// Policy represents a generic policy response
type Policy struct {
	ID                  string                      `json:"id"`
	VendorID            string                      `json:"vendorId"`
	Name                string                      `json:"name"`
	Description         string                      `json:"description,omitempty"`
	Type                string                      `json:"type"`
	Enabled             bool                        `json:"enabled"`
	AppIDs              []string                    `json:"appIds,omitempty"`
	TenantID            string                      `json:"tenantId,omitempty"`
	InternalToolIDs     []string                    `json:"internalToolIds,omitempty"`
	Targeting           *PolicyTargeting            `json:"targeting,omitempty"`
	Keys                []string                    `json:"keys,omitempty"`
	PolicyConfiguration *MaskingPolicyConfiguration `json:"policyConfiguration,omitempty"`
	Metadata            map[string]interface{}      `json:"metadata,omitempty"`
	CreatedAt           string                      `json:"createdAt,omitempty"`
	UpdatedAt           string                      `json:"updatedAt,omitempty"`
}

// CreateConditionalPolicyRequest represents the request to create a conditional policy
type CreateConditionalPolicyRequest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Enabled         bool                   `json:"enabled"`
	AppIDs          []string               `json:"appIds,omitempty"`
	TenantID        string                 `json:"tenantId,omitempty"`
	InternalToolIDs []string               `json:"internalToolIds"`
	Targeting       *PolicyTargeting       `json:"targeting,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRbacPolicyRequest represents the request to create an RBAC policy
type CreateRbacPolicyRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	Enabled         bool     `json:"enabled"`
	AppIDs          []string `json:"appIds,omitempty"`
	TenantID        string   `json:"tenantId,omitempty"`
	InternalToolIDs []string `json:"internalToolIds"`
	Type            string   `json:"type"` // "RBAC_ROLES" or "RBAC_PERMISSIONS"
	Keys            []string `json:"keys"`
}

// CreateMaskingPolicyRequest represents the request to create a masking policy
type CreateMaskingPolicyRequest struct {
	Name                string                      `json:"name"`
	Description         string                      `json:"description,omitempty"`
	Enabled             bool                        `json:"enabled"`
	AppIDs              []string                    `json:"appIds,omitempty"`
	TenantID            string                      `json:"tenantId,omitempty"`
	InternalToolIDs     []string                    `json:"internalToolIds"`
	Targeting           *PolicyTargeting            `json:"targeting,omitempty"`
	PolicyConfiguration *MaskingPolicyConfiguration `json:"policyConfiguration"`
	Metadata            map[string]interface{}      `json:"metadata,omitempty"`
}

// UpdateConditionalPolicyRequest represents the request to update a conditional policy
type UpdateConditionalPolicyRequest struct {
	Name            string                 `json:"name,omitempty"`
	Description     string                 `json:"description,omitempty"`
	Enabled         *bool                  `json:"enabled,omitempty"`
	AppIDs          []string               `json:"appIds,omitempty"`
	TenantID        string                 `json:"tenantId,omitempty"`
	InternalToolIDs []string               `json:"internalToolIds,omitempty"`
	Targeting       *PolicyTargeting       `json:"targeting,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateRbacPolicyRequest represents the request to update an RBAC policy
type UpdateRbacPolicyRequest struct {
	Name            string   `json:"name,omitempty"`
	Description     string   `json:"description,omitempty"`
	Enabled         *bool    `json:"enabled,omitempty"`
	AppIDs          []string `json:"appIds,omitempty"`
	TenantID        string   `json:"tenantId,omitempty"`
	InternalToolIDs []string `json:"internalToolIds,omitempty"`
	Keys            []string `json:"keys,omitempty"`
}

// UpdateMaskingPolicyRequest represents the request to update a masking policy
type UpdateMaskingPolicyRequest struct {
	Name                string                      `json:"name,omitempty"`
	Description         string                      `json:"description,omitempty"`
	Enabled             *bool                       `json:"enabled,omitempty"`
	AppIDs              []string                    `json:"appIds,omitempty"`
	TenantID            string                      `json:"tenantId,omitempty"`
	InternalToolIDs     []string                    `json:"internalToolIds,omitempty"`
	Targeting           *PolicyTargeting            `json:"targeting,omitempty"`
	PolicyConfiguration *MaskingPolicyConfiguration `json:"policyConfiguration,omitempty"`
	Metadata            map[string]interface{}      `json:"metadata,omitempty"`
}

// ============================================================================
// Conditional Policy CRUD
// ============================================================================

// CreateConditionalPolicy creates a new conditional policy
func (c *Client) CreateConditionalPolicy(ctx context.Context, req CreateConditionalPolicyRequest) (*Policy, error) {
	tflog.Info(ctx, "Creating conditional policy", map[string]interface{}{
		"name": req.Name,
	})

	resp, err := c.DoRequest(ctx, http.MethodPost, "/app-integrations/resources/policies/v1", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create conditional policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create conditional policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode policy response: %w", err)
	}

	// Fetch the full policy
	return c.GetConditionalPolicy(ctx, result.ID)
}

// GetConditionalPolicy retrieves a conditional policy by ID
func (c *Client) GetConditionalPolicy(ctx context.Context, id string) (*Policy, error) {
	tflog.Info(ctx, "Fetching conditional policy", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/app-integrations/resources/policies/v1/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get conditional policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get conditional policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return nil, fmt.Errorf("failed to decode policy response: %w", err)
	}

	return &policy, nil
}

// UpdateConditionalPolicy updates an existing conditional policy
func (c *Client) UpdateConditionalPolicy(ctx context.Context, id string, req UpdateConditionalPolicyRequest) (*Policy, error) {
	tflog.Info(ctx, "Updating conditional policy", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/app-integrations/resources/policies/v1/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update conditional policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update conditional policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Fetch the updated policy
	return c.GetConditionalPolicy(ctx, id)
}

// DeletePolicy deletes any policy by ID (shared across all policy types)
func (c *Client) DeletePolicy(ctx context.Context, id string) error {
	tflog.Info(ctx, "Deleting policy", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/app-integrations/resources/policies/v1/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// ============================================================================
// RBAC Policy CRUD
// ============================================================================

// CreateRbacPolicy creates a new RBAC policy
func (c *Client) CreateRbacPolicy(ctx context.Context, req CreateRbacPolicyRequest) (*Policy, error) {
	tflog.Info(ctx, "Creating RBAC policy", map[string]interface{}{
		"name": req.Name,
		"type": req.Type,
	})

	resp, err := c.DoRequest(ctx, http.MethodPost, "/app-integrations/resources/policies/v1/rbac", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create RBAC policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create RBAC policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode policy response: %w", err)
	}

	// Fetch the full policy
	return c.GetRbacPolicy(ctx, result.ID)
}

// GetRbacPolicy retrieves an RBAC policy by ID
func (c *Client) GetRbacPolicy(ctx context.Context, id string) (*Policy, error) {
	tflog.Info(ctx, "Fetching RBAC policy", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/app-integrations/resources/policies/v1/rbac/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get RBAC policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get RBAC policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return nil, fmt.Errorf("failed to decode policy response: %w", err)
	}

	return &policy, nil
}

// UpdateRbacPolicy updates an existing RBAC policy
func (c *Client) UpdateRbacPolicy(ctx context.Context, id string, req UpdateRbacPolicyRequest) (*Policy, error) {
	tflog.Info(ctx, "Updating RBAC policy", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/app-integrations/resources/policies/v1/rbac/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update RBAC policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update RBAC policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Fetch the updated policy
	return c.GetRbacPolicy(ctx, id)
}

// ============================================================================
// Masking Policy CRUD
// ============================================================================

// CreateMaskingPolicy creates a new masking policy
func (c *Client) CreateMaskingPolicy(ctx context.Context, req CreateMaskingPolicyRequest) (*Policy, error) {
	tflog.Info(ctx, "Creating masking policy", map[string]interface{}{
		"name": req.Name,
	})

	resp, err := c.DoRequest(ctx, http.MethodPost, "/app-integrations/resources/policies/v1/masking", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create masking policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create masking policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode policy response: %w", err)
	}

	// Fetch the full policy
	return c.GetMaskingPolicy(ctx, result.ID)
}

// GetMaskingPolicy retrieves a masking policy by ID
func (c *Client) GetMaskingPolicy(ctx context.Context, id string) (*Policy, error) {
	tflog.Info(ctx, "Fetching masking policy", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/app-integrations/resources/policies/v1/masking/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get masking policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get masking policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return nil, fmt.Errorf("failed to decode policy response: %w", err)
	}

	return &policy, nil
}

// UpdateMaskingPolicy updates an existing masking policy
func (c *Client) UpdateMaskingPolicy(ctx context.Context, id string, req UpdateMaskingPolicyRequest) (*Policy, error) {
	tflog.Info(ctx, "Updating masking policy", map[string]interface{}{
		"id": id,
	})

	path := fmt.Sprintf("/app-integrations/resources/policies/v1/masking/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update masking policy: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update masking policy with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Fetch the updated policy
	return c.GetMaskingPolicy(ctx, id)
}

// ============================================================================
// Tools Methods (additional)
// ============================================================================

// DeleteToolsBySource deletes all tools associated with a source
func (c *Client) DeleteToolsBySource(ctx context.Context, appID, sourceID string) error {
	tflog.Info(ctx, "Deleting tools by source", map[string]interface{}{
		"app_id":    appID,
		"source_id": sourceID,
	})

	// Get tools for this source and delete them
	path := fmt.Sprintf("/app-integrations/resources/internal-tools/v1?appId=%s&sourceId=%s", appID, sourceID)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("failed to get tools: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get tools with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Items []InternalTool `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode tools response: %w", err)
	}

	// Delete each tool
	for _, tool := range result.Items {
		if err := c.DeleteTool(ctx, appID, tool.ID); err != nil {
			tflog.Warn(ctx, "Failed to delete tool", map[string]interface{}{
				"tool_id": tool.ID,
				"error":   err.Error(),
			})
		}
	}

	return nil
}

// DeleteTool deletes a single tool
func (c *Client) DeleteTool(ctx context.Context, appID, toolID string) error {
	path := fmt.Sprintf("/app-integrations/resources/internal-tools/v1/%s?appId=%s", toolID, appID)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete tool: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete tool with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// VendorConfig represents the vendor configuration response
type VendorConfig struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

// IdentityConfiguration represents the identity configuration response
type IdentityConfiguration struct {
	ID                      string `json:"id"`
	DefaultTokenExpiration  int    `json:"defaultTokenExpiration"`
}

// UpdateIdentityConfigurationRequest represents the request to update identity configuration
type UpdateIdentityConfigurationRequest struct {
	DefaultTokenExpiration *int `json:"defaultTokenExpiration,omitempty"`
}

// UpdateAllowedOriginsRequest represents the request to update allowed origins
type UpdateAllowedOriginsRequest struct {
	AllowedOrigins []string `json:"allowedOrigins"`
}

// GetVendorConfig retrieves the vendor configuration
func (c *Client) GetVendorConfig(ctx context.Context) (*VendorConfig, error) {
	tflog.Info(ctx, "Fetching vendor configuration")

	resp, err := c.DoRequest(ctx, http.MethodGet, "/vendors", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get vendor config: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get vendor config with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config VendorConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode vendor config response: %w", err)
	}

	tflog.Info(ctx, "Successfully fetched vendor configuration", map[string]interface{}{
		"vendor_id":       config.ID,
		"allowed_origins": config.AllowedOrigins,
	})

	return &config, nil
}

// UpdateAllowedOrigins updates the vendor's allowed origins
func (c *Client) UpdateAllowedOrigins(ctx context.Context, origins []string) (*VendorConfig, error) {
	tflog.Info(ctx, "Updating allowed origins", map[string]interface{}{
		"origins": origins,
	})

	req := UpdateAllowedOriginsRequest{
		AllowedOrigins: origins,
	}

	resp, err := c.DoRequest(ctx, http.MethodPut, "/vendors", req)
	if err != nil {
		return nil, fmt.Errorf("failed to update allowed origins: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update allowed origins with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config VendorConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode vendor config response: %w", err)
	}

	tflog.Info(ctx, "Successfully updated allowed origins", map[string]interface{}{
		"vendor_id":       config.ID,
		"allowed_origins": config.AllowedOrigins,
	})

	return &config, nil
}

// ============================================================================
// Identity Configuration Methods
// ============================================================================

// GetIdentityConfiguration retrieves the identity configuration
func (c *Client) GetIdentityConfiguration(ctx context.Context) (*IdentityConfiguration, error) {
	tflog.Info(ctx, "Fetching identity configuration")

	// Use POST with empty body to get current config (addOrUpdate returns current state)
	resp, err := c.DoRequest(ctx, http.MethodPost, "/identity/resources/configurations/v1", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get identity configuration: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get identity configuration with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config IdentityConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode identity configuration response: %w", err)
	}

	tflog.Info(ctx, "Successfully fetched identity configuration", map[string]interface{}{
		"id":                       config.ID,
		"default_token_expiration": config.DefaultTokenExpiration,
	})

	return &config, nil
}

// UpdateIdentityConfiguration updates the identity configuration
func (c *Client) UpdateIdentityConfiguration(ctx context.Context, req UpdateIdentityConfigurationRequest) (*IdentityConfiguration, error) {
	tflog.Info(ctx, "Updating identity configuration", map[string]interface{}{
		"default_token_expiration": req.DefaultTokenExpiration,
	})

	resp, err := c.DoRequest(ctx, http.MethodPost, "/identity/resources/configurations/v1", req)
	if err != nil {
		return nil, fmt.Errorf("failed to update identity configuration: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update identity configuration with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config IdentityConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode identity configuration response: %w", err)
	}

	tflog.Info(ctx, "Successfully updated identity configuration", map[string]interface{}{
		"id":                       config.ID,
		"default_token_expiration": config.DefaultTokenExpiration,
	})

	return &config, nil
}
