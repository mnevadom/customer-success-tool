import React from 'react';
import ReactDOM from 'react-dom/client';
import { ApolloClient, InMemoryCache, ApolloProvider } from '@apollo/client';
import App from './App.jsx';

// Get backend URL - dynamically construct based on current host
const getBackendURL = () => {
  // Check if we have an environment variable (for local dev)
  if (import.meta.env.VITE_BACKEND_URL) {
    return import.meta.env.VITE_BACKEND_URL;
  }

  // For Okteto deployments, construct URL from current host
  if (typeof window !== 'undefined') {
    const currentHost = window.location.hostname;
    // Replace 'frontend' with 'backend' in the hostname
    if (currentHost.includes('frontend-')) {
      const backendHost = currentHost.replace('frontend-', 'backend-');
      const protocol = window.location.protocol;
      return `${protocol}//${backendHost}`;
    }
  }

  // Fallback for local development
  return 'http://localhost:8080';
};

const BACKEND_URL = getBackendURL();

// Create Apollo Client
const client = new ApolloClient({
  uri: `${BACKEND_URL}/graphql`,
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ApolloProvider client={client}>
      <App />
    </ApolloProvider>
  </React.StrictMode>
);
