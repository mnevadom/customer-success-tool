import { gql } from '@apollo/client';

export const GET_CLIENTS = gql`
  query GetClients {
    clients {
      id
      name
      status
      owner
    }
  }
`;

export const GET_CLIENT = gql`
  query GetClient($id: String!) {
    client(id: $id) {
      id
      name
      status
      owner
      createdAt
      lastActivity
      tags
      summary
      totalARR
      nextRenewalDate
      daysUntilRenewal
      numberOfUnits
      currentAccountsCreated
      currentMAU
      installType
      region
      saOwner
    }
  }
`;

export const GET_DASHBOARDS = gql`
  query GetDashboards {
    dashboards {
      id
      name
    }
  }
`;

export const GET_DASHBOARD = gql`
  query GetDashboard($id: String!) {
    dashboard(id: $id) {
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
`;
