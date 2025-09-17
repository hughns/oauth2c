package oauth2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterClient(t *testing.T) {
	// Test successful client registration
	t.Run("successful registration", func(t *testing.T) {
		expectedResponse := ClientRegistrationResponse{
			ClientID:                  "test-client-id",
			ClientSecret:              "test-client-secret",
			ClientIDIssuedAt:          1234567890,
			ClientSecretExpiresAt:     0,
			RegistrationAccessToken:   "test-access-token",
			RegistrationClientURI:     "https://example.com/register/test-client-id",
			ClientRegistrationRequest: DefaultRegistrationMetadata(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			// Verify request body
			var req ClientRegistrationRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "oauth2c CLI Client", req.ClientName)
			assert.Equal(t, []string{"http://localhost:9876/callback"}, req.RedirectURIs)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		config := DCRConfig{
			RegistrationEndpoint: server.URL,
			InitialAccessToken:   "test-token",
			Metadata:             DefaultRegistrationMetadata(),
		}

		ctx := context.Background()
		request, response, err := RegisterClient(ctx, config, server.Client())

		require.NoError(t, err)
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, server.URL, request.URL.String())
		assert.Equal(t, expectedResponse, response)
	})

	// Test registration without initial access token
	t.Run("registration without initial access token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should not have Authorization header
			assert.Empty(t, r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(ClientRegistrationResponse{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			})
		}))
		defer server.Close()

		config := DCRConfig{
			RegistrationEndpoint: server.URL,
			Metadata:             DefaultRegistrationMetadata(),
		}

		ctx := context.Background()
		_, _, err := RegisterClient(ctx, config, server.Client())
		require.NoError(t, err)
	})

	// Test error response
	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":             "invalid_client_metadata",
				"error_description": "Invalid redirect URI",
			})
		}))
		defer server.Close()

		config := DCRConfig{
			RegistrationEndpoint: server.URL,
			Metadata:             DefaultRegistrationMetadata(),
		}

		ctx := context.Background()
		_, _, err := RegisterClient(ctx, config, server.Client())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	// Test missing registration endpoint
	t.Run("missing registration endpoint", func(t *testing.T) {
		config := DCRConfig{
			Metadata: DefaultRegistrationMetadata(),
		}

		ctx := context.Background()
		_, _, err := RegisterClient(ctx, config, &http.Client{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "registration endpoint is required")
	})
}

func TestClientFromRegistration(t *testing.T) {
	response := ClientRegistrationResponse{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		ClientRegistrationRequest: ClientRegistrationRequest{
			RedirectURIs:            []string{"https://example.com/callback"},
			GrantTypes:              []string{AuthorizationCodeGrantType, ClientCredentialsGrantType},
			TokenEndpointAuthMethod: ClientSecretPostAuthMethod,
			Scope:                   "read write",
		},
	}

	config := ClientFromRegistration(response)

	assert.Equal(t, "test-client-id", config.ClientID)
	assert.Equal(t, "test-client-secret", config.ClientSecret)
	assert.Equal(t, "https://example.com/callback", config.RedirectURL)
	assert.Equal(t, AuthorizationCodeGrantType, config.GrantType)
	assert.Equal(t, ClientSecretPostAuthMethod, config.AuthMethod)
	assert.Equal(t, []string{"read", "write"}, config.Scopes)
}

func TestDefaultRegistrationMetadata(t *testing.T) {
	metadata := DefaultRegistrationMetadata()

	assert.Equal(t, []string{"http://localhost:9876/callback"}, metadata.RedirectURIs)
	assert.Equal(t, []string{"code"}, metadata.ResponseTypes)
	assert.Equal(t, []string{AuthorizationCodeGrantType}, metadata.GrantTypes)
	assert.Equal(t, "oauth2c CLI Client", metadata.ClientName)
	assert.Equal(t, ClientSecretBasicAuthMethod, metadata.TokenEndpointAuthMethod)
}