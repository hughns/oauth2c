package oauth2

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// Dynamic Client Registration types and constants per RFC 7591

// ClientRegistrationRequest represents the request body for client registration per RFC 7591
type ClientRegistrationRequest struct {
	// OPTIONAL. Array of redirection URI strings for use in redirect-based flows
	RedirectURIs []string `json:"redirect_uris,omitempty"`

	// OPTIONAL. JSON array containing a list of the OAuth 2.0 response_type values
	ResponseTypes []string `json:"response_types,omitempty"`

	// OPTIONAL. JSON array containing a list of the OAuth 2.0 grant type values
	GrantTypes []string `json:"grant_types,omitempty"`

	// OPTIONAL. String containing a human-readable name of the client
	ClientName string `json:"client_name,omitempty"`

	// OPTIONAL. String containing an URI of the client
	ClientURI string `json:"client_uri,omitempty"`

	// OPTIONAL. String containing an URI of the client logo
	LogoURI string `json:"logo_uri,omitempty"`

	// OPTIONAL. String containing a space-separated list of scope values
	Scope string `json:"scope,omitempty"`

	// OPTIONAL. Array of e-mail addresses of people responsible for this client
	Contacts []string `json:"contacts,omitempty"`

	// OPTIONAL. String containing an URI that points to a human-readable terms of service document
	TosURI string `json:"tos_uri,omitempty"`

	// OPTIONAL. String containing an URI that points to a human-readable privacy policy document
	PolicyURI string `json:"policy_uri,omitempty"`

	// OPTIONAL. String containing an URI referencing the client's JSON Web Key Set document
	JwksURI string `json:"jwks_uri,omitempty"`

	// OPTIONAL. Client's JSON Web Key Set document value
	Jwks interface{} `json:"jwks,omitempty"`

	// OPTIONAL. String containing an URI for the client's software identifier
	SoftwareID string `json:"software_id,omitempty"`

	// OPTIONAL. String containing a version identifier string for the client software
	SoftwareVersion string `json:"software_version,omitempty"`

	// Client authentication method for the token endpoint
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method,omitempty"`

	// JWS algorithm for signing JWTs used to authenticate the client at the token endpoint
	TokenEndpointAuthSigningAlg string `json:"token_endpoint_auth_signing_alg,omitempty"`
}

// ClientRegistrationResponse represents the response from client registration per RFC 7591
type ClientRegistrationResponse struct {
	// REQUIRED. OAuth 2.0 client identifier string
	ClientID string `json:"client_id"`

	// OPTIONAL. OAuth 2.0 client secret string
	ClientSecret string `json:"client_secret,omitempty"`

	// OPTIONAL. Time at which the client identifier was issued
	ClientIDIssuedAt int64 `json:"client_id_issued_at,omitempty"`

	// CONDITIONAL. Time at which the client secret will expire or 0 if it will not expire
	ClientSecretExpiresAt int64 `json:"client_secret_expires_at,omitempty"`

	// All the metadata from the request, echoed back
	ClientRegistrationRequest

	// OPTIONAL. String containing the access token to be used at the client configuration endpoint
	RegistrationAccessToken string `json:"registration_access_token,omitempty"`

	// OPTIONAL. String containing the location of the client configuration endpoint
	RegistrationClientURI string `json:"registration_client_uri,omitempty"`
}

// DCRConfig contains configuration for Dynamic Client Registration
type DCRConfig struct {
	RegistrationEndpoint string
	InitialAccessToken   string
	Metadata             ClientRegistrationRequest
}

// RegisterClient performs dynamic client registration per RFC 7591
func RegisterClient(
	ctx context.Context,
	config DCRConfig,
	hc *http.Client,
) (request Request, response ClientRegistrationResponse, err error) {
	var (
		req       *http.Request
		resp      *http.Response
		reqBody   []byte
		endpoint  string = config.RegistrationEndpoint
	)

	// Validate required endpoint
	if endpoint == "" {
		return request, response, errors.New("registration endpoint is required")
	}

	// Prepare request body
	if reqBody, err = json.Marshal(config.Metadata); err != nil {
		return request, response, errors.Wrap(err, "failed to marshal registration request")
	}

	// Create HTTP request
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(reqBody))); err != nil {
		return request, response, errors.Wrap(err, "failed to create registration request")
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add initial access token if provided
	if config.InitialAccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.InitialAccessToken)
	}

	// Capture request details
	request.Method = req.Method
	request.URL = req.URL
	request.Headers = req.Header

	// Send request
	if resp, err = hc.Do(req); err != nil {
		return request, response, errors.Wrap(err, "failed to send registration request")
	}

	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		return request, response, ParseError(resp)
	}

	// Parse response
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return request, response, errors.Wrap(err, "failed to parse registration response")
	}

	return request, response, nil
}

// ClientFromRegistration converts a registration response to ClientConfig
func ClientFromRegistration(response ClientRegistrationResponse) ClientConfig {
	var config ClientConfig

	config.ClientID = response.ClientID
	config.ClientSecret = response.ClientSecret

	if len(response.RedirectURIs) > 0 {
		config.RedirectURL = response.RedirectURIs[0]
	}

	if len(response.GrantTypes) > 0 {
		config.GrantType = response.GrantTypes[0]
	}

	if response.TokenEndpointAuthMethod != "" {
		config.AuthMethod = response.TokenEndpointAuthMethod
	}

	if response.Scope != "" {
		config.Scopes = strings.Split(response.Scope, " ")
	}

	return config
}

// DefaultRegistrationMetadata returns default metadata for client registration
func DefaultRegistrationMetadata() ClientRegistrationRequest {
	return ClientRegistrationRequest{
		RedirectURIs:            []string{"http://localhost:9876/callback"},
		ResponseTypes:           []string{"code"},
		GrantTypes:              []string{AuthorizationCodeGrantType},
		ClientName:              "oauth2c CLI Client",
		TokenEndpointAuthMethod: ClientSecretBasicAuthMethod,
	}
}