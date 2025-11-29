# cert-manager for TLS/SSL

This directory contains cert-manager configuration for automatic TLS certificate management.

## What's Installed

- **cert-manager**: Kubernetes-native certificate management
- **Webhook**: Validates certificate requests
- **CA Injector**: Injects CA bundles into resources

## Certificate Issuers

Three ClusterIssuers are configured:

### 1. selfsigned-issuer (Local Testing)
- Creates self-signed certificates
- Good for local development
- Browsers will show security warnings
- No external dependencies

```yaml
issuerRef:
  name: selfsigned-issuer
  kind: ClusterIssuer
```

### 2. letsencrypt-staging (Testing)
- Uses Let's Encrypt staging environment
- No rate limits
- Certificates not trusted by browsers
- Use for testing before production

**Requirements**:
- Public domain name
- Ingress controller (nginx, traefik, etc.)
- Update email in `issuers.yaml`

```yaml
issuerRef:
  name: letsencrypt-staging
  kind: ClusterIssuer
```

### 3. letsencrypt-prod (Production)
- Uses Let's Encrypt production environment
- Trusted by all browsers
- Rate limits apply (50 certs/week per domain)
- Only use after testing with staging

**Requirements**:
- Same as staging
- Test with staging first!

```yaml
issuerRef:
  name: letsencrypt-prod
  kind: ClusterIssuer
```

## Usage Examples

### Create a Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-app-tls
  namespace: default
spec:
  secretName: my-app-tls-secret
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
  dnsNames:
    - my-app.local
    - my-app.example.com
```

### Use with Ingress (Automatic)

Add annotation to your Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  annotations:
    cert-manager.io/cluster-issuer: selfsigned-issuer
spec:
  tls:
    - hosts:
        - my-app.local
      secretName: my-app-tls  # cert-manager will create this
  rules:
    - host: my-app.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 80
```

### Check Certificate Status

```bash
# List certificates
kubectl get certificates -A

# Check specific certificate
kubectl describe certificate my-app-tls -n default

# Check the secret was created
kubectl get secret my-app-tls-secret -n default
```

## Integration with Gateway API

For Gateway API (not traditional Ingress), create a Gateway with TLS:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: my-gateway
  annotations:
    cert-manager.io/cluster-issuer: selfsigned-issuer
spec:
  gatewayClassName: eg
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - name: my-gateway-tls
      allowedRoutes:
        namespaces:
          from: All
```

## Troubleshooting

### Certificate not issuing

```bash
# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager

# Check certificate events
kubectl describe certificate <name> -n <namespace>

# Check certificate request
kubectl get certificaterequest -n <namespace>
```

### Common Issues

1. **ACME challenges failing**: Ensure domain points to your cluster
2. **Rate limits**: Use staging issuer first
3. **Webhook errors**: Check cert-manager pods are running

## Monitoring

cert-manager exports Prometheus metrics:

```bash
kubectl port-forward -n cert-manager svc/cert-manager 9402:9402
curl http://localhost:9402/metrics
```

## See Also

- Example certificate: `../../examples/argocd-tls-certificate.yaml`
- [cert-manager docs](https://cert-manager.io/docs/)
