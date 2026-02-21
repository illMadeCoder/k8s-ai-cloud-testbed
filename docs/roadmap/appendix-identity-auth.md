# Identity & Authentication Series Roadmap

*Experiment-focused roadmap for application-level identity, authentication, and authorization. 10 experiments progressing from cryptographic primitives through protocol benchmarks to IdP comparison and authorization model shootouts. Each entry is scaffold-ready — enough detail to create the experiment YAML from.*

*This is distinct from Phase 3's infrastructure-level security (RBAC, admission control) and from [Appendix C: PKI & Certificates](appendix-pki-certs.md) which covers TLS/mTLS/cert-manager.*

---

## Series Overview

### Series Metadata

```json
{
  "id": "identity-and-auth",
  "name": "Identity & Authentication",
  "description": "Progressive exploration of identity infrastructure -- from JWT signing algorithms and password hashing costs through OAuth/OIDC/SAML protocol benchmarks to full IdP comparison and zero-trust integration patterns",
  "order": ["id-phb", "id-jwt", "id-oaf", "id-oidc", "id-saml", "id-idp", "id-prx", "id-sso", "id-mfa", "id-azp"],
  "color": "#e879f9"
}
```

**Naming convention:** All experiments use the `id-` prefix (identity), following the pattern established by `db-` (database) in the cloud-database-internals series. Each suffix is a 2-4 character mnemonic: `phb` = password hashing benchmark, `jwt` = JWT signing, `oaf` = OAuth flows, etc.

**Color rationale:** `#e879f9` (fuchsia/violet-400) — visually distinct from the existing palette (purple `#a78bfa`, green `#34d399`, orange `#f97316`, blue `#38bdf8`). Security/identity is commonly associated with purple/violet tones in security tooling UIs.

### Narrative Arc

| Phase | Experiments | Theme |
|-------|-------------|-------|
| 1. Cryptographic Foundations | `id-phb`, `id-jwt` | Raw primitive performance — no IdP needed |
| 2. Protocol Internals | `id-oaf`, `id-oidc`, `id-saml` | OAuth/OIDC/SAML protocol benchmarks with Keycloak |
| 3. Infrastructure Comparison | `id-idp`, `id-prx`, `id-sso` | IdP vs IdP, proxy vs proxy, federation cost |
| 4. Advanced Capabilities | `id-mfa`, `id-azp` | MFA mechanism costs, authorization model shootout |

### Appendix B.1–B.12 Coverage Matrix

| Appendix Section | Covered By |
|-----------------|------------|
| B.1 Password Management & Credential Storage | `id-phb` (primary) |
| B.2 JWT Fundamentals & Internals | `id-jwt` (primary), `id-oidc` (validation) |
| B.3 JWE – JSON Web Encryption | `id-jwt` (JWE sub-benchmark) |
| B.4 OAuth 2.0 Flows | `id-oaf` (primary) |
| B.5 OpenID Connect (OIDC) | `id-oidc` (primary), `id-saml` (comparison baseline), `id-sso` (federation) |
| B.6 Session Management | `id-oaf` (refresh tokens), `id-prx` (session cookies), `id-mfa` (session establishment) |
| B.7 Identity Provider Deployment | `id-idp` (primary), `id-saml` (SAML IdP config), `id-sso` (brokering) |
| B.8 Application Integration Patterns | `id-prx` (primary), `id-sso` (federation), `id-azp` (service-to-service) |
| B.9 Multi-Factor Authentication | `id-mfa` (primary) |
| B.10 Authorization Patterns | `id-azp` (primary) |
| B.11 Security Operations | Covered as security analysis sections in experiments 3–10 |
| B.12 API Keys & PATs | Deferred — see [Future Work](#future-work) |

---

## Prerequisites

### New Component Category: `components/identity/`

| Component | Source | Description |
|-----------|--------|-------------|
| `keycloak` | Helm: `bitnami/keycloak` (OCI registry) | Keycloak IdP with PostgreSQL. Params: realm config, SAML/OIDC client definitions |
| `keycloak-postgres` | Helm: `bitnami/postgresql` | PostgreSQL for Keycloak backend. Params: database name, credentials |
| `dex` | Helm: `dex/dex` (`https://charts.dexidp.io`) | Dex OIDC federation proxy. Params: static clients, connectors |
| `oauth2-proxy` | Helm: `oauth2-proxy/oauth2-proxy` (`https://oauth2-proxy.github.io/manifests`) | OAuth2 reverse proxy. Params: OIDC provider URL, client ID/secret |
| `pomerium` | Helm: `pomerium/pomerium` (`https://helm.pomerium.io`) | Identity-aware access proxy. Params: IdP config, routes, policies |
| `opa` | Helm or raw manifests | Open Policy Agent for authorization decisions. Params: policy bundle |
| `spicedb` | Helm: `authzed/spicedb` (`https://authzed.github.io/helm-charts`) | Zanzibar-style authorization. Params: schema, datastore config |

### Custom Bench Apps: `components/apps/`

| Component | Description | Used By |
|-----------|-------------|---------|
| `auth-hash-bench` | Password hashing benchmark service (Go). Exposes bcrypt/argon2id/scrypt/PBKDF2 endpoints with Prometheus histograms | `id-phb` |
| `auth-jwt-bench` | JWT sign/verify/encrypt benchmark service (Go). All algorithm families with histograms | `id-jwt` |
| `auth-oauth-bench` | OAuth flow benchmark client. Drives AuthCode+PKCE, Client Credentials, Device flows against an IdP | `id-oaf` |
| `auth-resource-api` | Simple protected API (validates tokens, returns 200/401). Resource server for auth experiments | `id-oaf` |
| `auth-oidc-validator` | OIDC token validation benchmark. Three strategies: local JWKS, introspection, UserInfo | `id-oidc` |
| `auth-saml-sp` | SAML Service Provider with instrumented assertion parsing and XML signature verification | `id-saml` |
| `auth-idp-bench` | IdP benchmark client. Authenticates against both Keycloak and Dex, measures token issuance | `id-idp` |
| `auth-mfa-bench` | MFA benchmark client. Simulates TOTP code generation and WebAuthn ceremonies | `id-mfa` |
| `auth-authz-bench` | Authorization benchmark API. Three backends: in-process RBAC, OPA sidecar, SpiceDB | `id-azp` |

All bench apps are CI-built as container images, following the existing `components/apps/` pattern.

### Load Testing: `components/testing/`

| Component | Description |
|-----------|-------------|
| `k6-auth-loadtest` | k6 load test scripts for auth experiments. Configurable scenarios for each experiment type |

### Domain Taxonomy Addition

Add to `site/data/_categories.json`:

```json
{
  "id": "identity",
  "name": "Identity & Auth",
  "description": "Authentication, authorization, SSO, IdP deployment, and identity federation",
  "subdomains": ["authentication", "authorization", "federation", "idp"]
}
```

Tag-to-domain mappings to add in `site/src/lib/experiments.ts`:

| Tag | Domain | Subdomain |
|-----|--------|-----------|
| `authentication`, `oauth`, `oidc`, `saml`, `jwt`, `sso`, `mfa`, `password-hashing` | `identity` | `authentication` |
| `authorization`, `rbac`, `abac`, `rebac`, `opa` | `identity` | `authorization` |
| `keycloak`, `dex`, `idp` | `identity` | `idp` |
| `identity-brokering`, `identity-federation` | `identity` | `federation` |
| `identity` | `identity` | — |

### Secrets in OpenBao

| OpenBao Path | Purpose |
|-------------|---------|
| `secret/experiments/keycloak-admin` | Keycloak admin credentials for realm configuration |
| `secret/experiments/oidc-client` | Shared OIDC client ID and secret for test apps |
| `secret/experiments/saml-keystore` | SAML signing certificate and key for SP/IdP |

Synced via ExternalSecret to the `experiments` namespace, following the existing pattern in `platform/manifests/external-secrets-config/`.

### Keycloak Realm Configuration

Experiments 3–9 share a reusable Kubernetes Job that configures Keycloak via its Admin REST API. The Job reads from a ConfigMap containing:

- Realm export JSON (realm name, token settings, login settings)
- Client definitions (confidential and public)
- Test users with pre-set passwords
- SAML SP metadata (experiment 5)
- Identity brokering configuration (experiment 8)

Deployed via ArgoCD sync waves:

1. **Wave 0:** PostgreSQL
2. **Wave 1:** Keycloak (depends on PostgreSQL)
3. **Wave 2:** Realm configuration Job
4. **Wave 3:** Application components (bench apps)
5. **Wave 4:** Load generator

---

## Phase 1: Cryptographic Foundations

### id-phb — Password Hashing Algorithm Benchmark

**Title:** Password Hashing Showdown: bcrypt vs Argon2id vs scrypt vs PBKDF2

**Description:** Benchmarks the four major password hashing algorithms across multiple cost parameters to measure the latency/security tradeoff. Answers "how much does recommended-strength hashing actually cost?" with hard numbers on hash time, verify time, throughput, and resource consumption.

**Hypothesis:** Argon2id at 128 MiB memory cost will produce hash times within 2x of bcrypt cost-12, but consume 10x more memory, making it the recommended choice for security-critical systems where memory is available, while PBKDF2 at NIST-recommended 600K iterations will be the fastest but least resistant to GPU-based attacks.

**Deploys:**
- `auth-hash-bench` — custom Go service exposing endpoints per algorithm, instrumented with Prometheus histograms
- `k6-auth-loadtest` — concurrent hash/verify requests at varying work-factor parameters
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `bcrypt_hash_p99` | `histogram_quantile(0.99, sum(rate(auth_hash_duration_seconds_bucket{algorithm="bcrypt",operation="hash",namespace=~"$EXPERIMENT"}[$DURATION])) by (le,cost))` | instant | seconds | bcrypt |
| `argon2_hash_p99` | `histogram_quantile(0.99, sum(rate(auth_hash_duration_seconds_bucket{algorithm="argon2id",operation="hash",namespace=~"$EXPERIMENT"}[$DURATION])) by (le,memory_mib))` | instant | seconds | argon2 |
| `hash_throughput_by_algorithm` | `sum(rate(auth_hash_operations_total{operation="hash",namespace=~"$EXPERIMENT"}[$DURATION])) by (algorithm)` | instant | ops/s | throughput |
| `hash_cpu_by_algorithm` | `sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",container="auth-hash-bench"}[5m])) by (pod)` | range | cores | resources |
| `hash_memory_by_algorithm` | `max(container_memory_working_set_bytes{namespace=~"$EXPERIMENT",container="auth-hash-bench"}) by (pod)` | range | bytes | resources |

**Tags:** `comparison`, `identity`, `authentication`, `password-hashing`

**GKE sizing:** `e2-standard-8` (Argon2id at 256 MiB memory cost needs headroom), 2 nodes

**Prerequisites:** None — no external infrastructure needed

**Appendix coverage:** B.1

---

### id-jwt — JWT Signing Algorithm Benchmark

**Title:** JWT Signing Speed: HMAC vs RSA vs ECDSA vs EdDSA

**Description:** Benchmarks JWT sign/verify operations across all major algorithm families (HS256, RS256, ES256, Ed25519, PS256), plus JWE encrypt/decrypt overhead for nested sign-then-encrypt tokens. Measures token size differences that affect HTTP header overhead.

**Hypothesis:** EdDSA (Ed25519) will deliver the lowest sign+verify combined latency among asymmetric algorithms and the smallest token size, while HMAC (HS256) remains fastest overall but requires shared secret distribution.

**Deploys:**
- `auth-jwt-bench` — JWT token service exposing sign/verify/encrypt endpoints per algorithm, with histogram instrumentation
- `k6-auth-loadtest`
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `hs256_sign_p99` | `histogram_quantile(0.99, sum(rate(auth_jwt_operation_seconds_bucket{algorithm="HS256",operation="sign",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | hmac |
| `es256_verify_p99` | `histogram_quantile(0.99, sum(rate(auth_jwt_operation_seconds_bucket{algorithm="ES256",operation="verify",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | ecdsa |
| `token_size_by_algorithm` | `avg(auth_jwt_token_size_bytes{namespace=~"$EXPERIMENT"}) by (algorithm)` | instant | bytes | size |
| `jwe_encrypt_overhead_p99` | `histogram_quantile(0.99, sum(rate(auth_jwt_operation_seconds_bucket{operation="encrypt",namespace=~"$EXPERIMENT"}[$DURATION])) by (le,algorithm))` | instant | seconds | jwe |
| `sign_throughput_by_algorithm` | `sum(rate(auth_jwt_operations_total{operation="sign",namespace=~"$EXPERIMENT"}[$DURATION])) by (algorithm)` | instant | ops/s | throughput |

**Tags:** `comparison`, `identity`, `authentication`, `jwt`

**GKE sizing:** `e2-standard-4`, 2 nodes

**Prerequisites:** None — no external infrastructure needed

**Appendix coverage:** B.2, B.3

---

## Phase 2: Protocol Internals

### id-oaf — OAuth 2.0 Flow Latency Benchmark

**Title:** OAuth 2.0 Grant Type Latencies: Authorization Code vs Client Credentials vs Device Flow

**Description:** Measures end-to-end flow completion time for each major OAuth 2.0 grant type using Keycloak as the authorization server. Benchmarks token exchange, client credentials under concurrency, and token refresh — the operations that happen on every API call in a real system.

**Hypothesis:** Client Credentials grant will complete in under 5 ms p99 (no redirect, no user context), while Authorization Code + PKCE will cost 50–200 ms due to redirect round-trips. Token refresh will match client credentials latency since it is a direct token exchange.

**Deploys:**
- `keycloak` + `keycloak-postgres` — authorization server
- `auth-oauth-bench` — OAuth client implementing AuthCode+PKCE, Client Credentials, and Device Authorization flows
- `auth-resource-api` — simulated resource server that validates tokens
- `k6-auth-loadtest`
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `authcode_flow_p99` | `histogram_quantile(0.99, sum(rate(auth_oauth_flow_duration_seconds_bucket{flow="authorization_code",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | authcode |
| `client_credentials_p99` | `histogram_quantile(0.99, sum(rate(auth_oauth_flow_duration_seconds_bucket{flow="client_credentials",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | client_creds |
| `token_refresh_p99` | `histogram_quantile(0.99, sum(rate(auth_oauth_flow_duration_seconds_bucket{flow="refresh",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | refresh |
| `keycloak_cpu` | `sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",pod=~"keycloak.*",container!="POD",container!=""}[5m]))` | range | cores | idp_resources |
| `keycloak_memory` | `sum(container_memory_working_set_bytes{namespace=~"$EXPERIMENT",pod=~"keycloak.*",container!="POD",container!=""})` | range | bytes | idp_resources |
| `oauth_error_rate` | `sum(rate(auth_oauth_errors_total{namespace=~"$EXPERIMENT"}[$DURATION])) by (flow,error_type)` | instant | errors/s | errors |

**Tags:** `comparison`, `identity`, `authentication`, `oauth`, `keycloak`

**GKE sizing:** `e2-standard-4`, 2 nodes

**Prerequisites:** Keycloak component must exist in `components/identity/`

**Appendix coverage:** B.4, B.6

---

### id-oidc — OIDC Discovery and Token Validation Benchmark

**Title:** OIDC Token Validation: Local JWKS vs Introspection vs UserInfo Endpoint

**Description:** Compares three OIDC token validation strategies at the resource server: local JWKS-cached signature verification (fastest, but misses revocations), token introspection endpoint (real-time revocation at the cost of IdP load), and UserInfo endpoint fetch (most complete claims, highest latency). Measures the trade-off between validation speed and revocation awareness.

**Hypothesis:** Local JWKS validation will be 100x+ faster than introspection (sub-millisecond vs network round-trip) and impose zero load on the IdP, but will miss revoked tokens until the JWKS TTL expires. Introspection provides real-time revocation at the cost of making the IdP a throughput bottleneck.

**Deploys:**
- `keycloak` + `keycloak-postgres`
- `auth-oidc-validator` — resource server implementing all three validation strategies
- `k6-auth-loadtest` — sends pre-authenticated requests with valid tokens
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `local_jwks_validation_p99` | `histogram_quantile(0.99, sum(rate(auth_oidc_validation_seconds_bucket{strategy="local_jwks",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | local_jwks |
| `introspection_validation_p99` | `histogram_quantile(0.99, sum(rate(auth_oidc_validation_seconds_bucket{strategy="introspection",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | introspection |
| `userinfo_validation_p99` | `histogram_quantile(0.99, sum(rate(auth_oidc_validation_seconds_bucket{strategy="userinfo",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | userinfo |
| `jwks_cache_hit_rate` | `sum(rate(auth_oidc_jwks_cache_hits_total{namespace=~"$EXPERIMENT"}[$DURATION])) / sum(rate(auth_oidc_jwks_lookups_total{namespace=~"$EXPERIMENT"}[$DURATION]))` | instant | ratio | cache |
| `validation_throughput_by_strategy` | `sum(rate(auth_oidc_validations_total{namespace=~"$EXPERIMENT"}[$DURATION])) by (strategy)` | instant | ops/s | throughput |

**Tags:** `comparison`, `identity`, `authentication`, `oidc`, `keycloak`

**GKE sizing:** `e2-standard-4`, 2 nodes

**Prerequisites:** `id-oaf` (Keycloak deployment pattern established)

**Appendix coverage:** B.5, B.2

---

### id-saml — SAML 2.0 Protocol Performance Benchmark

**Title:** SAML 2.0 SSO: Assertion Processing, XML Signature Cost, and Redirect Latency

**Description:** Benchmarks SAML 2.0 SSO flows side-by-side with OIDC to quantify the protocol overhead difference. Measures SAML-specific costs: XML digital signature creation/verification, assertion serialization, and payload size compared to compact JWTs. Uses Keycloak configured as both SAML IdP and OIDC provider for a controlled comparison.

**Hypothesis:** SAML SSO flows will incur 2–5x higher end-to-end latency than OIDC due to XML serialization, XML digital signature computation, and larger assertion payloads (4–10 KB vs 500-byte JWT). SAML XML signature verification on the SP side will be the dominant cost, not the redirect round-trips.

**Deploys:**
- `keycloak` + `keycloak-postgres` — configured as both SAML IdP and OIDC provider
- `auth-saml-sp` — SAML service provider with instrumented assertion parsing and XML signature verification
- OIDC counterpart app for side-by-side comparison
- `k6-auth-loadtest`
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `saml_assertion_gen_p99` | `histogram_quantile(0.99, sum(rate(auth_saml_assertion_duration_seconds_bucket{operation="generate",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | saml_idp |
| `saml_xml_signature_verify_p99` | `histogram_quantile(0.99, sum(rate(auth_saml_assertion_duration_seconds_bucket{operation="verify_signature",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | saml_sp |
| `saml_sso_flow_p99` | `histogram_quantile(0.99, sum(rate(auth_sso_flow_duration_seconds_bucket{protocol="saml",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | sso_flow |
| `oidc_sso_flow_p99` | `histogram_quantile(0.99, sum(rate(auth_sso_flow_duration_seconds_bucket{protocol="oidc",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | sso_flow |
| `saml_assertion_size` | `avg(auth_saml_assertion_size_bytes{namespace=~"$EXPERIMENT"})` | instant | bytes | payload_size |
| `oidc_id_token_size` | `avg(auth_oidc_token_size_bytes{namespace=~"$EXPERIMENT"})` | instant | bytes | payload_size |
| `saml_xml_parse_p99` | `histogram_quantile(0.99, sum(rate(auth_saml_xml_parse_duration_seconds_bucket{namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | saml_sp |
| `keycloak_cpu_saml_vs_oidc` | `sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",pod=~"keycloak.*",container!="POD",container!=""}[5m]))` | range | cores | idp_resources |

**Tags:** `comparison`, `identity`, `authentication`, `saml`, `oidc`, `keycloak`

**GKE sizing:** `e2-standard-4`, 2 nodes

**Prerequisites:** `id-oaf` (Keycloak deployment pattern established)

**Appendix coverage:** B.7 (SAML brokering), B.5 (OIDC comparison baseline)

---

## Phase 3: Infrastructure Comparison

### id-idp — Identity Provider Comparison: Keycloak vs Dex

**Title:** IdP Shootout: Keycloak vs Dex — Resource Footprint, Startup, and Token Throughput

**Description:** Head-to-head comparison of a full-featured IdP (Keycloak with PostgreSQL) vs a lightweight federation proxy (Dex, stateless). Measures cold start time, idle resource footprint, token issuance throughput under load (50/100/200 concurrent users), and the resource cost of Keycloak's richer feature set (user federation, admin UI, fine-grained authorization).

**Hypothesis:** Dex will use 5–10x less memory than Keycloak at idle (50–100 MB vs 500+ MB due to JVM) and start 10x faster, but Keycloak will deliver higher sustained token throughput due to session caching and connection pooling. Including PostgreSQL, Keycloak's total footprint will be 3–5x larger.

**Deploys:**
- `keycloak` + `keycloak-postgres`
- `dex` — with static configuration, identical OIDC client registrations
- `auth-idp-bench` — shared test client authenticating against both IdPs
- `k6-auth-loadtest`
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `keycloak_startup_time` | `auth_idp_startup_seconds{idp="keycloak",namespace=~"$EXPERIMENT"}` | instant | seconds | startup |
| `dex_startup_time` | `auth_idp_startup_seconds{idp="dex",namespace=~"$EXPERIMENT"}` | instant | seconds | startup |
| `keycloak_idle_cpu` | `avg_over_time(sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",pod=~"keycloak.*",container!="POD",container!=""}[5m]))[5m:])` | instant | cores | idle |
| `dex_idle_cpu` | `avg_over_time(sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",pod=~"dex.*",container!="POD",container!=""}[5m]))[5m:])` | instant | cores | idle |
| `keycloak_token_throughput` | `sum(rate(auth_idp_token_issuance_total{idp="keycloak",namespace=~"$EXPERIMENT"}[$DURATION]))` | instant | ops/s | throughput |
| `dex_token_throughput` | `sum(rate(auth_idp_token_issuance_total{idp="dex",namespace=~"$EXPERIMENT"}[$DURATION]))` | instant | ops/s | throughput |
| `keycloak_memory_peak` | `max_over_time(sum(container_memory_working_set_bytes{namespace=~"$EXPERIMENT",pod=~"keycloak.*",container!="POD",container!=""})[${DURATION}:])` | instant | bytes | resources |
| `dex_memory_peak` | `max_over_time(sum(container_memory_working_set_bytes{namespace=~"$EXPERIMENT",pod=~"dex.*",container!="POD",container!=""})[${DURATION}:])` | instant | bytes | resources |
| `postgres_memory` | `sum(container_memory_working_set_bytes{namespace=~"$EXPERIMENT",pod=~"postgres.*",container!="POD",container!=""})` | range | bytes | dependencies |

**Tags:** `comparison`, `identity`, `idp`, `keycloak`, `dex`

**GKE sizing:** `e2-standard-4`, 2 nodes (Keycloak ~2 GB + Dex ~100 MB + observability fits)

**Prerequisites:** `id-oaf` (Keycloak component), Dex component must exist in `components/identity/`

**Appendix coverage:** B.7

---

### id-prx — Auth Proxy Comparison: OAuth2-Proxy vs Pomerium vs Envoy ext_authz

**Title:** Auth Proxy Shootout: OAuth2-Proxy vs Pomerium vs Envoy ext_authz

**Description:** Compares three approaches to adding authentication in front of an existing application: OAuth2-Proxy (standalone reverse proxy), Pomerium (identity-aware proxy with built-in policy engine), and Envoy with ext_authz filter using an OPA sidecar. Measures the per-request auth overhead added by each approach, plus session cookie validation, token refresh handling, and policy evaluation cost.

**Hypothesis:** Envoy ext_authz with OPA will add the least per-request latency (sub-2ms) because the authz decision is an in-process gRPC call, while OAuth2-Proxy will add 5–15 ms due to session cookie decryption and upstream token validation. Pomerium will land between them due to its built-in policy engine.

**Deploys:**
- `keycloak` + `keycloak-postgres` — shared OIDC provider
- `oauth2-proxy`, `pomerium`, and Envoy with ext_authz + OPA sidecar — all fronting the same backend app (nginx)
- `k6-auth-loadtest` — both authenticated and unauthenticated requests
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `oauth2proxy_auth_overhead_p99` | `histogram_quantile(0.99, sum(rate(auth_proxy_request_duration_seconds_bucket{proxy="oauth2-proxy",namespace=~"$EXPERIMENT"}[$DURATION])) by (le)) - histogram_quantile(0.99, sum(rate(auth_proxy_request_duration_seconds_bucket{proxy="direct",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | overhead |
| `pomerium_auth_overhead_p99` | `histogram_quantile(0.99, sum(rate(auth_proxy_request_duration_seconds_bucket{proxy="pomerium",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | overhead |
| `envoy_extauthz_p99` | `histogram_quantile(0.99, sum(rate(auth_proxy_request_duration_seconds_bucket{proxy="envoy-extauthz",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | overhead |
| `proxy_cpu_by_type` | `sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",pod=~"oauth2-proxy.*\|pomerium.*\|envoy-authz.*",container!="POD",container!=""}[5m])) by (pod)` | range | cores | resources |
| `proxy_memory_by_type` | `sum(container_memory_working_set_bytes{namespace=~"$EXPERIMENT",pod=~"oauth2-proxy.*\|pomerium.*\|envoy-authz.*",container!="POD",container!=""}) by (pod)` | range | bytes | resources |

**Tags:** `comparison`, `identity`, `authentication`, `sso`, `oauth`

**GKE sizing:** `e2-standard-4`, 2 nodes

**Prerequisites:** `id-oaf` (Keycloak component), proxy components must exist in `components/identity/`

**Appendix coverage:** B.8

---

### id-sso — Cross-Protocol SSO: SAML + OIDC Federation

**Title:** Cross-Protocol SSO: SAML-to-OIDC Federation via Keycloak Identity Brokering

**Description:** Measures the cost of cross-protocol federation: a downstream Keycloak realm provides OIDC tokens while brokering authentication to an upstream SAML IdP. Compares direct OIDC SSO latency vs SAML-brokered OIDC to quantify the protocol translation overhead. Also measures attribute/claim mapping fidelity — do claims survive federation?

**Hypothesis:** SAML-brokered OIDC SSO will add 100–300 ms over direct OIDC due to the double redirect (SP to broker to SAML IdP and back) and XML signature processing at the broker. The protocol translation itself (SAML assertion to OIDC claims mapping) will contribute less than 10 ms.

**Deploys:**
- Keycloak "upstream" realm as SAML IdP (simulating enterprise SAML source)
- Keycloak "downstream" realm as OIDC provider, brokering to upstream SAML realm
- Service provider app consuming OIDC tokens from downstream realm
- `k6-auth-loadtest`
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `direct_oidc_sso_p99` | `histogram_quantile(0.99, sum(rate(auth_sso_flow_duration_seconds_bucket{flow="direct_oidc",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | sso_flow |
| `brokered_saml_oidc_sso_p99` | `histogram_quantile(0.99, sum(rate(auth_sso_flow_duration_seconds_bucket{flow="saml_brokered_oidc",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | sso_flow |
| `protocol_translation_p99` | `histogram_quantile(0.99, sum(rate(auth_broker_translation_duration_seconds_bucket{namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | broker |
| `broker_keycloak_cpu` | `sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",pod=~"keycloak.*",container!="POD",container!=""}[5m]))` | range | cores | resources |
| `claim_mapping_accuracy` | `sum(rate(auth_broker_claim_mapping_total{result="matched",namespace=~"$EXPERIMENT"}[$DURATION])) / sum(rate(auth_broker_claim_mapping_total{namespace=~"$EXPERIMENT"}[$DURATION]))` | instant | ratio | fidelity |

**Tags:** `comparison`, `identity`, `sso`, `saml`, `oidc`, `identity-brokering`

**GKE sizing:** `e2-standard-4`, 2 nodes

**Prerequisites:** `id-saml` (SAML configuration pattern), `id-oidc` (OIDC baseline)

**Appendix coverage:** B.5, B.7, B.8

---

## Phase 4: Advanced Capabilities

### id-mfa — Multi-Factor Authentication Latency Impact

**Title:** MFA Cost: TOTP vs WebAuthn vs Conditional MFA on Authentication Throughput

**Description:** Measures the server-side latency each MFA mechanism adds to the authentication flow. Compares TOTP (software token — HMAC computation + clock drift window), WebAuthn (simulated FIDO2 — challenge generation + public key verification), and conditional/risk-based MFA (only triggered above a threshold). The dominant MFA cost is user interaction time, which is not measured here — this isolates the server-side processing overhead.

**Hypothesis:** TOTP adds minimal server-side latency (sub-millisecond HMAC validation) but WebAuthn's challenge-response ceremony and public key verification will add 5–15 ms of server-side processing. The dominant MFA cost is user interaction time (not measured here), making server-side overhead negligible for all methods.

**Deploys:**
- `keycloak` + `keycloak-postgres` — configured with TOTP, WebAuthn, and conditional MFA modes
- `auth-mfa-bench` — authentication client simulating logins with each MFA mode, including TOTP code generation and WebAuthn ceremony simulation
- `k6-auth-loadtest`
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `no_mfa_auth_p99` | `histogram_quantile(0.99, sum(rate(auth_mfa_login_duration_seconds_bucket{mfa_mode="none",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | baseline |
| `totp_auth_p99` | `histogram_quantile(0.99, sum(rate(auth_mfa_login_duration_seconds_bucket{mfa_mode="totp",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | totp |
| `webauthn_auth_p99` | `histogram_quantile(0.99, sum(rate(auth_mfa_login_duration_seconds_bucket{mfa_mode="webauthn",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | webauthn |
| `conditional_mfa_trigger_rate` | `sum(rate(auth_mfa_conditional_triggered_total{namespace=~"$EXPERIMENT"}[$DURATION])) / sum(rate(auth_mfa_conditional_evaluated_total{namespace=~"$EXPERIMENT"}[$DURATION]))` | instant | ratio | conditional |
| `auth_throughput_by_mfa_mode` | `sum(rate(auth_mfa_logins_total{namespace=~"$EXPERIMENT"}[$DURATION])) by (mfa_mode)` | instant | ops/s | throughput |

**Tags:** `comparison`, `identity`, `authentication`, `mfa`, `keycloak`

**GKE sizing:** `e2-standard-4`, 2 nodes

**Prerequisites:** `id-oaf` (Keycloak component and realm configuration pattern)

**Appendix coverage:** B.9, B.6

---

### id-azp — Authorization Model Benchmark: RBAC vs OPA vs SpiceDB

**Title:** Authorization Model Shootout: K8s RBAC vs OPA vs SpiceDB (Zanzibar)

**Description:** Benchmarks three authorization models at increasing complexity: in-process RBAC (role lookup from JWT claims — effectively a map lookup), OPA sidecar (Rego policy evaluation via localhost gRPC with a moderately complex policy: 10 rules, 5 roles, resource-level permissions), and SpiceDB (Zanzibar-style relationship check via gRPC — graph traversal for Google Docs-style sharing models). Measures per-request decision latency, throughput under concurrency, and the resource cost of each engine.

**Hypothesis:** In-process RBAC will be fastest (sub-microsecond, just a map lookup on JWT claims), OPA sidecar will add 0.5–2 ms per decision due to Rego evaluation, and SpiceDB will cost 2–10 ms due to gRPC round-trip + relationship graph traversal. SpiceDB's cost is justified only when relationship-based access patterns (Google Docs-style sharing) are required.

**Deploys:**
- `auth-authz-bench` — API service with three authorization backends: in-process RBAC, OPA sidecar, SpiceDB
- `opa` — with moderately complex policy bundle
- `spicedb` — with embedded SQLite or CockroachDB backend
- `k6-auth-loadtest`
- `kube-prometheus-stack`, `metrics-agent`, `metrics-egress`

**Key metrics:**

| Name | Query | Type | Unit | Group |
|------|-------|------|------|-------|
| `rbac_decision_p99` | `histogram_quantile(0.99, sum(rate(auth_authz_decision_seconds_bucket{engine="rbac",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | rbac |
| `opa_decision_p99` | `histogram_quantile(0.99, sum(rate(auth_authz_decision_seconds_bucket{engine="opa",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | opa |
| `spicedb_decision_p99` | `histogram_quantile(0.99, sum(rate(auth_authz_decision_seconds_bucket{engine="spicedb",namespace=~"$EXPERIMENT"}[$DURATION])) by (le))` | instant | seconds | spicedb |
| `authz_throughput_by_engine` | `sum(rate(auth_authz_decisions_total{namespace=~"$EXPERIMENT"}[$DURATION])) by (engine)` | instant | ops/s | throughput |
| `opa_cpu` | `sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",pod=~"opa.*",container!="POD",container!=""}[5m]))` | range | cores | resources |
| `spicedb_cpu` | `sum(rate(container_cpu_usage_seconds_total{namespace=~"$EXPERIMENT",pod=~"spicedb.*",container!="POD",container!=""}[5m]))` | range | cores | resources |
| `spicedb_memory` | `sum(container_memory_working_set_bytes{namespace=~"$EXPERIMENT",pod=~"spicedb.*",container!="POD",container!=""})` | range | bytes | resources |

**Tags:** `comparison`, `identity`, `authorization`, `opa`, `rbac`

**GKE sizing:** `e2-standard-4`, 2 nodes

**Prerequisites:** OPA and SpiceDB components must exist in `components/identity/`

**Appendix coverage:** B.10, B.8

---

## Future Work

### id-apk — API Keys & Personal Access Tokens (B.12)

API key management is difficult to benchmark quantitatively — it is more of a design pattern than a performance-measurable system. A future experiment could benchmark API key lookup/validation strategies:

- Hash-based lookup (bcrypt/argon2 hash comparison on every request)
- HMAC-prefix lookup (use key prefix for DB lookup, then HMAC verification)
- Database lookup with caching (Redis/in-memory)
- PAT scoping and validation overhead

This was deliberately omitted from the initial 10 experiments. If added, it would cover B.12 and parts of B.1 (credential storage patterns applied to API keys).

---

## Metrics Summary

| # | Experiment | Primary Metrics | Secondary Metrics |
|---|-----------|-----------------|-------------------|
| 1 | `id-phb` | Hash time p99 per algorithm/cost, verify time, throughput | CPU/memory per algorithm |
| 2 | `id-jwt` | Sign p99, verify p99, JWE overhead, token size | Throughput per algorithm |
| 3 | `id-oaf` | Flow completion time per grant type, token exchange latency | Keycloak CPU/memory, error rates |
| 4 | `id-oidc` | Validation latency per strategy, JWKS cache hit rate | Validation throughput, IdP load |
| 5 | `id-saml` | SAML assertion gen time, XML sig verify time, SAML vs OIDC flow latency | Assertion size, XML parse overhead |
| 6 | `id-idp` | Startup time, idle footprint, token throughput, memory peak | PostgreSQL overhead, loaded CPU delta |
| 7 | `id-prx` | Auth overhead p99 per proxy, session cookie validation | Proxy CPU/memory, policy eval time |
| 8 | `id-sso` | Direct vs brokered SSO latency, protocol translation time | Claim mapping accuracy, broker CPU |
| 9 | `id-mfa` | Auth latency by MFA mode, TOTP/WebAuthn server processing | Conditional MFA trigger rate, throughput |
| 10 | `id-azp` | Decision latency per engine, decision throughput | OPA/SpiceDB CPU/memory, cache hit rate |

---

## Dependency Graph

```
Phase 1 (no IdP needed):
  id-phb ──┐
  id-jwt ──┤  (independent, can run in parallel)
           │
Phase 2 (Keycloak required):
           ├──► id-oaf ──┬──► id-oidc
           │             └──► id-saml
           │
Phase 3 (additional infrastructure):
           ├──► id-idp  (needs Keycloak + Dex)
           ├──► id-prx  (needs Keycloak + 3 proxies)
           └──► id-sso  (needs id-saml + id-oidc patterns)
                    │
Phase 4 (advanced):
                    ├──► id-mfa  (needs Keycloak)
                    └──► id-azp  (needs OPA + SpiceDB, independent of IdP experiments)
```

## Implementation Sequencing

### Phase 1: Foundation (Experiments 1–2)

No external infrastructure needed. Build the custom bench apps (Go microservices with Prometheus instrumentation) and k6 load scripts. These establish the identity component category and CI pipeline patterns.

**Create:** `components/apps/auth-hash-bench/`, `components/apps/auth-jwt-bench/`, `components/testing/k6-auth-loadtest/`, `experiments/id-phb/`, `experiments/id-jwt/`

### Phase 2: IdP Infrastructure (Experiments 3–5)

Deploy Keycloak component and the realm configuration pattern. Once Keycloak works, experiments 3–5 iterate rapidly.

**Create:** `components/identity/keycloak/`, `components/identity/keycloak-postgres/`, realm configuration Job manifests, `components/apps/auth-oauth-bench/`, `auth-oidc-validator/`, `auth-saml-sp/`, `experiments/id-oaf/`, `id-oidc/`, `id-saml/`

### Phase 3: Comparison Layer (Experiments 6–8)

Add Dex, OAuth2-Proxy, Pomerium, and Envoy ext_authz components. Mostly Helm charts, not custom code.

**Create:** `components/identity/dex/`, `oauth2-proxy/`, `pomerium/`, `components/apps/auth-idp-bench/`, `experiments/id-idp/`, `id-prx/`, `id-sso/`

### Phase 4: Advanced (Experiments 9–10)

MFA simulation and authorization engines. More complex test harnesses.

**Create:** `components/identity/opa/`, `spicedb/`, `components/apps/auth-mfa-bench/`, `auth-authz-bench/`, `experiments/id-mfa/`, `id-azp/`

### Site Changes (All Phases)

- Add identity domain to `site/data/_categories.json`
- Add series to `site/data/_series.json`
- Add tag mappings to `site/src/lib/experiments.ts`
