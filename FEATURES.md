# HTTP-CLI Feature Specification

A comprehensive CLI tool with full Postman-equivalent capabilities.

---

## Table of Contents

1. [Core HTTP Features](#1-core-http-features)
2. [Request Building](#2-request-building--configuration)
3. [Authentication](#3-authentication)
4. [Collections & Organization](#4-collections--organization)
5. [Environment & Variables](#5-environment--variables)
6. [Scripting](#6-pre-request--post-response-scripts)
7. [Testing & Assertions](#7-testing--assertions)
8. [Automation & Collection Runner](#8-automation--collection-runner)
9. [API Documentation](#9-api-documentation)
10. [Mock Servers](#10-mock-servers)
11. [Monitoring](#11-monitoring--scheduled-runs)
12. [Import/Export](#12-importexport--format-conversion)
13. [Proxy & Certificates](#13-proxy--certificate-management)
14. [WebSocket, gRPC & GraphQL](#14-websocket-grpc--graphql)
15. [Cookie Management](#15-cookie-management)
16. [Response Handling](#16-response-handling--visualization)
17. [Performance Testing](#17-performance-testing)
18. [CLI-Specific Features](#18-cli-specific-features)
19. [Collaboration & Sharing](#19-collaboration--sharing)

---

## 1. Core HTTP Features

### HTTP Methods
- GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE, CONNECT
- Custom HTTP methods support

### Protocol Versions
- HTTP/1.1, HTTP/2, HTTP/3

### Request Configuration
- Connection timeout and response timeout settings
- Rate limiting and throttling control
- Configurable redirect behavior (auto, manual, never, max hops)
- Keep-alive connection management
- Custom Host header support

### Response Handling
- Status code capture and display
- Response headers parsing
- Response body in multiple formats (JSON, XML, HTML, raw text, binary)
- Response size and timing information
- Compression handling (gzip, deflate, brotli)
- Streaming for large payloads
- Download response bodies to files

---

## 2. Request Building & Configuration

### URL & Query Parameters
- Query string builder with key-value pairs
- Automatic URL encoding/decoding
- URL validation
- URL templating with `{{variable}}` syntax

### Path Variables
- Dynamic path parameter substitution
- Optional path parameters

### Request Headers
- Custom key-value header creation
- Preset common headers (User-Agent, Accept, Content-Type, etc.)
- Header inheritance from collection/folder level
- Enable/disable individual headers without deletion
- Auto-generated headers (Content-Length, Host)

### Body Types
| Type | Description |
|------|-------------|
| **JSON** | Direct input, pretty-print, minify, schema validation |
| **Form Data** | `multipart/form-data` with key-value pairs and file uploads |
| **URL-Encoded** | `application/x-www-form-urlencoded` with automatic encoding |
| **Raw** | Plain text, JSON, XML, HTML, custom content types |
| **Binary** | File upload for images, PDFs, etc. |
| **GraphQL** | Query + variables support with validation |
| **XML/SOAP** | XML input with validation, SOAP envelope support |

---

## 3. Authentication

| Auth Type | Details |
|-----------|---------|
| **Basic Auth** | Username/password with Base64 encoding |
| **Bearer Token** | Token with customizable prefix |
| **API Key** | Key-value in header or query parameter |
| **OAuth 2.0** | Authorization Code, Client Credentials, Password, Refresh Token flows. Automatic token refresh, caching, expiration tracking |
| **OAuth 1.0** | HMAC-SHA1, RSA-SHA1, PLAINTEXT signature methods |
| **Digest Auth** | Challenge-response with QoP support |
| **AWS Signature** | AWS4-HMAC-SHA256 with region/service config |
| **NTLM** | Windows challenge/response with domain support |
| **Hawk** | ID/Key with SHA256/SHA1 algorithm selection |
| **Akamai EdgeGrid** | Client token, client secret, access token signing |
| **Custom Auth** | Manual header construction from variables |

- Auth inheritance from collection/folder level
- Auth can be set per-request, per-folder, or per-collection

---

## 4. Collections & Organization

### Collection Structure
- Create, edit, delete, clone collections
- Collection metadata (name, description, version, author)
- Collection-level auth, variables, and scripts

### Folder Hierarchy
- Nested folders within collections
- Folder-level settings inheritance (auth, variables, scripts)
- Folder descriptions and metadata

### Workspaces
- Create and switch between workspaces
- Personal vs team workspaces
- Workspace-scoped settings

### Search & Navigation
- Search collections by name, description, tag
- Filter requests by method, URL, tag
- Request history within collections

---

## 5. Environment & Variables

### Variable Scopes (Precedence: highest to lowest)
1. **Local/Request Variables** - Available only during a single request
2. **Data Variables** - From iteration data files
3. **Environment Variables** - Scoped to active environment
4. **Collection Variables** - Scoped to a collection
5. **Global Variables** - Available across all collections

### Variable Types
- String, Boolean, Numeric, JSON objects, Arrays
- Initial value vs current value separation
- Secret variables (masked in output)

### Dynamic Variables
| Variable | Output |
|----------|--------|
| `{{$timestamp}}` | Current Unix timestamp |
| `{{$isoTimestamp}}` | ISO 8601 timestamp |
| `{{$randomInt}}` | Random integer |
| `{{$randomUUID}}` | UUID v4 |
| `{{$randomEmail}}` | Random email address |
| `{{$randomIPv4}}` | Random IPv4 address |
| `{{$randomPassword}}` | Random password string |
| `{{$guid}}` | Random GUID |

### Environment Management
- Create, edit, delete, switch environments
- Environment file import/export (JSON, .env)
- Multi-environment support (dev, staging, production)
- Variable substitution in URLs, headers, body, auth fields

---

## 6. Pre-request & Post-response Scripts

### Script Execution Order
1. Collection-level pre-request script
2. Folder-level pre-request script
3. Request-level pre-request script
4. **--- Request is sent ---**
5. Request-level post-response script
6. Folder-level post-response script
7. Collection-level post-response script

### Scripting Language
- JavaScript (ES5+ via embedded engine)
- Access to `pm` API object for Postman-compatible scripting

### Pre-request Script Capabilities
- Modify request headers, body, URL, and parameters dynamically
- Generate auth tokens and calculate signatures
- Set/get/unset variables at any scope
- Conditional logic, date/time calculations, string manipulation
- Crypto operations (HMAC, SHA, MD5)

### Post-response Script Capabilities
- Access response body, status code, headers, cookies, response time
- Parse JSON/XML responses and extract data
- Set variables from response data for chaining
- Call other APIs programmatically
- Console logging for debugging

### Test Assertions (BDD-style)
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Body contains user", function () {
    pm.expect(pm.response.json().name).to.eql("John");
});
```

---

## 7. Testing & Assertions

### Response Validation
- Status code checks (exact match, range, class e.g. 2xx)
- Response body assertions (contains, exact match, regex, JSON schema)
- Response header validation (presence, value, format)
- Response time assertions (max, min, average)
- Response size checks

### Body Testing
- JSON schema validation (Draft 4/6/7)
- XML structure validation
- Regex pattern matching
- Content-type validation

### Cookie Testing
- Cookie presence and value assertion
- Domain/path validation
- Expiration checking

### Test Reporting
- Pass/fail summary per request and per collection
- Failed assertion details with expected vs actual
- Test execution timeline
- Exportable test results (JSON, JUnit XML)

---

## 8. Automation & Collection Runner

### Collection Execution
- Run entire collection or specific folders
- Sequential and parallel execution modes
- Configurable iteration count
- Stop on first failure option
- Configurable delay between requests

### Data-Driven Testing
- CSV data file support for iteration variables
- JSON data file support for iteration variables
- Variable mapping from data columns
- Access iteration number and data in scripts

### Run Results
- Summary statistics (total, passed, failed, skipped)
- Individual request results with timing
- Failed assertion details
- Console output capture
- Export results to JSON/JUnit/HTML

### Scheduled Execution
- Cron-based scheduling for collection runs
- One-time scheduled runs
- Logging and monitoring of scheduled results

---

## 9. API Documentation

### Auto-Generation
- Generate documentation from collection structure
- Include request descriptions, parameters, headers
- Include request/response example pairs
- Status code and error documentation

### Export Formats
- Markdown
- HTML (single-page and multi-page)
- OpenAPI 3.0/3.1 specification
- Custom template support

---

## 10. Mock Servers

### Local Mock Server
- Create mock server from collection on localhost
- Configurable port
- Hot reload on collection changes
- Custom response delay injection

### Request Matching
- Path matching (exact and pattern-based)
- HTTP method matching
- Query parameter matching
- Header and body matching (optional)
- Multiple response examples with selection logic
- Default response for unmatched requests

### Mock Configuration
- Response delays per endpoint
- Status code overrides
- Custom response headers
- Error simulation (5xx, timeouts, slow responses)

---

## 11. Monitoring & Scheduled Runs

### Monitor Configuration
- Create monitor from collection
- Schedule: minutely, hourly, daily, weekly, custom cron
- Alert configuration (email, webhook, Slack)

### Metrics Tracked
- Success/failure rate over time
- Response time history and trends
- Uptime/downtime tracking
- P50/P90/P95/P99 response time percentiles

### Health Check Features
- API availability validation
- Performance baseline and anomaly detection
- SLA reporting

---

## 12. Import/Export & Format Conversion

### Import Formats
| Format | Support |
|--------|---------|
| Postman Collection JSON (v2.1) | Full |
| OpenAPI 3.0 / 3.1 | Full |
| Swagger 2.0 | Full |
| cURL commands | Full |
| HAR (HTTP Archive) | Full |
| RAML 0.8 / 1.0 | Basic |
| WSDL 1.1 / 2.0 | Basic |
| Insomnia exports | Full |
| GraphQL schemas | Basic |

### Export Formats
- Postman Collection JSON (v2.1)
- OpenAPI 3.0/3.1 specification
- cURL commands (per request)
- HAR format
- Raw HTTP dump

### Conversion
- Bidirectional format conversion between supported formats
- Collection format migration
- Spec validation during import

---

## 13. Proxy & Certificate Management

### Proxy Configuration
- HTTP/HTTPS proxy (host, port, auth)
- SOCKS4/SOCKS5 proxy support
- Proxy bypass rules (no-proxy hosts)
- Per-domain proxy configuration
- System proxy detection

### Client Certificates
- PEM and PKCS#12 (PFX) format support
- Certificate per domain/host mapping
- Private key passphrase protection
- Certificate chain validation

### CA Certificates
- Custom CA certificate import
- Self-signed certificate acceptance
- SSL/TLS verification toggle (per-request and global)
- SSL/TLS version control
- Cipher suite configuration

---

## 14. WebSocket, gRPC & GraphQL

### WebSocket
- Connection establishment with custom headers
- Send/receive text and binary messages
- Message history/timeline display
- Ping/pong support
- Auto-reconnect on disconnect
- Connection status monitoring

### gRPC
- Service discovery from .proto files
- Unary RPC calls
- Server-side, client-side, and bidirectional streaming
- Protobuf message encoding/decoding
- gRPC metadata (headers/trailers)
- gRPC status code handling

### GraphQL
- Query, mutation, and subscription support
- Query variables input
- Schema introspection
- Query validation against schema
- Custom headers per GraphQL request

---

## 15. Cookie Management

### Cookie Jar
- Automatic cookie persistence across requests
- Domain and path scoping
- Expiration handling
- Secure, HttpOnly, SameSite attribute respect

### Cookie Operations
- Manually add, edit, delete cookies
- Clear all cookies or per-domain
- Export/import cookie jars
- Transfer cookies between domains
- Cookie-to-variable conversion

### Behavior
- Auto-send matching cookies with requests
- Auto-update cookies from `Set-Cookie` response headers
- Cookie inheritance in collection runs

---

## 16. Response Handling & Visualization

### Display Modes
- Pretty-printed JSON with syntax highlighting
- Formatted XML display
- HTML content preview
- Raw text output
- Binary file download/save

### Response Analysis
- Response size, timing, and encoding info
- Headers display
- Status code with reason phrase
- Content-type detection and auto-formatting

### Response History
- Per-request response history
- Side-by-side response comparison
- Response data extraction log

### Data Export
- Copy response to clipboard
- Save response to file (with format options)
- Generate code snippets from request/response (cURL, Python, JS, Go, etc.)

---

## 17. Performance Testing

### Load Configuration
- Virtual user (VU) count
- Test duration
- Ramp-up and cool-down periods

### Load Profiles
| Profile | Description |
|---------|-------------|
| **Fixed** | Constant VU count for duration |
| **Ramp-up** | Gradual increase from 0 to target VUs |
| **Spike** | Sudden increase then decrease |
| **Peak** | Increase, maintain, decrease |
| **Custom** | User-defined load curves |

### Metrics
- Average, min, max response times
- P50, P90, P95, P99 percentiles
- Throughput (requests/second)
- Error rate and failed request count
- Per-request breakdown

### Reporting
- Real-time metrics during test (terminal progress)
- Summary report at completion
- Export to JSON/HTML/CSV

---

## 18. CLI-Specific Features

### Command Interface
```
http <method> <url> [flags]
http run <collection> [flags]
http env <subcommand> [flags]
http mock <subcommand> [flags]
http import <file> [flags]
http export <collection> [flags]
http doc <collection> [flags]
http monitor <subcommand> [flags]
http perf <collection> [flags]
```

### Output Formatting
| Flag | Output |
|------|--------|
| `--json` | JSON structured output |
| `--table` | Tabular display |
| `--minimal` | Status code and body only |
| `--pretty` | Colored, formatted output (default for TTY) |
| `--raw` | Raw response body (default for pipes) |
| `--csv` | CSV output for tabular data |
| `--no-color` | Disable colored output |
| `--template` | Custom Go template output |

### Piping & Shell Integration
- Accept stdin for request body (`echo '{}' | http post /api`)
- Pipe-friendly output (auto-detect TTY vs pipe)
- Meaningful exit codes (0 success, 1 HTTP error, 2 network error, etc.)
- JSON streaming output for batch operations

### Configuration
- Config file: `~/.http-cli/config.yaml` (global) and `.http-cli.yaml` (project)
- Profile support (named configurations)
- Default headers, auth, proxy per profile
- Shell completion scripts (bash, zsh, fish, PowerShell)

### Interactive Mode
- REPL-style interactive prompt (`http -i`)
- Command history with arrow navigation
- Auto-completion for commands, URLs, headers
- Request preview before execution
- Interactive variable and environment selection

### Debugging
| Flag | Purpose |
|------|---------|
| `-v, --verbose` | Show request/response headers |
| `-vv` | Show full request/response including body |
| `--debug` | Internal debug logging |
| `--trace` | Network trace with timing breakdown (DNS, TCP, TLS, etc.) |
| `--dry-run` | Show what would be sent without sending |

### Timing
- Per-request timing (DNS, connect, TLS handshake, first byte, total)
- `--time` flag for timing summary
- Timing breakdown in verbose mode

---

## 19. Collaboration & Sharing

### Git-Friendly Format
- Collections stored as human-readable YAML/JSON files
- One file per request for minimal merge conflicts
- Folder structure mirrors collection hierarchy
- Environment files separate from collection files

### Version Control Integration
- `http diff` - Show changes in collection
- Collection versioning via Git
- Merge conflict resolution for collection files

### Sharing
- Export collection as single file for sharing
- Import shared collections
- Collection forking/cloning

---

## Implementation Priority

### Phase 1 - MVP (Core CLI)
- Basic HTTP methods (GET, POST, PUT, PATCH, DELETE)
- JSON and form-data body support
- Headers and query parameters
- Basic Auth, Bearer Token, API Key authentication
- Pretty-printed colored output
- cURL import/export
- Environment variables and variable substitution
- Config file support
- Verbose/debug output modes
- Shell completions

### Phase 2 - Collections & Scripting
- Collection and folder management
- Collection runner with sequential execution
- Pre-request and post-response JavaScript scripts
- Test assertions with reporting
- Data-driven testing (CSV/JSON)
- All authentication types
- Cookie management
- OpenAPI/Swagger import

### Phase 3 - Advanced Features
- Mock servers
- Monitoring and scheduled runs
- WebSocket support
- GraphQL support
- gRPC support
- Performance/load testing
- API documentation generation
- Interactive REPL mode

### Phase 4 - Ecosystem
- Plugin/extension system
- Full import/export format support
- Advanced proxy and certificate management
- Response history and comparison
- Code generation from requests
