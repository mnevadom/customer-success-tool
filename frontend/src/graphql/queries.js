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

export const GET_THENA_REQUESTS = gql`
  query GetThenaRequests {
    thenaRequests {
      id
      requestId
      thenaId
      eventId
      status
      subStatus
      subStatusName
      subStatusDesc
      customerName
      crmId
      crmName
      channelId
      channelName
      permalink
      requestLink
      thenaUrl
      assignedToId
      assignedToName
      assignedToEmail
      requestorId
      requestorName
      requestorEmail
      createdAt
      updatedAt
      replyCount
      description
      receivedAt
    }
  }
`;

export const GET_FEATURE_REQUESTS = gql`
  query GetFeatureRequests {
    featureRequests {
      id
      name
      description
      customerName
      customersNumberOfRequests
      jiraLink
      jiraKey
      status
      priority
      createdAt
      updatedAt
      createdBy
      tags
      estimatedEffort
      targetRelease
    }
  }
`;

export const GET_PENDING_FEATURE_REQUESTS = gql`
  query GetPendingFeatureRequests {
    pendingFeatureRequests {
      id
      name
      description
      customerName
      customersNumberOfRequests
      jiraLink
      jiraKey
      status
      priority
      createdAt
      updatedAt
      createdBy
      tags
      estimatedEffort
      targetRelease
    }
  }
`;

export const CREATE_FEATURE_REQUEST = gql`
  mutation CreateFeatureRequest(
    $name: String!
    $description: String!
    $customerName: String!
    $customersNumberOfRequests: Int
    $jiraLink: String
    $jiraKey: String
    $priority: String
    $tags: [String]
    $estimatedEffort: String
    $targetRelease: String
  ) {
    createFeatureRequest(
      name: $name
      description: $description
      customerName: $customerName
      customersNumberOfRequests: $customersNumberOfRequests
      jiraLink: $jiraLink
      jiraKey: $jiraKey
      priority: $priority
      tags: $tags
      estimatedEffort: $estimatedEffort
      targetRelease: $targetRelease
    ) {
      id
      name
      description
      customerName
      customersNumberOfRequests
      jiraLink
      jiraKey
      status
      priority
      createdAt
      updatedAt
    }
  }
`;
