## Appendix: Identity & Authentication

*Deep dive into application-level identity, authentication, and secrets management. Covers the full stack from password hashing to OAuth flows to IdP deployment. This is distinct from Phase 3's infrastructure-level security (RBAC, admission control).*

### B.1 Password Management & Credential Storage

**Goal:** Understand secure password handling from hashing algorithms to storage patterns

**Learning objectives:**
- Understand password hashing algorithms and their trade-offs
- Implement secure credential storage patterns
- Design password policies that balance security and usability

**Tasks:**
- [ ] Create `experiments/scenarios/password-security/`
- [ ] Password hashing algorithms:
  - [ ] bcrypt - work factor tuning, cost parameter
  - [ ] Argon2 (Argon2id) - memory-hard, recommended for new systems
  - [ ] scrypt - memory-hard alternative
  - [ ] PBKDF2 - NIST-approved, FIPS compliance
  - [ ] Why MD5/SHA1/SHA256 alone are insufficient (rainbow tables, speed)
- [ ] Salting and peppering:
  - [ ] Per-password random salts
  - [ ] Application-level pepper (secret key)
  - [ ] Salt storage patterns
- [ ] Password policies:
  - [ ] Length vs complexity requirements (NIST 800-63B guidance)
  - [ ] Breached password checking (HaveIBeenPwned API, k-anonymity)
  - [ ] Rate limiting and lockout strategies
  - [ ] Password history and rotation (when required vs harmful)
- [ ] Credential storage patterns:
  - [ ] Database schema design for credentials
  - [ ] Encryption at rest for credential tables
  - [ ] Key rotation for encrypted credentials
- [ ] Password reset flows:
  - [ ] Secure token generation (CSPRNG)
  - [ ] Token expiration and single-use
  - [ ] Account enumeration prevention
- [ ] Build demo application demonstrating patterns
- [ ] **ADR:** Document password hashing algorithm selection

---

### B.2 JWT Fundamentals & Internals

**Goal:** Master JSON Web Tokens - structure, signing, validation, and security considerations

**Learning objectives:**
- Understand JWT structure and standard claims
- Implement signing and verification with different algorithms
- Recognize and prevent common JWT vulnerabilities

**Tasks:**
- [ ] Create `experiments/scenarios/jwt-deep-dive/`
- [ ] JWT structure (RFC 7519):
  - [ ] Header (alg, typ, kid)
  - [ ] Payload (claims)
  - [ ] Signature
  - [ ] Base64URL encoding
- [ ] Standard claims:
  - [ ] iss (issuer), sub (subject), aud (audience)
  - [ ] exp (expiration), nbf (not before), iat (issued at)
  - [ ] jti (JWT ID) for revocation
- [ ] Custom claims:
  - [ ] Claim naming conventions
  - [ ] Claim size considerations (header bloat)
  - [ ] Sensitive data in claims (don't)
- [ ] Signing algorithms:
  - [ ] HMAC (HS256, HS384, HS512) - symmetric
  - [ ] RSA (RS256, RS384, RS512) - asymmetric
  - [ ] ECDSA (ES256, ES384, ES512) - asymmetric, smaller keys
  - [ ] EdDSA (Ed25519) - modern, fast
  - [ ] PS256 (RSA-PSS) - probabilistic signatures
  - [ ] Algorithm selection criteria
- [ ] Key management:
  - [ ] JWKS (JSON Web Key Sets)
  - [ ] Key rotation strategies
  - [ ] kid (Key ID) usage
  - [ ] Public key distribution
- [ ] JWT vulnerabilities and mitigations:
  - [ ] Algorithm confusion (alg:none, RS256→HS256)
  - [ ] Key injection attacks
  - [ ] Token replay
  - [ ] Insufficient signature validation
  - [ ] Sensitive data exposure
- [ ] Token lifecycle:
  - [ ] Access tokens (short-lived, ~15 min)
  - [ ] Refresh tokens (longer-lived, secure storage)
  - [ ] Token refresh flow
  - [ ] Sliding vs absolute expiration
- [ ] Revocation strategies:
  - [ ] Short expiration (stateless)
  - [ ] Token blacklist/blocklist
  - [ ] Token versioning (jti + database)
  - [ ] Refresh token rotation
- [ ] Build JWT library wrapper with security defaults
- [ ] **ADR:** Document JWT vs opaque tokens decision

---

### B.3 JWE - JSON Web Encryption

**Goal:** Understand encrypted tokens for sensitive payloads

**Learning objectives:**
- Understand JWE structure and encryption modes
- Implement encrypted tokens for sensitive data
- Know when JWE is appropriate vs JWS

**Tasks:**
- [ ] Create `experiments/scenarios/jwe-encryption/`
- [ ] JWE structure (RFC 7516):
  - [ ] Protected header
  - [ ] Encrypted key
  - [ ] Initialization vector
  - [ ] Ciphertext
  - [ ] Authentication tag
- [ ] Key management algorithms:
  - [ ] RSA-OAEP, RSA-OAEP-256
  - [ ] A128KW, A256KW (AES Key Wrap)
  - [ ] ECDH-ES (Elliptic Curve Diffie-Hellman)
  - [ ] dir (direct encryption)
- [ ] Content encryption algorithms:
  - [ ] A128GCM, A256GCM (AES-GCM)
  - [ ] A128CBC-HS256 (AES-CBC with HMAC)
- [ ] Nested JWT (signed then encrypted):
  - [ ] JWS inside JWE pattern
  - [ ] When to use nested tokens
- [ ] Use cases:
  - [ ] Sensitive claims that must be hidden
  - [ ] Tokens passing through untrusted intermediaries
  - [ ] Regulatory requirements for encryption
- [ ] Performance considerations:
  - [ ] Encryption overhead
  - [ ] Key size impact
  - [ ] When JWS is sufficient
- [ ] Implement JWE token service
- [ ] **ADR:** Document when to use JWE vs JWS

---

### B.4 OAuth 2.0 Flows

**Goal:** Master all OAuth 2.0 grant types and their appropriate use cases

**Learning objectives:**
- Understand each OAuth flow and security properties
- Implement flows correctly with security best practices
- Choose the right flow for each application type

**Tasks:**
- [ ] Create `experiments/scenarios/oauth-flows/`
- [ ] OAuth 2.0 fundamentals (RFC 6749):
  - [ ] Roles: Resource Owner, Client, Authorization Server, Resource Server
  - [ ] Client types: Confidential vs Public
  - [ ] Client registration and credentials
- [ ] Authorization Code Grant:
  - [ ] Flow walkthrough (redirect-based)
  - [ ] State parameter (CSRF protection)
  - [ ] Authorization code exchange
  - [ ] When to use: server-side web apps
- [ ] Authorization Code + PKCE (RFC 7636):
  - [ ] Code verifier and code challenge
  - [ ] S256 vs plain transformation
  - [ ] Why PKCE for all clients (OAuth 2.1)
  - [ ] When to use: SPAs, mobile apps, all public clients
- [ ] Client Credentials Grant:
  - [ ] Machine-to-machine authentication
  - [ ] No user context
  - [ ] When to use: service accounts, daemons, microservices
- [ ] Device Authorization Grant (RFC 8628):
  - [ ] Limited input device flow
  - [ ] User code and verification URI
  - [ ] Polling for token
  - [ ] When to use: TVs, CLI tools, IoT devices
- [ ] Refresh Token Grant:
  - [ ] Obtaining refresh tokens
  - [ ] Refresh token rotation
  - [ ] Refresh token binding
- [ ] Deprecated/Legacy flows (understand why deprecated):
  - [ ] Implicit Grant - why removed in OAuth 2.1
  - [ ] Resource Owner Password Credentials - why dangerous
- [ ] Token types:
  - [ ] Bearer tokens
  - [ ] DPoP (Demonstrating Proof of Possession)
  - [ ] mTLS-bound tokens
- [ ] Security considerations:
  - [ ] Redirect URI validation (exact match)
  - [ ] Token leakage via referrer
  - [ ] Open redirector vulnerabilities
  - [ ] CSRF protection
- [ ] Implement each flow with demo applications
- [ ] **ADR:** Document OAuth flow selection criteria

---

### B.5 OpenID Connect (OIDC)

**Goal:** Understand OIDC as the identity layer on OAuth 2.0

**Learning objectives:**
- Understand OIDC's additions to OAuth 2.0
- Implement OIDC authentication correctly
- Configure OIDC discovery and validation

**Tasks:**
- [ ] Create `experiments/scenarios/oidc-fundamentals/`
- [ ] OIDC overview:
  - [ ] Authentication vs Authorization (OIDC vs OAuth)
  - [ ] ID Token introduction
  - [ ] Standard scopes (openid, profile, email, address, phone)
- [ ] ID Token:
  - [ ] Required claims (iss, sub, aud, exp, iat)
  - [ ] Standard claims (name, email, picture, etc.)
  - [ ] Nonce for replay protection
  - [ ] at_hash, c_hash for token binding
- [ ] OIDC flows:
  - [ ] Authorization Code Flow (recommended)
  - [ ] Implicit Flow (ID token only, legacy)
  - [ ] Hybrid Flow (code + id_token)
- [ ] Discovery (RFC 8414):
  - [ ] .well-known/openid-configuration
  - [ ] Metadata fields
  - [ ] JWKS URI for key retrieval
- [ ] UserInfo endpoint:
  - [ ] When to use vs ID token claims
  - [ ] Scope-based claim filtering
- [ ] Token validation:
  - [ ] Signature verification
  - [ ] Issuer validation
  - [ ] Audience validation
  - [ ] Expiration checking
  - [ ] Nonce verification
- [ ] Session management:
  - [ ] Front-channel logout
  - [ ] Back-channel logout
  - [ ] Session state and check_session_iframe
- [ ] OIDC for Kubernetes:
  - [ ] API server OIDC configuration
  - [ ] kubectl credential plugins (kubelogin)
  - [ ] Group claims mapping to RBAC
- [ ] Implement OIDC client library usage
- [ ] **ADR:** Document OIDC provider selection

---

### B.6 Session Management

**Goal:** Implement secure session handling for web applications

**Learning objectives:**
- Understand session security fundamentals
- Implement secure session storage and lifecycle
- Prevent session-based attacks

**Tasks:**
- [ ] Create `experiments/scenarios/session-management/`
- [ ] Session fundamentals:
  - [ ] Stateful vs stateless sessions
  - [ ] Session ID generation (CSPRNG, entropy)
  - [ ] Session ID length (128+ bits)
- [ ] Session storage:
  - [ ] Server-side: Redis, Memcached, database
  - [ ] Client-side: encrypted cookies (limitations)
  - [ ] Hybrid approaches
- [ ] Cookie security:
  - [ ] HttpOnly flag (XSS protection)
  - [ ] Secure flag (HTTPS only)
  - [ ] SameSite attribute (CSRF protection)
  - [ ] Domain and Path scoping
  - [ ] Cookie prefixes (__Host-, __Secure-)
- [ ] Session lifecycle:
  - [ ] Creation on authentication
  - [ ] Regeneration on privilege change
  - [ ] Absolute timeout vs idle timeout
  - [ ] Explicit logout and invalidation
- [ ] Session attacks and mitigations:
  - [ ] Session fixation - regenerate on login
  - [ ] Session hijacking - secure transport, binding
  - [ ] Session prediction - CSPRNG
  - [ ] Cross-site request forgery - SameSite, tokens
- [ ] Token storage for SPAs:
  - [ ] localStorage risks (XSS)
  - [ ] sessionStorage limitations
  - [ ] HttpOnly cookies (BFF pattern)
  - [ ] In-memory with refresh rotation
- [ ] Concurrent session handling:
  - [ ] Allow multiple sessions
  - [ ] Limit concurrent sessions
  - [ ] Device/session management UI
- [ ] Deploy Redis for session storage
- [ ] Build session management demo
- [ ] **ADR:** Document session storage strategy

---

### B.7 Identity Provider Deployment

**Goal:** Deploy and configure identity providers on Kubernetes

**Learning objectives:**
- Deploy Keycloak and Dex on Kubernetes
- Configure realms, clients, and identity federation
- Integrate IdP with platform services

**Tasks:**
- [ ] Create `experiments/scenarios/identity-provider/`
- [ ] Keycloak deployment:
  - [ ] Helm chart installation
  - [ ] PostgreSQL backend
  - [ ] High availability configuration
  - [ ] Resource requirements
- [ ] Keycloak configuration:
  - [ ] Realm creation and configuration
  - [ ] Client registration (confidential, public)
  - [ ] User federation (LDAP, Active Directory)
  - [ ] Identity brokering (social login, SAML)
  - [ ] Custom themes
  - [ ] User self-registration
- [ ] Keycloak advanced features:
  - [ ] Fine-grained authorization (policies, permissions)
  - [ ] Custom authentication flows
  - [ ] Required actions (MFA setup, password change)
  - [ ] Admin REST API
- [ ] Dex deployment:
  - [ ] Helm chart installation
  - [ ] Static vs dynamic clients
  - [ ] Connector configuration
- [ ] Dex connectors:
  - [ ] LDAP connector
  - [ ] OIDC connector (upstream IdP)
  - [ ] GitHub, GitLab connectors
  - [ ] SAML connector
- [ ] Keycloak vs Dex comparison:
  - [ ] Feature set (Keycloak: full IdP, Dex: federation)
  - [ ] Resource footprint
  - [ ] Operational complexity
  - [ ] Use case fit
- [ ] Platform integration:
  - [ ] ArgoCD OIDC configuration
  - [ ] Grafana OIDC configuration
  - [ ] Harbor OIDC configuration
  - [ ] Kubernetes API server OIDC
- [ ] Multi-tenancy:
  - [ ] Keycloak realms per tenant
  - [ ] Dex with multiple connectors
  - [ ] Tenant isolation patterns
- [ ] **ADR:** Document Keycloak vs Dex selection

---

### B.8 Application Integration Patterns

**Goal:** Implement authentication in different application architectures

**Learning objectives:**
- Understand authentication patterns for different app types
- Implement Backend for Frontend (BFF) pattern
- Secure service-to-service communication

**Tasks:**
- [ ] Create `experiments/scenarios/auth-integration/`
- [ ] Traditional web application:
  - [ ] Server-side session management
  - [ ] OIDC integration with session binding
  - [ ] CSRF protection
- [ ] Single Page Application (SPA):
  - [ ] PKCE flow implementation
  - [ ] Token storage options and trade-offs
  - [ ] Silent refresh / token renewal
  - [ ] Logout handling
- [ ] Backend for Frontend (BFF) pattern:
  - [ ] BFF as OAuth client (confidential)
  - [ ] Session cookies to browser
  - [ ] Token exchange for downstream APIs
  - [ ] Security benefits over direct SPA auth
- [ ] API Gateway authentication:
  - [ ] Token validation at gateway
  - [ ] Token introspection vs local validation
  - [ ] Rate limiting per identity
  - [ ] Gateway options: Kong, Ambassador, Envoy
- [ ] Service-to-service authentication:
  - [ ] mTLS (mutual TLS)
  - [ ] JWT with client credentials
  - [ ] Service mesh identity (SPIFFE/SPIRE)
  - [ ] Kubernetes service account tokens
- [ ] Mobile application:
  - [ ] PKCE flow (required)
  - [ ] Secure token storage (Keychain/Keystore)
  - [ ] Biometric authentication integration
  - [ ] App attestation
- [ ] CLI tool authentication:
  - [ ] Device authorization flow
  - [ ] Local token caching
  - [ ] Credential helpers
- [ ] Microservices patterns:
  - [ ] Token propagation
  - [ ] Token exchange (RFC 8693)
  - [ ] Audience restriction
  - [ ] Scope-based authorization
- [ ] Build reference implementations for each pattern
- [ ] **ADR:** Document authentication architecture by app type

---

### B.9 Multi-Factor Authentication (MFA)

**Goal:** Implement and understand MFA mechanisms

**Learning objectives:**
- Understand MFA factors and methods
- Implement TOTP and WebAuthn
- Design MFA enrollment and recovery flows

**Tasks:**
- [ ] Create `experiments/scenarios/mfa-implementation/`
- [ ] MFA fundamentals:
  - [ ] Something you know, have, are
  - [ ] MFA vs 2FA vs 2SV
  - [ ] Risk-based authentication
- [ ] TOTP (RFC 6238):
  - [ ] Algorithm internals (HMAC-SHA1, time step)
  - [ ] QR code provisioning (otpauth:// URI)
  - [ ] Clock drift handling
  - [ ] Backup codes
- [ ] WebAuthn / FIDO2:
  - [ ] Authenticator types (platform, roaming)
  - [ ] Registration ceremony
  - [ ] Authentication ceremony
  - [ ] Resident keys / discoverable credentials
  - [ ] User verification vs user presence
- [ ] SMS/Email OTP:
  - [ ] Implementation patterns
  - [ ] Security limitations (SIM swap, interception)
  - [ ] When acceptable vs not
- [ ] Push notification:
  - [ ] Approve/deny flow
  - [ ] Number matching
  - [ ] Implementation complexity
- [ ] MFA in Keycloak:
  - [ ] OTP policy configuration
  - [ ] WebAuthn configuration
  - [ ] Conditional MFA (authentication flows)
- [ ] Recovery mechanisms:
  - [ ] Backup codes
  - [ ] Recovery email/phone
  - [ ] Admin reset procedures
  - [ ] Account recovery vs security balance
- [ ] Implement MFA demo application
- [ ] **ADR:** Document MFA strategy

---

### B.10 Authorization Patterns

**Goal:** Understand authorization models beyond authentication

**Learning objectives:**
- Understand RBAC, ABAC, and ReBAC models
- Implement fine-grained authorization
- Integrate with policy engines

**Tasks:**
- [ ] Create `experiments/scenarios/authorization-patterns/`
- [ ] Authorization fundamentals:
  - [ ] Authentication vs Authorization
  - [ ] Coarse-grained vs fine-grained
  - [ ] Centralized vs distributed decisions
- [ ] Role-Based Access Control (RBAC):
  - [ ] Users, roles, permissions
  - [ ] Role hierarchy
  - [ ] Kubernetes RBAC mapping
- [ ] Attribute-Based Access Control (ABAC):
  - [ ] Attributes: subject, resource, action, environment
  - [ ] Policy language concepts
  - [ ] Dynamic authorization
- [ ] Relationship-Based Access Control (ReBAC):
  - [ ] Object relationships (owner, editor, viewer)
  - [ ] Google Zanzibar model
  - [ ] Graph-based authorization
- [ ] Policy engines:
  - [ ] Open Policy Agent (OPA)
  - [ ] Cedar (AWS)
  - [ ] Casbin
  - [ ] SpiceDB (Zanzibar implementation)
- [ ] OAuth scopes vs permissions:
  - [ ] Scopes as consent boundaries
  - [ ] Permissions as fine-grained access
  - [ ] Scope-to-permission mapping
- [ ] Token claims for authorization:
  - [ ] Groups/roles in tokens
  - [ ] Permission claims
  - [ ] Claim size vs lookup trade-off
- [ ] Implement OPA-based authorization
- [ ] Implement SpiceDB for ReBAC
- [ ] **ADR:** Document authorization model selection

---

### B.11 Security Operations

**Goal:** Operate identity infrastructure securely

**Learning objectives:**
- Monitor and audit authentication systems
- Respond to identity-related incidents
- Implement security controls and hardening

**Tasks:**
- [ ] Create `experiments/scenarios/identity-secops/`
- [ ] Logging and auditing:
  - [ ] Authentication events (success, failure)
  - [ ] Authorization decisions
  - [ ] Token issuance and revocation
  - [ ] Admin actions
- [ ] Monitoring:
  - [ ] Failed login rates (brute force detection)
  - [ ] Token error rates
  - [ ] Latency metrics
  - [ ] IdP availability
- [ ] Alerting:
  - [ ] Credential stuffing patterns
  - [ ] Account takeover indicators
  - [ ] Anomalous access patterns
  - [ ] Certificate expiration
- [ ] Incident response:
  - [ ] Mass token revocation
  - [ ] Emergency credential reset
  - [ ] IdP compromise procedures
- [ ] Hardening:
  - [ ] TLS configuration
  - [ ] Key rotation schedules
  - [ ] Secrets management integration (OpenBao)
  - [ ] Network segmentation
- [ ] Compliance:
  - [ ] Audit trail requirements
  - [ ] Data retention policies
  - [ ] GDPR considerations (consent, data access)
- [ ] Build identity security dashboard
- [ ] Create incident response runbook
- [ ] **ADR:** Document identity security monitoring strategy

---

### B.12 API Keys & Personal Access Tokens

**Goal:** Understand non-interactive authentication mechanisms for developers and automation

**Learning objectives:**
- Understand when to use API keys vs OAuth vs PATs
- Implement secure API key management
- Design PAT systems with appropriate scoping

**Tasks:**
- [ ] Create `experiments/scenarios/api-key-management/`
- [ ] API keys fundamentals:
  - [ ] What API keys are (shared secrets, not user-bound)
  - [ ] API keys vs OAuth tokens (no user context, no expiration by default)
  - [ ] Use cases: public APIs, rate limiting, billing attribution
  - [ ] When NOT to use API keys (user data access, sensitive operations)
- [ ] API key security:
  - [ ] Generation (CSPRNG, sufficient entropy, 256+ bits)
  - [ ] Storage (hash server-side, never store plaintext)
  - [ ] Prefix patterns for identification (sk_live_, pk_test_)
  - [ ] Key rotation without downtime
  - [ ] Revocation mechanisms
- [ ] API key scoping:
  - [ ] Read vs write permissions
  - [ ] Resource-level restrictions
  - [ ] IP allowlisting
  - [ ] Rate limiting per key
- [ ] Personal Access Tokens (PATs):
  - [ ] PATs as user-scoped credentials for automation
  - [ ] Difference from API keys (user context, audit trail)
  - [ ] GitHub/GitLab PAT patterns
  - [ ] Fine-grained vs classic tokens
- [ ] PAT implementation:
  - [ ] Token generation and display (show once)
  - [ ] Scope selection UI
  - [ ] Expiration policies (required vs optional)
  - [ ] Token naming/description for management
  - [ ] Last used tracking
- [ ] PAT security:
  - [ ] Hashing (like passwords)
  - [ ] Binding to user lifecycle (delete on user deletion)
  - [ ] Scope validation on every request
  - [ ] Activity logging
- [ ] Service accounts vs PATs:
  - [ ] When to use service accounts (CI/CD, automation)
  - [ ] When to use PATs (personal scripts, testing)
  - [ ] Shared credentials anti-pattern
- [ ] API key/PAT in practice:
  - [ ] Header patterns (Authorization: Bearer, X-API-Key)
  - [ ] Query parameter risks (logs, referrer leakage)
  - [ ] Client SDK patterns
- [ ] Management UI:
  - [ ] List active tokens
  - [ ] Revoke individual tokens
  - [ ] Bulk revocation
  - [ ] Usage analytics
- [ ] Implement API key service
- [ ] Implement PAT system with scoping
- [ ] **ADR:** Document API key vs OAuth vs PAT selection criteria

---
