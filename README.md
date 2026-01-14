# Customer Success Application

A modern Customer Success / client-tracking web application built with React (frontend) and Go (backend), designed to run on Okteto.

## Architecture

This application consists of two services:

- **Frontend**: React-based UI built with Vite, Apollo Client for GraphQL queries, and Okteto brand colors
- **Backend**: Go-based GraphQL API with mock data, CORS-enabled for development

## Features

### Information Tab
- View all clients in a sidebar list
- Click to see detailed client information including:
  - Client status (Active / At risk)
  - Customer Success owner
  - Creation and last activity dates
  - Tags and summary

### Dashboards Tab
- View available dashboards
- Display widgets with KPIs, charts, and text content
- Real-time data visualization

## Tech Stack

### Frontend
- **Framework**: React 18 with Vite
- **GraphQL Client**: Apollo Client
- **Styling**: CSS with Okteto brand theme tokens
- **Container**: Node 18 Alpine

### Backend
- **Language**: Go 1.21
- **GraphQL**: graphql-go library
- **CORS**: rs/cors middleware
- **Container**: Go Alpine with multi-stage builds

## Quick Start with Okteto

### Prerequisites
- Okteto CLI installed
- Access to an Okteto namespace

### Deploy to Okteto

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd <repository-name>
   ```

2. **Deploy the application**
   ```bash
   okteto deploy --wait
   ```

   This command will:
   - Build both frontend and backend Docker images
   - Deploy both services to your Okteto namespace
   - Create an ingress for the frontend
   - Wait for all pods to be healthy

3. **Get the public URL**
   ```bash
   okteto endpoints
   ```

   Open the frontend URL in your browser to access the application.

### Development Mode with Okteto

For live development with hot-reload:

1. **Start development mode**
   ```bash
   okteto up
   ```

   Choose which service to develop (backend or frontend).

2. **Frontend development**
   ```bash
   okteto up frontend
   ```

   - Your local `./frontend` directory syncs to the container
   - Vite dev server runs with hot module replacement
   - Access at `http://localhost:3000`

3. **Backend development**
   ```bash
   okteto up backend
   ```

   - Your local `./backend` directory syncs to the container
   - Run `go run main.go` in the remote terminal
   - Access GraphQL at `http://localhost:8080/graphql`

## Local Development (without Okteto)

### Backend

```bash
cd backend
go mod download
go run main.go
```

The backend will start on `http://localhost:8080`:
- GraphQL endpoint: `/graphql`
- Health check: `/health`
- GraphiQL IDE: Navigate to `/graphql` in browser

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:3000`.

Create a `.env` file in the frontend directory:
```
VITE_BACKEND_URL=http://localhost:8080
```

## API Endpoints

### Backend GraphQL Queries

**Get all clients:**
```graphql
query {
  clients {
    id
    name
    status
    owner
  }
}
```

**Get single client:**
```graphql
query {
  client(id: "client-1") {
    id
    name
    status
    owner
    createdAt
    lastActivity
    tags
    summary
  }
}
```

**Get all dashboards:**
```graphql
query {
  dashboards {
    id
    name
  }
}
```

**Get single dashboard:**
```graphql
query {
  dashboard(id: "dashboard-1") {
    id
    name
    widgets {
      id
      title
      type
      data
    }
  }
}
```

## Mock Data

The backend includes mock data for demonstration:

### Clients
- **Acme Corp**: Active enterprise client with onboarding tag
- **Beta Ltd**: At-risk trial client needing follow-up
- **Gamma Industries**: Active enterprise client interested in expansion

### Dashboards
- **Customer Health Overview**: KPIs for active clients, at-risk clients, revenue, and activity
- **Sales Pipeline**: Pipeline value and conversion rate metrics

## Project Structure

```
.
├── backend/
│   ├── main.go              # GraphQL server with mock data
│   ├── go.mod               # Go dependencies
│   ├── Dockerfile           # Multi-stage Docker build
│   └── .gitignore
├── frontend/
│   ├── src/
│   │   ├── components/      # React components
│   │   │   ├── TopNav.jsx
│   │   │   ├── Sidebar.jsx
│   │   │   ├── ClientDetail.jsx
│   │   │   └── DashboardView.jsx
│   │   ├── graphql/
│   │   │   └── queries.js   # GraphQL queries
│   │   ├── styles/
│   │   │   ├── theme.css    # Okteto brand colors
│   │   │   └── App.css      # Component styles
│   │   ├── App.jsx          # Main app component
│   │   └── main.jsx         # Apollo Client setup
│   ├── package.json
│   ├── vite.config.js
│   ├── Dockerfile           # Multi-stage Docker build
│   └── .gitignore
├── okteto.yml               # Okteto manifest
└── README.md
```

## Okteto Manifest Details

The `okteto.yml` defines:

### Build Configuration
- **backend**: Builds Go service from `./backend` directory
- **frontend**: Builds React service from `./frontend` directory (development target)

### Deploy Configuration
- Creates Kubernetes Deployments for both services
- Sets up Services for internal communication
- Creates an Ingress for external frontend access
- Configures health checks and resource limits

### Dev Configuration
- **backend**: Syncs code, forwards port 8080, provides bash shell
- **frontend**: Syncs code, forwards port 3000, runs npm dev server

## Environment Variables

### Backend
- `PORT`: Server port (default: 8080)

### Frontend
- `VITE_BACKEND_URL`: Backend GraphQL endpoint URL

## Testing

### Backend Health Check
```bash
curl http://backend:8080/health
```

### GraphQL API Test
```bash
curl -X POST http://backend:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ clients { id name status } }"}'
```

## Troubleshooting

### Validate Okteto Manifest
```bash
okteto validate
```

### Check Pod Status
```bash
kubectl get pods -n ${OKTETO_NAMESPACE}
```

### View Logs
```bash
# Backend logs
kubectl logs -n ${OKTETO_NAMESPACE} deployment/backend

# Frontend logs
kubectl logs -n ${OKTETO_NAMESPACE} deployment/frontend
```

### Destroy and Redeploy
```bash
okteto destroy
okteto deploy --wait
```

## Future Enhancements

This is iteration 1 with mock data. Future iterations could include:

- Real database integration (PostgreSQL, MongoDB)
- Authentication and authorization
- GraphQL mutations for creating/updating clients
- Pagination and filtering
- Real-time updates with subscriptions
- Advanced dashboard widgets and charts
- Export functionality
- Mobile responsive design improvements
- Unit and integration tests

## Support

For Okteto-specific questions, visit:
- [Okteto Documentation](https://www.okteto.com/docs/)
- [Okteto CLI Reference](https://www.okteto.com/docs/reference/okteto-cli)
- [Okteto Community](https://community.okteto.com/)

---

**Built with ❤️ using Okteto AI Powered by Claude Code**
