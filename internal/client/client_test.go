package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://api.example.com", "client-id", "secret")

	if c.baseURL != "https://api.example.com" {
		t.Errorf("expected baseURL to be 'https://api.example.com', got '%s'", c.baseURL)
	}
	if c.clientID != "client-id" {
		t.Errorf("expected clientID to be 'client-id', got '%s'", c.clientID)
	}
	if c.secret != "secret" {
		t.Errorf("expected secret to be 'secret', got '%s'", c.secret)
	}
}

func TestAuthenticate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/vendor" {
			t.Errorf("expected path '/auth/vendor', got '%s'", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got '%s'", r.Method)
		}

		// Verify request body
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["clientId"] != "test-client" {
			t.Errorf("expected clientId 'test-client', got '%s'", body["clientId"])
		}
		if body["secret"] != "test-secret" {
			t.Errorf("expected secret 'test-secret', got '%s'", body["secret"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Token:     "test-token",
			ExpiresIn: 3600,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-client", "test-secret")
	err := c.Authenticate(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c.accessToken != "test-token" {
		t.Errorf("expected accessToken 'test-token', got '%s'", c.accessToken)
	}
}

func TestAuthenticateFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid credentials"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "bad-client", "bad-secret")
	err := c.Authenticate(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetApplications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/applications/resources/applications/v1":
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			apps := []Application{
				{ID: "app-1", Name: "App One", VendorID: "vendor-1"},
				{ID: "app-2", Name: "App Two", VendorID: "vendor-1"},
			}
			json.NewEncoder(w).Encode(apps)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	apps, err := c.GetApplications(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(apps) != 2 {
		t.Errorf("expected 2 apps, got %d", len(apps))
	}
	if apps[0].Name != "App One" {
		t.Errorf("expected first app name 'App One', got '%s'", apps[0].Name)
	}
}

func TestCreateApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/applications/resources/applications/v1":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req CreateApplicationRequest
			json.NewDecoder(r.Body).Decode(&req)

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Application{
				ID:       "new-app-id",
				Name:     req.Name,
				AppURL:   req.AppURL,
				LoginURL: req.LoginURL,
				VendorID: "vendor-1",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	app, err := c.CreateApplication(context.Background(), CreateApplicationRequest{
		Name:     "Test App",
		AppURL:   "https://app.test.com",
		LoginURL: "https://app.test.com/login",
		Type:     "agent",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.ID != "new-app-id" {
		t.Errorf("expected ID 'new-app-id', got '%s'", app.ID)
	}
	if app.Name != "Test App" {
		t.Errorf("expected name 'Test App', got '%s'", app.Name)
	}
}

func TestGetApplicationByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/applications/resources/applications/v1/app-123":
			json.NewEncoder(w).Encode(Application{
				ID:       "app-123",
				Name:     "My App",
				VendorID: "vendor-1",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	app, err := c.GetApplicationByID(context.Background(), "app-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.ID != "app-123" {
		t.Errorf("expected ID 'app-123', got '%s'", app.ID)
	}
}

func TestGetApplicationByIDNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	app, err := c.GetApplicationByID(context.Background(), "nonexistent")

	if err != nil {
		t.Fatalf("expected no error for not found, got %v", err)
	}
	if app != nil {
		t.Errorf("expected nil app, got %+v", app)
	}
}

func TestUpdateApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/applications/resources/applications/v1/app-123":
			if r.Method == http.MethodPatch {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			} else if r.Method == http.MethodGet {
				json.NewEncoder(w).Encode(Application{
					ID:       "app-123",
					Name:     "Updated App",
					VendorID: "vendor-1",
				})
			}
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	app, err := c.UpdateApplication(context.Background(), "app-123", UpdateApplicationRequest{
		Name: "Updated App",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Name != "Updated App" {
		t.Errorf("expected name 'Updated App', got '%s'", app.Name)
	}
}

func TestDeleteApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/applications/resources/applications/v1/app-123":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	err := c.DeleteApplication(context.Background(), "app-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGetSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/app-mcp-configuration-sources/v1":
			if r.URL.Query().Get("appId") != "app-123" {
				t.Errorf("expected appId query param 'app-123', got '%s'", r.URL.Query().Get("appId"))
			}
			sources := []Source{
				{ID: "src-1", Name: "Source One", Type: "REST"},
				{ID: "src-2", Name: "Source Two", Type: "GRAPHQL"},
			}
			json.NewEncoder(w).Encode(sources)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	sources, err := c.GetSources(context.Background(), "app-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(sources))
	}
}

func TestCreateSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/app-mcp-configuration-sources/v1":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req CreateSourceRequest
			json.NewDecoder(r.Body).Decode(&req)

			json.NewEncoder(w).Encode(Source{
				ID:        "new-src-id",
				Name:      req.Name,
				Type:      req.Type,
				SourceURL: req.SourceURL,
				AppID:     req.AppID,
				VendorID:  "vendor-1",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	source, err := c.CreateSource(context.Background(), CreateSourceRequest{
		AppID:      "app-123",
		Name:       "My Source",
		Type:       "REST",
		SourceURL:  "https://api.example.com",
		APITimeout: 3000,
		Enabled:    true,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if source.ID != "new-src-id" {
		t.Errorf("expected ID 'new-src-id', got '%s'", source.ID)
	}
}

func TestUpdateSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/app-mcp-configuration-sources/v1/src-123":
			if r.Method != http.MethodPatch {
				t.Errorf("expected PATCH, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(Source{
				ID:        "src-123",
				Name:      "Updated Source",
				Type:      "REST",
				SourceURL: "https://updated.example.com",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	source, err := c.UpdateSource(context.Background(), "src-123", UpdateSourceRequest{
		Name:      "Updated Source",
		SourceURL: "https://updated.example.com",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if source.Name != "Updated Source" {
		t.Errorf("expected name 'Updated Source', got '%s'", source.Name)
	}
}

func TestDeleteSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/app-mcp-configuration-sources/v1/src-123":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Query().Get("appId") != "app-123" {
				t.Errorf("expected appId 'app-123', got '%s'", r.URL.Query().Get("appId"))
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	err := c.DeleteSource(context.Background(), "app-123", "src-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateOrUpdateMcpConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/app-mcp-configurations/v1":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req CreateOrUpdateMcpConfigurationRequest
			json.NewDecoder(r.Body).Decode(&req)

			json.NewEncoder(w).Encode(McpConfiguration{
				ID:         "mcp-config-id",
				AppID:      req.AppID,
				BaseURL:    req.BaseURL,
				APITimeout: req.APITimeout,
				VendorID:   "vendor-1",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	config, err := c.CreateOrUpdateMcpConfiguration(context.Background(), CreateOrUpdateMcpConfigurationRequest{
		AppID:      "app-123",
		BaseURL:    "https://api.example.com",
		APITimeout: 5000,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config.ID != "mcp-config-id" {
		t.Errorf("expected ID 'mcp-config-id', got '%s'", config.ID)
	}
	if config.BaseURL != "https://api.example.com" {
		t.Errorf("expected BaseURL 'https://api.example.com', got '%s'", config.BaseURL)
	}
}

func TestGetMcpConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/app-mcp-configurations/v1":
			if r.URL.Query().Get("appId") != "app-123" {
				t.Errorf("expected appId 'app-123', got '%s'", r.URL.Query().Get("appId"))
			}
			json.NewEncoder(w).Encode(McpConfiguration{
				ID:         "mcp-config-id",
				AppID:      "app-123",
				BaseURL:    "https://api.example.com",
				APITimeout: 5000,
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	config, err := c.GetMcpConfiguration(context.Background(), "app-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config.ID != "mcp-config-id" {
		t.Errorf("expected ID 'mcp-config-id', got '%s'", config.ID)
	}
}

func TestCreateConditionalPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/policies/v1":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"id": "policy-123"})
		case "/app-integrations/resources/policies/v1/policy-123":
			json.NewEncoder(w).Encode(Policy{
				ID:      "policy-123",
				Name:    "Test Policy",
				Type:    "CONDITIONAL",
				Enabled: true,
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	policy, err := c.CreateConditionalPolicy(context.Background(), CreateConditionalPolicyRequest{
		Name:            "Test Policy",
		Enabled:         true,
		InternalToolIDs: []string{},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if policy.ID != "policy-123" {
		t.Errorf("expected ID 'policy-123', got '%s'", policy.ID)
	}
}

func TestGetConditionalPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/policies/v1/policy-123":
			json.NewEncoder(w).Encode(Policy{
				ID:      "policy-123",
				Name:    "Test Policy",
				Type:    "CONDITIONAL",
				Enabled: true,
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	policy, err := c.GetConditionalPolicy(context.Background(), "policy-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if policy.Name != "Test Policy" {
		t.Errorf("expected name 'Test Policy', got '%s'", policy.Name)
	}
}

func TestCreateRbacPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/policies/v1/rbac":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req CreateRbacPolicyRequest
			json.NewDecoder(r.Body).Decode(&req)

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"id": "rbac-policy-123"})
		case "/app-integrations/resources/policies/v1/rbac/rbac-policy-123":
			json.NewEncoder(w).Encode(Policy{
				ID:      "rbac-policy-123",
				Name:    "Admin RBAC",
				Type:    "RBAC_ROLES",
				Enabled: true,
				Keys:    []string{"admin"},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	policy, err := c.CreateRbacPolicy(context.Background(), CreateRbacPolicyRequest{
		Name:            "Admin RBAC",
		Enabled:         true,
		Type:            "RBAC_ROLES",
		Keys:            []string{"admin"},
		InternalToolIDs: []string{"tool-1"},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if policy.ID != "rbac-policy-123" {
		t.Errorf("expected ID 'rbac-policy-123', got '%s'", policy.ID)
	}
}

func TestGetRbacPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/policies/v1/rbac/rbac-123":
			json.NewEncoder(w).Encode(Policy{
				ID:   "rbac-123",
				Name: "RBAC Policy",
				Type: "RBAC_ROLES",
				Keys: []string{"admin", "editor"},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	policy, err := c.GetRbacPolicy(context.Background(), "rbac-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(policy.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(policy.Keys))
	}
}

func TestCreateMaskingPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/policies/v1/masking":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req CreateMaskingPolicyRequest
			json.NewDecoder(r.Body).Decode(&req)

			if !req.PolicyConfiguration.CreditCard {
				t.Error("expected credit card masking to be enabled")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"id": "masking-policy-123"})
		case "/app-integrations/resources/policies/v1/masking/masking-policy-123":
			json.NewEncoder(w).Encode(Policy{
				ID:      "masking-policy-123",
				Name:    "PII Masking",
				Type:    "MASKING",
				Enabled: true,
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	policy, err := c.CreateMaskingPolicy(context.Background(), CreateMaskingPolicyRequest{
		Name:            "PII Masking",
		Enabled:         true,
		InternalToolIDs: []string{},
		PolicyConfiguration: &MaskingPolicyConfiguration{
			CreditCard:   true,
			EmailAddress: true,
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if policy.ID != "masking-policy-123" {
		t.Errorf("expected ID 'masking-policy-123', got '%s'", policy.ID)
	}
}

func TestGetMaskingPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/policies/v1/masking/masking-123":
			json.NewEncoder(w).Encode(Policy{
				ID:      "masking-123",
				Name:    "Masking Policy",
				Type:    "MASKING",
				Enabled: true,
				PolicyConfiguration: &MaskingPolicyConfiguration{
					CreditCard:   true,
					EmailAddress: true,
				},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	policy, err := c.GetMaskingPolicy(context.Background(), "masking-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if policy.PolicyConfiguration == nil {
		t.Fatal("expected policy configuration, got nil")
	}
	if !policy.PolicyConfiguration.CreditCard {
		t.Error("expected credit card masking to be enabled")
	}
}

func TestDeletePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/policies/v1/policy-123":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	err := c.DeletePolicy(context.Background(), "policy-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpsertTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/internal-tools/v1/upsert":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req UpsertToolsRequest
			json.NewDecoder(r.Body).Decode(&req)

			if len(req.Tools) != 2 {
				t.Errorf("expected 2 tools, got %d", len(req.Tools))
			}

			json.NewEncoder(w).Encode([]InternalTool{
				{ID: "tool-1", Name: "Tool 1"},
				{ID: "tool-2", Name: "Tool 2"},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	tools, err := c.UpsertTools(context.Background(), UpsertToolsRequest{
		AppID:    "app-123",
		ToolType: "REST",
		Tools: []InternalTool{
			{Name: "Tool 1", Description: "First tool"},
			{Name: "Tool 2", Description: "Second tool"},
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

func TestDeleteTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/internal-tools/v1/tool-123":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Query().Get("appId") != "app-123" {
				t.Errorf("expected appId 'app-123', got '%s'", r.URL.Query().Get("appId"))
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	err := c.DeleteTool(context.Background(), "app-123", "tool-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFindApplicationByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/applications/resources/applications/v1":
			apps := []Application{
				{ID: "app-1", Name: "First App"},
				{ID: "app-2", Name: "Target App"},
				{ID: "app-3", Name: "Third App"},
			}
			json.NewEncoder(w).Encode(apps)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")

	// Test finding existing app
	app, err := c.FindApplicationByName(context.Background(), "Target App")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app == nil {
		t.Fatal("expected app, got nil")
	}
	if app.ID != "app-2" {
		t.Errorf("expected ID 'app-2', got '%s'", app.ID)
	}

	// Test not finding app
	app, err = c.FindApplicationByName(context.Background(), "Nonexistent App")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app != nil {
		t.Errorf("expected nil, got %+v", app)
	}
}

func TestFindSourceByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/vendor":
			json.NewEncoder(w).Encode(AuthResponse{Token: "token", ExpiresIn: 3600})
		case "/app-integrations/resources/app-mcp-configuration-sources/v1":
			sources := []Source{
				{ID: "src-1", Name: "First Source"},
				{ID: "src-2", Name: "Target Source"},
			}
			json.NewEncoder(w).Encode(sources)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "client", "secret")
	source, err := c.FindSourceByName(context.Background(), "app-123", "Target Source")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if source == nil {
		t.Fatal("expected source, got nil")
	}
	if source.ID != "src-2" {
		t.Errorf("expected ID 'src-2', got '%s'", source.ID)
	}
}
