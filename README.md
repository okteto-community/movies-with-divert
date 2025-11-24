# Okteto Divert Demo - Movies App

## What is Divert?

Okteto's Divert feature allows developers to work on individual microservices without deploying the entire application stack. It uses HTTP header-based routing to intelligently route traffic between shared services and your personal development instances.

**Note**: This demo uses the **Nginx driver with Linkerd** for header-based routing. Okteto Divert also works natively with **Istio** (no Linkerd required). Choose the driver based on your existing service mesh infrastructure.

### Key Benefits

- **Massive Resource Savings**: Deploy only the services you're working on
- **Faster Setup**: Environment ready in seconds instead of waiting for all services to deploy
- **Isolation**: Your changes don't affect other developers
- **Production-like**: Test against real shared services
- **Cost Efficient**: Share expensive infrastructure (databases, message queues)

### How It Works

1. **Shared Namespace**: A complete Movies app stack runs in a shared staging namespace
2. **Personal Namespace**: Your namespace contains only the service(s) you're developing
3. **Smart Routing**: Nginx ingress (with Linkerd sidecar) routes requests with the `baggage: okteto-divert=<your-namespace>` header to your services
4. **Header Propagation**: The baggage header is automatically propagated through all service calls

### Divert Drivers

Okteto Divert supports two drivers:

- **Nginx + Linkerd** (this demo): Uses Nginx ingress controller with Linkerd service mesh for header-based routing
- **Istio** (native): Uses Istio's built-in VirtualService for header-based routing without requiring additional components

To use Istio instead, change `driver: nginx` to `driver: istio` in the divert configurations.

## Quick Start

### Prerequisites

- Okteto CLI installed (`brew install okteto` or download from [okteto.com](https://okteto.com))
- kubectl configured
- Access to an Okteto cluster

### Setup Steps

#### Deploy the Shared Environment

1. **Clone the repository**
   ```bash
   git clone https://github.com/okteto/movies
   cd movies
   ```

2. **Deploy the shared environment **
   ```bash
   okteto preview deploy --repository https://github.com/okteto-community/movies-with-divert --label=okteto-shared movies-shared
   ```

The movies application is composed of five microservices: Frontend, API Gateway, Catalog, Rent, and Worker. Each service is defined in a subfolder. Note that Okteto Divert works independently of the deployment mechanism or architecture. 

This is what the application looks like:

![Architecture diagram](docs/architecture-diagram.png)

#### Deploy your personal Development Environment

The key advantage of divert is that you only need to deploy the service(s) you are actively working, rather than the full application.

To deploy your development environment export the `OKTETO_SHARED_NAMESPACE` environment variable, and run the `okteto deploy` command with the corresponding okteto manifest. This repository contains samples for configurations below.

Available Divert Configurations:

1. Frontend Development (`okteto.frontend.yaml`)
**Use when**: Working on React UI components, user interactions, or frontend features

**What it deploys**:
- Frontend service only (React/Node.js)

**Shares**:
- Catalog service
- API Gateway service
- Rent service
- Worker service
- Infrastructure

2. Catalog Development (`okteto.catalog.yaml`)
**Use when**: Working on movie catalog, inventory management, or MongoDB integration

**What it deploys**:
- Catalog service only (Node.js/Express)

**Shares**:
- Frontend
- API service
- Rent service
- Worker service
- Infrastructure


3. API Gateway Development (`okteto.api.yaml`)
**Use when**: Working on API endpoints, user management, or PostgreSQL integration

**What it deploys**:
- API Gateway service only (Golang)

**Shares**:
- Frontend
- Catalog service
- Rent service
- Worker Service
- Infrastructure

4. Rent Development (`okteto.rent.yaml`)
**Use when**: Working on rental logic, Kafka integration, or Spring Boot backend

**What it deploys**:
- Rent service only (Java/Spring Boot)

**Shares**:
- Frontend
- Catalog service
- API service
- Worker Service
- Infrastructure

4. Worker Development (`okteto.worker.yaml`)
**Use when**: Working on message processing logic

**What it deploys**:
- Worker service only (Golang)

**Shares**:
- Frontend
- Catalog service
- API service
- Rent Service
- Infrastructure


## Baggage Header Propagation
All the servicers of the movies app have been instrumented with baggage header propagation to ensure Divert routing works seamlessly across all services.


## Testing Your Divert Setup

### 1. Test Header Propagation Locally

```bash
# Start each service and verify header logging
curl -H "baggage: okteto-divert=test" http://localhost:8080/catalog
```

### 2. Test Divert in Okteto

```bash
# Test with baggage header (routes to your namespace)
curl -H "baggage: okteto-divert=alice-movies" https://movies-movies-staging.okteto.dev

# Test direct access (bypasses Divert)
curl https://movies-alice-movies.okteto.dev
```

### 3. Verify Divert Resources

```bash
# Check Divert custom resources
kubectl get diverts -n alice-movies

# Check HTTPRoutes in shared namespace
kubectl get httproutes -n movies-staging | grep okteto-

# Check service endpoints
kubectl get endpoints -n alice-movies
```

### 4. Browser Testing

1. Open the shared staging URL: `https://movies-movies-staging.okteto.dev`
2. Add query parameter: `?baggage=okteto-divert%3Dalice-movies`
3. Verify your service changes appear
4. Check that other services work normally (shared from staging)


## Troubleshooting

### Issue: Services can't communicate

**Solution**: Verify network policies allow cross-namespace communication
```bash
kubectl get networkpolicies -n alice-movies
kubectl get networkpolicies -n movies-staging
```

### Issue: Baggage header not propagating

**Solution**: Check service logs for header values
```bash
kubectl logs -n alice-movies deployment/frontend -f
```

### Issue: Divert routing not working

**Solution**: Verify Divert custom resource is created
```bash
kubectl describe divert -n alice-movies
kubectl get httproutes -n movies-staging
```

### Issue: Can't connect to shared databases

**Solution**: Verify service discovery and DNS
```bash
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- nslookup mongodb.movies-staging.svc.cluster.local
```

### Issue: Linkerd not injecting sidecar (Nginx driver only)

**Solution**: Ensure shared namespace has `okteto-shared` label for Linkerd sidecar injection
```bash
kubectl get namespace movies-staging -o yaml | grep okteto-shared
```

**Note**: This issue only applies when using the Nginx driver. Istio driver does not require Linkerd.


## Best Practices

1. **Use descriptive namespace names**: `<your-name>-movies` or `<feature-name>-movies`
2. **Clean up when done**: Delete personal namespaces after development
   ```bash
   okteto namespace delete alice-movies
   ```
3. **Monitor resource usage**: Check your namespace doesn't exceed quotas
4. **Keep shared staging updated**: Regularly update the shared environment
5. **Use preview environments**: For major changes, consider full preview environments

## Advanced Usage

### Switching to Istio Driver

This demo is configured for **Nginx + Linkerd**. To use **Istio** instead:

1. Edit the divert configuration files (e.g., `okteto-frontend-divert.yaml`)
2. Change the driver:
   ```yaml
   divert:
     namespace: ${SHARED_NAMESPACE:-staging}
     driver: istio  # Changed from 'nginx'
   ```
3. Deploy normally - Istio will handle routing without Linkerd


## Support

- **Documentation**: [okteto.com/docs/divert](https://okteto.com/docs)
- **GitHub Issues**: [github.com/okteto/movies/issues](https://github.com/okteto/movies/issues)
- **Community**: [community.okteto.com](https://community.okteto.com)
- **Slack**: [okteto.com/slack](https://okteto.com/slack)

## License
Apache License 2.0 - See LICENSE file for details