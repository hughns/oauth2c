package cmd

import (
	"github.com/cloudentity/oauth2c/internal/oauth2"
)

// PromptForDCRConfig prompts the user for DCR configuration if needed
func PromptForDCRConfig(config oauth2.DCRConfig, serverConfig oauth2.ServerConfig) oauth2.DCRConfig {
	// Use registration endpoint from server config if available
	if serverConfig.RegistrationEndpoint != "" && config.RegistrationEndpoint == "" {
		config.RegistrationEndpoint = serverConfig.RegistrationEndpoint
	}

	// Prompt for registration endpoint if not provided
	if config.RegistrationEndpoint == "" {
		config.RegistrationEndpoint = PromptString("Registration endpoint URL")
	}

	// Prompt for initial access token if needed
	if config.InitialAccessToken == "" {
		config.InitialAccessToken = PromptString("Initial access token (leave empty if not required)")
	}

	// Prompt for client metadata if not already set
	if config.Metadata.ClientName == "" {
		config.Metadata.ClientName = PromptString("Client name")
	}

	if len(config.Metadata.RedirectURIs) == 0 {
		redirectURI := PromptString("Redirect URI")
		if redirectURI != "" {
			config.Metadata.RedirectURIs = []string{redirectURI}
		}
	}

	if len(config.Metadata.GrantTypes) == 0 {
		grantTypes := []string{
			oauth2.AuthorizationCodeGrantType,
			oauth2.ClientCredentialsGrantType,
			oauth2.ImplicitGrantType,
			oauth2.PasswordGrantType,
			oauth2.RefreshTokenGrantType,
		}
		selectedGrantType := PromptStringSlice("Grant type", grantTypes)
		if selectedGrantType != "" {
			config.Metadata.GrantTypes = []string{selectedGrantType}
		}
	}

	if config.Metadata.TokenEndpointAuthMethod == "" {
		authMethods := []string{
			oauth2.ClientSecretBasicAuthMethod,
			oauth2.ClientSecretPostAuthMethod,
			oauth2.ClientSecretJwtAuthMethod,
			oauth2.PrivateKeyJwtAuthMethod,
			oauth2.NoneAuthMethod,
		}
		config.Metadata.TokenEndpointAuthMethod = PromptStringSlice("Token endpoint auth method", authMethods)
	}

	// Prompt for scopes
	if config.Metadata.Scope == "" {
		scope := PromptString("Scopes (space separated, leave empty for default)")
		if scope != "" {
			config.Metadata.Scope = scope
		}
	}

	return config
}