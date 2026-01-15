import React from 'react';
import { useQuery } from '@apollo/client';
import { GET_THENA_REQUESTS } from '../graphql/queries';
import '../styles/ThenaBoardView.css';

const ThenaBoardView = () => {
  const { loading, error, data } = useQuery(GET_THENA_REQUESTS, {
    pollInterval: 10000, // Poll every 10 seconds for new requests
  });

  if (loading) {
    return (
      <div className="thena-board-container">
        <div className="thena-board-header">
          <h1>Thena Board</h1>
          <p className="subtitle">Live customer requests from Thena</p>
        </div>
        <div className="loading-container">
          <div className="loading-spinner"></div>
          <div className="loading-text">Loading requests...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="thena-board-container">
        <div className="thena-board-header">
          <h1>Thena Board</h1>
          <p className="subtitle">Live customer requests from Thena</p>
        </div>
        <div className="error-container">
          <div className="error-title">Error loading requests</div>
          <div>{error.message}</div>
        </div>
      </div>
    );
  }

  const requests = data?.thenaRequests || [];

  const getStatusColor = (status) => {
    const statusColors = {
      'open': '#3b82f6',
      'OPEN': '#3b82f6',
      'in_progress': '#f59e0b',
      'IN_PROGRESS': '#f59e0b',
      'waiting': '#8b5cf6',
      'WAITING': '#8b5cf6',
      'ONHOLD': '#8b5cf6',
      'resolved': '#10b981',
      'RESOLVED': '#10b981',
      'closed': '#6b7280',
      'CLOSED': '#6b7280',
    };
    return statusColors[status] || '#6b7280';
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 60) {
      return `${diffMins} min${diffMins !== 1 ? 's' : ''} ago`;
    } else if (diffHours < 24) {
      return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
    } else if (diffDays < 7) {
      return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  return (
    <div className="thena-board-container">
      <div className="thena-board-header">
        <h1>Thena Board</h1>
        <p className="subtitle">
          {requests.length} {requests.length === 1 ? 'request' : 'requests'} from customers
        </p>
      </div>

      {requests.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📬</div>
          <h3>No requests yet</h3>
          <p>Waiting for Thena webhooks with customer information...</p>
        </div>
      ) : (
        <div className="thena-cards-grid">
          {requests.map((request) => (
            <div key={request.id} className="thena-card">
              <div className="thena-card-header">
                <div className="thena-card-title-section">
                  <div className="thena-request-id">#{request.requestId}</div>
                  <div
                    className="thena-status-badge"
                    style={{ backgroundColor: getStatusColor(request.status) }}
                  >
                    {request.status}
                  </div>
                </div>
                {request.thenaUrl && (
                  <a
                    href={request.thenaUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="thena-link-button"
                    title="Open in Thena"
                  >
                    🔗
                  </a>
                )}
              </div>

              <div className="thena-card-customer">
                <div className="customer-icon">🏢</div>
                <div className="customer-name">{request.customerName}</div>
              </div>

              {request.subStatusName && (
                <div className="thena-substatus">
                  <span className="substatus-label">Status:</span>
                  <span className="substatus-value">{request.subStatusName}</span>
                </div>
              )}

              {request.description && (
                <div className="thena-description">
                  {request.description}
                </div>
              )}

              <div className="thena-card-meta">
                {request.assignedToName && (
                  <div className="meta-item">
                    <span className="meta-icon">👤</span>
                    <span className="meta-label">Assigned:</span>
                    <span className="meta-value">{request.assignedToName}</span>
                  </div>
                )}
                {request.channelName && (
                  <div className="meta-item">
                    <span className="meta-icon">💬</span>
                    <span className="meta-label">Channel:</span>
                    <span className="meta-value">{request.channelName}</span>
                  </div>
                )}
                {request.replyCount > 0 && (
                  <div className="meta-item">
                    <span className="meta-icon">💭</span>
                    <span className="meta-value">{request.replyCount} {request.replyCount === 1 ? 'reply' : 'replies'}</span>
                  </div>
                )}
              </div>

              <div className="thena-card-footer">
                {request.requestorName && (
                  <div className="requestor-info">
                    From: {request.requestorName}
                    {request.requestorEmail && (
                      <span className="requestor-email"> ({request.requestorEmail})</span>
                    )}
                  </div>
                )}
                <div className="thena-timestamp">
                  {formatDate(request.updatedAt || request.createdAt || request.receivedAt)}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ThenaBoardView;
