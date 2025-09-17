# OAuth2c: User-friendly OAuth CLI

OAuth2c is a Go command-line tool for interacting with OAuth 2.0 authorization servers. It supports multiple grant flows, client authentication methods, and advanced OAuth2/OIDC features.

**ALWAYS reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Working Effectively

### Setup and Build
- Ensure Go 1.24+ is installed: `go version`
- Download dependencies: `go mod download` -- takes ~0-8 seconds (instant if cached)
- Build the project: `go build ./...` -- takes ~15-22 seconds on first build, <1 second cached
- Alternative fast build: `make build` -- takes <1 second (uses go build ./...)
- Install binary: `make install` -- installs to ~/go/bin/oauth2c, takes <1 second
- **NEVER CANCEL** build commands - set timeout to 60+ seconds minimum

### Testing
- **CRITICAL**: Most tests require network access and will FAIL in sandboxed environments
- Run unit tests (network-free): `go test ./internal/oauth2` -- takes ~3 seconds, ALWAYS PASSES
- Full test suite: `go test ./...` -- takes ~7 seconds but FAILS due to network restrictions
- The cmd package tests attempt to connect to `oauth2c.us.authz.cloudentity.io` and will fail with TLS/network errors in restricted environments
- **VALIDATION REQUIREMENT**: Always run `go test ./internal/oauth2` to validate core OAuth2 logic changes

### Linting
- **Docker lint (may fail)**: `make lint` -- uses golangci-lint:v1.64.5 in Docker
- If Docker fails, document the limitation - this is expected in some environments
- **NEVER CANCEL** lint commands - they may take 60+ seconds to pull Docker images

## Validation

### Manual Testing Scenarios
- **Binary functionality**: Test `~/go/bin/oauth2c version` and `~/go/bin/oauth2c --help`
- **Core CLI verification**: Ensure help text shows all major flags (grant-type, client-id, etc.)
- **Build verification**: After changes, run `make build && make install` to ensure no compilation errors
- **Unit test verification**: Run `go test ./internal/oauth2 -v` to validate core OAuth2 logic

### Network-Dependent Testing
- **LIMITATION**: Full integration tests require external OAuth2 servers
- Tests in `cmd/oauth2_test.go` and `cmd/oauth2_token_test.go` need network access
- In restricted environments, focus on unit tests in `internal/oauth2/`

## Repository Structure

### Key Directories
```
/home/runner/work/oauth2c/oauth2c/
├── cmd/                    # CLI command implementations
│   ├── oauth2.go          # Main OAuth2 command and flow orchestration  
│   ├── oauth2_authorize_code.go # Authorization code flow
│   ├── oauth2_device.go   # Device authorization flow
│   ├── oauth2_token.go    # Token endpoint flows
│   └── *_test.go         # Command tests (require network)
├── internal/oauth2/       # Core OAuth2 logic (network-independent)
│   ├── oauth2.go         # Main OAuth2 implementation
│   ├── jwt.go           # JWT handling
│   ├── crypto.go        # Cryptographic operations
│   └── *_test.go        # Unit tests (work without network)
├── main.go              # Application entry point
├── Makefile            # Build automation
├── go.mod              # Go module definition
└── .github/workflows/   # CI/CD pipelines
```

### Important Files to Check After Changes
- **Always check `internal/oauth2/oauth2.go`** when modifying core OAuth2 flows
- **Always check `cmd/oauth2.go`** when modifying CLI interfaces or command structure
- **Always run unit tests** after modifying anything in `internal/oauth2/`

## Common Development Tasks

### Adding New OAuth2 Features
1. Implement core logic in `internal/oauth2/`
2. Add unit tests in `internal/oauth2/*_test.go`
3. Add CLI command handling in appropriate `cmd/oauth2_*.go` file
4. Test with: `make build && make install && ~/go/bin/oauth2c --help`
5. Validate with: `go test ./internal/oauth2`

### Build Commands Reference
```bash
# Download dependencies (0-8 seconds, instant if cached)
go mod download

# Full build (15-22 seconds first time, <1 second cached)
go build ./...

# Quick build via make (same as go build)
make build

# Install binary (1 second)
make install

# Unit tests only (3 seconds, always works)
go test ./internal/oauth2

# All tests (7 seconds, fails without network)
go test ./...

# Lint (may fail in Docker environments)
make lint
```

### Time Expectations
- **go mod download**: ~0-8 seconds (instant if cached)
- **go build ./...** (first time): ~15-22 seconds  
- **make build** (uses go build): <1-15 seconds depending on cache
- **make install**: <1 second
- **go test ./internal/oauth2**: ~3 seconds
- **go test ./...**: ~7 seconds (but fails due to network)

## Troubleshooting

### Network Test Failures
- **Expected**: cmd package tests fail with TLS/network connectivity issues 
- **Solution**: Focus on `internal/oauth2` tests which don't require network
- **Never skip validation** due to network issues - use unit tests instead

### Docker Lint Failures  
- **Expected**: `make lint` may fail with Docker runtime errors
- **Solution**: Document the limitation, don't attempt workarounds
- **Alternative**: Install golangci-lint locally if available

### Build Issues
- **Go version**: Requires Go 1.24+, check with `go version`
- **Module issues**: Run `go mod tidy` if dependency problems occur
- **Cache issues**: Use `go clean -cache` if builds behave unexpectedly

## Development Workflow

1. **Always build first**: `make build` to check compilation
2. **Always test core logic**: `go test ./internal/oauth2` 
3. **Always install and test binary**: `make install && ~/go/bin/oauth2c --help`
4. **Document network limitations** if adding network-dependent features
5. **Set appropriate timeouts** for all commands (60+ seconds minimum)