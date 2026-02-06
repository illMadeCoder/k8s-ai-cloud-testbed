## Appendix: API Design & Contracts

*Deep dive into API design principles, protocols, versioning, and contract testing. Covers REST, GraphQL, gRPC, and the patterns that make APIs maintainable and evolvable.*

### F.1 REST API Design Principles

**Goal:** Master RESTful API design beyond basic CRUD

**Learning objectives:**
- Understand REST constraints and their benefits
- Design consistent, intuitive APIs
- Handle complex operations in RESTful ways

**Tasks:**
- [ ] Create `experiments/rest-api-design/`
- [ ] REST fundamentals:
  - [ ] REST constraints (client-server, stateless, cacheable, uniform interface, layered, code-on-demand)
  - [ ] Richardson Maturity Model (levels 0-3)
  - [ ] HATEOAS (Hypermedia as the Engine of Application State)
- [ ] Resource design:
  - [ ] Nouns, not verbs
  - [ ] Resource naming conventions
  - [ ] Hierarchical vs flat URIs
  - [ ] Plural vs singular
- [ ] HTTP methods:
  - [ ] GET (safe, idempotent)
  - [ ] POST (create, not idempotent by default)
  - [ ] PUT (full replace, idempotent)
  - [ ] PATCH (partial update)
  - [ ] DELETE (idempotent)
  - [ ] OPTIONS, HEAD
- [ ] HTTP status codes:
  - [ ] 2xx success patterns
  - [ ] 3xx redirects (when to use)
  - [ ] 4xx client errors (400 vs 422, 401 vs 403)
  - [ ] 5xx server errors
  - [ ] Consistent error response format
- [ ] Query parameters:
  - [ ] Filtering (?status=active)
  - [ ] Sorting (?sort=created_at:desc)
  - [ ] Pagination (offset, cursor-based)
  - [ ] Field selection (?fields=id,name)
- [ ] Request/response design:
  - [ ] JSON conventions
  - [ ] Envelope vs flat responses
  - [ ] Pagination metadata
  - [ ] Error response structure
- [ ] Complex operations:
  - [ ] Bulk operations
  - [ ] Long-running operations (202 Accepted, polling, webhooks)
  - [ ] Actions that don't fit CRUD
  - [ ] Sub-resources vs query parameters
- [ ] HATEOAS in practice:
  - [ ] Link relations
  - [ ] HAL, JSON:API formats
  - [ ] When HATEOAS helps vs overcomplicates
- [ ] Build example REST API with best practices
- [ ] **ADR:** Document REST API conventions

---

### F.2 GraphQL Design Patterns

**Goal:** Understand GraphQL design and when to use it

**Learning objectives:**
- Design effective GraphQL schemas
- Understand GraphQL trade-offs vs REST
- Implement GraphQL performance optimizations

**Tasks:**
- [ ] Create `experiments/graphql-patterns/`
- [ ] GraphQL fundamentals:
  - [ ] Schema definition language (SDL)
  - [ ] Queries, mutations, subscriptions
  - [ ] Type system (scalars, objects, enums, interfaces, unions)
  - [ ] Single endpoint model
- [ ] Schema design:
  - [ ] Thinking in graphs
  - [ ] Connections and edges (Relay pattern)
  - [ ] Nullable vs non-nullable
  - [ ] Input types
- [ ] Resolver patterns:
  - [ ] Field resolvers
  - [ ] Parent/context passing
  - [ ] DataLoader for batching (N+1 problem)
  - [ ] Resolver composition
- [ ] Query complexity:
  - [ ] Depth limiting
  - [ ] Complexity analysis
  - [ ] Query cost calculation
  - [ ] Persisted queries
- [ ] Pagination patterns:
  - [ ] Offset-based
  - [ ] Cursor-based (Relay connection spec)
  - [ ] Total count considerations
- [ ] Error handling:
  - [ ] Errors array in response
  - [ ] Partial success
  - [ ] Error extensions
  - [ ] Union types for expected errors
- [ ] Authentication/authorization:
  - [ ] Context-based auth
  - [ ] Field-level authorization
  - [ ] Directive-based permissions
- [ ] Federation and stitching:
  - [ ] Apollo Federation
  - [ ] Schema stitching
  - [ ] Distributed graphs
- [ ] REST vs GraphQL:
  - [ ] Over-fetching/under-fetching
  - [ ] Caching challenges (no HTTP caching)
  - [ ] Tooling and ecosystem
  - [ ] When each is appropriate
- [ ] Build GraphQL API with best practices
- [ ] **ADR:** Document GraphQL vs REST decision

---

### F.3 gRPC & Protocol Buffers

**Goal:** Master gRPC for high-performance service communication

**Learning objectives:**
- Design Protocol Buffer schemas
- Implement gRPC services effectively
- Understand gRPC streaming patterns

**Tasks:**
- [ ] Create `experiments/grpc-patterns/`
- [ ] Protocol Buffers (protobuf):
  - [ ] Proto3 syntax
  - [ ] Scalar types, messages, enums
  - [ ] Nested messages and imports
  - [ ] Maps and repeated fields
  - [ ] Well-known types (Timestamp, Duration, Any)
- [ ] Schema evolution:
  - [ ] Field numbering rules
  - [ ] Adding/removing fields safely
  - [ ] Reserved fields
  - [ ] Backward/forward compatibility
- [ ] gRPC fundamentals:
  - [ ] Service definition
  - [ ] Code generation
  - [ ] Channel and stub management
  - [ ] Metadata (headers)
- [ ] RPC patterns:
  - [ ] Unary (request-response)
  - [ ] Server streaming
  - [ ] Client streaming
  - [ ] Bidirectional streaming
- [ ] Error handling:
  - [ ] Status codes (OK, CANCELLED, UNKNOWN, etc.)
  - [ ] Error details
  - [ ] Rich error model
- [ ] Deadlines and timeouts:
  - [ ] Deadline propagation
  - [ ] Context cancellation
  - [ ] Timeout best practices
- [ ] Load balancing:
  - [ ] Client-side (pick_first, round_robin)
  - [ ] Proxy-based (Envoy, Linkerd)
  - [ ] Service mesh integration
- [ ] Interceptors:
  - [ ] Unary and stream interceptors
  - [ ] Logging, metrics, auth
  - [ ] Chaining interceptors
- [ ] gRPC-Web:
  - [ ] Browser support
  - [ ] Envoy proxy requirement
  - [ ] Limitations vs native gRPC
- [ ] gRPC vs REST:
  - [ ] Performance (binary, HTTP/2)
  - [ ] Type safety
  - [ ] Tooling
  - [ ] Browser support
- [ ] Build gRPC services with streaming
- [ ] **ADR:** Document gRPC adoption strategy

---

### F.4 API Versioning Strategies

**Goal:** Understand API versioning approaches and evolution

**Learning objectives:**
- Choose appropriate versioning strategy
- Evolve APIs without breaking clients
- Manage deprecation gracefully

**Tasks:**
- [ ] Create `experiments/api-versioning/`
- [ ] Why version:
  - [ ] Breaking vs non-breaking changes
  - [ ] Client compatibility
  - [ ] Independent evolution
- [ ] Versioning strategies:
  - [ ] URI path (/v1/users)
  - [ ] Query parameter (?version=1)
  - [ ] Header (Accept: application/vnd.api+json;version=1)
  - [ ] Content negotiation
- [ ] URI versioning:
  - [ ] Most visible/discoverable
  - [ ] Caching friendly
  - [ ] "Not RESTful" criticism
  - [ ] Most commonly used
- [ ] Header versioning:
  - [ ] Cleaner URIs
  - [ ] Content negotiation alignment
  - [ ] Harder to test/discover
  - [ ] Proxy/cache complications
- [ ] No explicit versioning:
  - [ ] Evolutionary design
  - [ ] MUST be backwards compatible
  - [ ] Tolerant readers
  - [ ] Stripe's approach
- [ ] Breaking changes:
  - [ ] What constitutes breaking
  - [ ] Removing fields
  - [ ] Changing types
  - [ ] Semantic changes
- [ ] Non-breaking evolution:
  - [ ] Adding optional fields
  - [ ] Adding endpoints
  - [ ] Deprecation annotations
- [ ] Deprecation strategy:
  - [ ] Sunset headers
  - [ ] Deprecation timelines
  - [ ] Usage monitoring
  - [ ] Client communication
- [ ] Multi-version support:
  - [ ] Code organization (adapters, transformers)
  - [ ] Testing multiple versions
  - [ ] Documentation per version
  - [ ] Operational overhead
- [ ] GraphQL versioning:
  - [ ] Schema evolution without versioning
  - [ ] Deprecation directive
  - [ ] Field removal lifecycle
- [ ] gRPC versioning:
  - [ ] Package versioning
  - [ ] Protobuf evolution rules
  - [ ] Service versioning
- [ ] **ADR:** Document API versioning strategy

---

### F.5 OpenAPI & API Documentation

**Goal:** Create comprehensive, accurate API documentation

**Learning objectives:**
- Write effective OpenAPI specifications
- Generate documentation from specs
- Keep documentation in sync with implementation

**Tasks:**
- [ ] Create `experiments/api-documentation/`
- [ ] OpenAPI fundamentals:
  - [ ] OpenAPI 3.0 vs 3.1 vs 2.0 (Swagger)
  - [ ] Specification structure
  - [ ] YAML vs JSON format
- [ ] Writing specifications:
  - [ ] Info and servers
  - [ ] Paths and operations
  - [ ] Parameters (path, query, header)
  - [ ] Request bodies
  - [ ] Responses and status codes
- [ ] Schema definitions:
  - [ ] $ref for reusability
  - [ ] Components section
  - [ ] Inheritance (allOf, oneOf, anyOf)
  - [ ] Examples
- [ ] Security schemes:
  - [ ] API key
  - [ ] OAuth2 flows
  - [ ] OpenID Connect
  - [ ] Security requirements
- [ ] Code generation:
  - [ ] Server stubs
  - [ ] Client SDKs
  - [ ] Validation middleware
  - [ ] OpenAPI Generator, oapi-codegen
- [ ] Documentation tools:
  - [ ] Swagger UI
  - [ ] Redoc
  - [ ] Stoplight
  - [ ] Hosting options
- [ ] Design-first vs code-first:
  - [ ] Design-first workflow
  - [ ] Code-first with annotations
  - [ ] Keeping spec and code in sync
  - [ ] Linting and validation
- [ ] API style guides:
  - [ ] Spectral for linting
  - [ ] Custom rules
  - [ ] Organizational standards
- [ ] AsyncAPI for events:
  - [ ] Event-driven API documentation
  - [ ] Message formats
  - [ ] Channel definitions
- [ ] Set up OpenAPI workflow
- [ ] **ADR:** Document API documentation approach

---

### F.6 Contract Testing

**Goal:** Verify API contracts between services

**Learning objectives:**
- Understand contract testing vs integration testing
- Implement consumer-driven contract testing
- Integrate contract testing in CI/CD

**Tasks:**
- [ ] Create `experiments/contract-testing/`
- [ ] Contract testing fundamentals:
  - [ ] What is a contract
  - [ ] Provider vs consumer
  - [ ] Why contracts matter (microservices)
- [ ] Contract testing vs alternatives:
  - [ ] vs Integration tests (speed, isolation)
  - [ ] vs E2E tests (scope, flakiness)
  - [ ] vs API mocking (drift prevention)
- [ ] Consumer-driven contracts (CDC):
  - [ ] Consumer defines expectations
  - [ ] Provider verifies against contracts
  - [ ] Pact framework
- [ ] Pact workflow:
  - [ ] Consumer tests generate contracts
  - [ ] Contracts shared via Pact Broker
  - [ ] Provider verifies contracts
  - [ ] Can-I-Deploy checks
- [ ] Pact implementation:
  - [ ] Consumer test setup
  - [ ] Defining interactions
  - [ ] Matchers (flexible matching)
  - [ ] Provider states
- [ ] Provider verification:
  - [ ] Running against real provider
  - [ ] Provider states setup
  - [ ] Verification in CI
- [ ] Pact Broker:
  - [ ] Contract storage
  - [ ] Version management
  - [ ] Deployment safety checks
  - [ ] Webhooks for CI triggers
- [ ] Schema-based contracts:
  - [ ] OpenAPI as contract
  - [ ] Prism for mocking
  - [ ] Dredd for validation
  - [ ] Schemathesis for fuzzing
- [ ] gRPC contract testing:
  - [ ] Protobuf as contract
  - [ ] Breaking change detection
  - [ ] buf tool for linting
- [ ] GraphQL contract testing:
  - [ ] Schema as contract
  - [ ] Breaking change detection
  - [ ] Apollo Studio
- [ ] Set up Pact-based contract testing
- [ ] **ADR:** Document contract testing strategy

---

### F.7 API Gateway Patterns

**Goal:** Understand API gateway architectures and patterns

**Learning objectives:**
- Understand API gateway responsibilities
- Choose between gateway patterns
- Configure common gateway features

**Tasks:**
- [ ] Create `experiments/api-gateway/`
- [ ] API gateway fundamentals:
  - [ ] Edge service pattern
  - [ ] Responsibilities (routing, auth, rate limiting)
  - [ ] Gateway vs service mesh
- [ ] Gateway patterns:
  - [ ] Single gateway
  - [ ] Backend for Frontend (BFF)
  - [ ] Gateway per team/domain
  - [ ] Sidecar pattern
- [ ] Routing:
  - [ ] Path-based routing
  - [ ] Header-based routing
  - [ ] Host-based routing
  - [ ] Request transformation
- [ ] Authentication at gateway:
  - [ ] Token validation
  - [ ] OAuth2/OIDC integration
  - [ ] API key validation
  - [ ] mTLS termination
- [ ] Rate limiting:
  - [ ] Per client/API key
  - [ ] Per endpoint
  - [ ] Sliding window vs fixed window
  - [ ] Distributed rate limiting
- [ ] Request/response transformation:
  - [ ] Header manipulation
  - [ ] Body transformation
  - [ ] Protocol translation
- [ ] Caching:
  - [ ] Response caching
  - [ ] Cache invalidation
  - [ ] Conditional requests (ETags)
- [ ] Observability:
  - [ ] Access logging
  - [ ] Metrics collection
  - [ ] Distributed tracing
  - [ ] Error aggregation
- [ ] Gateway options:
  - [ ] Kong
  - [ ] Ambassador/Emissary
  - [ ] AWS API Gateway
  - [ ] Azure API Management
  - [ ] Envoy (as gateway)
- [ ] Gateway deployment:
  - [ ] Kubernetes ingress integration
  - [ ] High availability
  - [ ] Blue-green gateway updates
- [ ] Deploy and configure Kong
- [ ] **ADR:** Document API gateway selection

---

### F.8 API Security Patterns

**Goal:** Secure APIs against common attacks

**Learning objectives:**
- Implement API authentication patterns
- Protect against common API attacks
- Design defense in depth for APIs

**Tasks:**
- [ ] Create `experiments/api-security/`
- [ ] OWASP API Security Top 10:
  - [ ] Broken Object Level Authorization (BOLA)
  - [ ] Broken Authentication
  - [ ] Broken Object Property Level Authorization
  - [ ] Unrestricted Resource Consumption
  - [ ] Broken Function Level Authorization
  - [ ] Mass Assignment
  - [ ] Security Misconfiguration
  - [ ] Lack of Protection from Automated Threats
  - [ ] Improper Asset Management
  - [ ] Server-Side Request Forgery (SSRF)
- [ ] Authentication patterns:
  - [ ] Token-based (JWT, opaque)
  - [ ] API keys (when appropriate)
  - [ ] mTLS for service-to-service
  - [ ] OAuth2 scopes
- [ ] Authorization patterns:
  - [ ] Resource-level authorization
  - [ ] Field-level authorization
  - [ ] Ownership validation
  - [ ] Role-based vs attribute-based
- [ ] Input validation:
  - [ ] Schema validation
  - [ ] Type coercion attacks
  - [ ] Injection prevention
  - [ ] File upload security
- [ ] Rate limiting security:
  - [ ] Brute force prevention
  - [ ] Credential stuffing protection
  - [ ] Cost-based limiting
- [ ] API abuse prevention:
  - [ ] Bot detection
  - [ ] Behavioral analysis
  - [ ] Device fingerprinting
  - [ ] CAPTCHAs (sparingly)
- [ ] Response security:
  - [ ] Sensitive data exposure
  - [ ] Error message information leakage
  - [ ] Security headers
- [ ] API inventory:
  - [ ] Discovering shadow APIs
  - [ ] Deprecated endpoint management
  - [ ] Documentation alignment
- [ ] Security testing:
  - [ ] DAST for APIs
  - [ ] Fuzzing
  - [ ] Penetration testing
- [ ] **ADR:** Document API security controls

---
