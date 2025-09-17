package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cloudentity/oauth2c/internal/oauth2"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

// createHTTPClient creates an HTTP client similar to the main oauth2 command
func createHTTPClient(insecure bool, timeout time.Duration) *http.Client {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
			MinVersion:         tls.VersionTLS12,
		},
	}

	return &http.Client{Timeout: timeout, Transport: tr}
}

type RegisterCmd struct {
	*cobra.Command
}

func NewRegisterCmd() *RegisterCmd {
	var (
		dcrConfig oauth2.DCRConfig
		timeout   time.Duration
		insecure  bool
		output    string
	)

	cmd := &RegisterCmd{
		Command: &cobra.Command{
			Use:   "register [issuer url]",
			Short: "Register a new OAuth2 client using RFC 7591 Dynamic Client Registration",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				var (
					sconfig oauth2.ServerConfig
					err     error
					hc      = createHTTPClient(insecure, timeout)
				)

				ctx := context.Background()

				// If issuer URL is provided, fetch well-known configuration
				if len(args) > 0 {
					issuerURL := args[0]
					if _, sconfig, err = oauth2.FetchOpenIDConfiguration(ctx, issuerURL, hc); err != nil {
						return fmt.Errorf("failed to fetch OpenID configuration: %w", err)
					}
				}

				// Use registration endpoint from server config if available
				if sconfig.RegistrationEndpoint != "" && dcrConfig.RegistrationEndpoint == "" {
					dcrConfig.RegistrationEndpoint = sconfig.RegistrationEndpoint
				}

				// Validate that we have a registration endpoint
				if dcrConfig.RegistrationEndpoint == "" {
					return fmt.Errorf("registration endpoint is required (use --registration-endpoint flag or provide issuer URL with registration_endpoint in .well-known/openid-configuration)")
				}

				// Set defaults if metadata is empty
				if len(dcrConfig.Metadata.RedirectURIs) == 0 &&
					len(dcrConfig.Metadata.GrantTypes) == 0 &&
					dcrConfig.Metadata.ClientName == "" {
					dcrConfig.Metadata = oauth2.DefaultRegistrationMetadata()
				}

				// Register the client
				request, response, err := oauth2.RegisterClient(ctx, dcrConfig, hc)
				if err != nil {
					return fmt.Errorf("client registration failed: %w", err)
				}

				// Output the result
				switch output {
				case "json":
					data, err := json.Marshal(response)
					if err != nil {
						return err
					}
					fmt.Println(string(pretty.Pretty(data)))
				case "config":
					// Output as oauth2c config format
					clientConfig := oauth2.ClientFromRegistration(response)
					config := Config{
						ClientID:                clientConfig.ClientID,
						ClientSecret:            clientConfig.ClientSecret,
						OpenIDDiscoveryEndpoint: args[0] + oauth2.OpenIDConfigurationPath,
					}
					data, err := json.MarshalIndent(config, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
				default:
					// Default human-readable output
					fmt.Printf("Client registration successful!\n\n")
					fmt.Printf("Client ID: %s\n", response.ClientID)
					if response.ClientSecret != "" {
						fmt.Printf("Client Secret: %s\n", response.ClientSecret)
					}
					if response.ClientIDIssuedAt > 0 {
						fmt.Printf("Client ID Issued At: %s\n", time.Unix(response.ClientIDIssuedAt, 0).Format(time.RFC3339))
					}
					if response.ClientSecretExpiresAt > 0 {
						fmt.Printf("Client Secret Expires At: %s\n", time.Unix(response.ClientSecretExpiresAt, 0).Format(time.RFC3339))
					}
					if response.RegistrationAccessToken != "" {
						fmt.Printf("Registration Access Token: %s\n", response.RegistrationAccessToken)
					}
					if response.RegistrationClientURI != "" {
						fmt.Printf("Registration Client URI: %s\n", response.RegistrationClientURI)
					}

					fmt.Printf("\nYou can now use this client with oauth2c:\n")
					fmt.Printf("oauth2c %s --client-id %s", args[0], response.ClientID)
					if response.ClientSecret != "" {
						fmt.Printf(" --client-secret %s", response.ClientSecret)
					}
					fmt.Printf("\n")
				}

				// Print the request details if not silent
				if !silent {
					fmt.Fprintf(os.Stderr, "\nRequest:\n%s %s\n", request.Method, request.URL)
					for k, v := range request.Headers {
						for _, val := range v {
							if k == "Authorization" {
								fmt.Fprintf(os.Stderr, "%s: Bearer [REDACTED]\n", k)
							} else {
								fmt.Fprintf(os.Stderr, "%s: %s\n", k, val)
							}
						}
					}
					fmt.Fprintf(os.Stderr, "\n")
				}

				return nil
			},
		},
	}

	// DCR-specific flags
	cmd.Flags().StringVar(&dcrConfig.RegistrationEndpoint, "registration-endpoint", "", "OAuth2 client registration endpoint URL")
	cmd.Flags().StringVar(&dcrConfig.InitialAccessToken, "initial-access-token", "", "initial access token for client registration")

	// Client metadata flags
	cmd.Flags().StringSliceVar(&dcrConfig.Metadata.RedirectURIs, "redirect-uris", []string{"http://localhost:9876/callback"}, "redirect URIs for the client")
	cmd.Flags().StringSliceVar(&dcrConfig.Metadata.ResponseTypes, "response-types", []string{"code"}, "response types for the client")
	cmd.Flags().StringSliceVar(&dcrConfig.Metadata.GrantTypes, "grant-types", []string{oauth2.AuthorizationCodeGrantType}, "grant types for the client")
	cmd.Flags().StringVar(&dcrConfig.Metadata.ClientName, "client-name", "oauth2c CLI Client", "human-readable client name")
	cmd.Flags().StringVar(&dcrConfig.Metadata.ClientURI, "client-uri", "", "URI of the client")
	cmd.Flags().StringVar(&dcrConfig.Metadata.LogoURI, "logo-uri", "", "URI of the client logo")
	cmd.Flags().StringVar(&dcrConfig.Metadata.Scope, "scope", "", "space-separated scope values")
	cmd.Flags().StringSliceVar(&dcrConfig.Metadata.Contacts, "contacts", []string{}, "contact email addresses")
	cmd.Flags().StringVar(&dcrConfig.Metadata.TosURI, "tos-uri", "", "URI of terms of service")
	cmd.Flags().StringVar(&dcrConfig.Metadata.PolicyURI, "policy-uri", "", "URI of privacy policy")
	cmd.Flags().StringVar(&dcrConfig.Metadata.JwksURI, "jwks-uri", "", "URI of JSON Web Key Set")
	cmd.Flags().StringVar(&dcrConfig.Metadata.SoftwareID, "software-id", "", "software identifier")
	cmd.Flags().StringVar(&dcrConfig.Metadata.SoftwareVersion, "software-version", "", "software version")
	cmd.Flags().StringVar(&dcrConfig.Metadata.TokenEndpointAuthMethod, "token-endpoint-auth-method", oauth2.ClientSecretBasicAuthMethod, "token endpoint authentication method")

	// Global flags
	cmd.Flags().DurationVar(&timeout, "http-timeout", 10*time.Second, "HTTP client timeout")
	cmd.Flags().BoolVar(&insecure, "insecure", false, "allow insecure TLS connections")
	cmd.Flags().StringVarP(&output, "output", "o", "default", "output format (default, json, config)")

	return cmd
}