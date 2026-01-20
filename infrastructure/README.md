# Infrastructure

This folder contains shared infrastructure components for the Customer Success Tool.

## PostgreSQL Database

PostgreSQL database for storing customer success data including Thena requests, Salesforce sync data, and other persistent information.

### Configuration

- **Database**: `customer_success`
- **User**: `postgres`
- **Password**: `postgres` (default for development)
- **Port**: `5432`
- **Storage**: 5Gi persistent volume

### Connection String

From within the Kubernetes cluster:
```
postgresql://postgres:postgres@postgres:5432/customer_success
```

### Accessing the Database

From a pod in the cluster:
```bash
kubectl exec -it deployment/postgres -n <namespace> -- psql -U postgres -d customer_success
```

### Deployment

PostgreSQL is automatically deployed via the okteto.yml manifest:
```bash
okteto deploy
```

### Customization

To modify PostgreSQL settings, edit:
- `infrastructure/postgres/chart/values.yaml` - Helm chart values
- Update password via environment variables in production

### Next Steps

1. Create database schema/migrations
2. Update backend to connect to PostgreSQL
3. Migrate from in-memory storage to persistent database
