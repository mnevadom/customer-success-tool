import React, { useState, useMemo } from 'react';
import { useQuery } from '@apollo/client';
import { GET_THENA_REQUESTS } from '../graphql/queries';
import '../styles/ThenaBoardView.css';

const ThenaBoardView = () => {
  const [selectedStatus, setSelectedStatus] = useState('ALL');

  const { loading, error, data } = useQuery(GET_THENA_REQUESTS, {
    pollInterval: 30000, // Poll every 30 seconds for new requests
  });

  const requests = data?.thenaRequests || [];

  // Get unique statuses and subStatusNames
  const statuses = useMemo(() => {
    const uniqueStatuses = [...new Set(requests.map(r => r.status).filter(Boolean))];
    return uniqueStatuses.sort();
  }, [requests]);

  const subStatusNames = useMemo(() => {
    const unique = [...new Set(requests.map(r => r.subStatusName).filter(Boolean))];
    return unique.sort();
  }, [requests]);

  // Filter requests by selected status
  const filteredRequests = useMemo(() => {
    if (selectedStatus === 'ALL') {
      return requests;
    }
    return requests.filter(r => r.status === selectedStatus);
  }, [requests, selectedStatus]);

  // Group filtered requests by subStatusName
  const groupedBySubStatus = useMemo(() => {
    const groups = {};
    filteredRequests.forEach(request => {
      const subStatus = request.subStatusName || 'Unassigned';
      if (!groups[subStatus]) {
        groups[subStatus] = [];
      }
      groups[subStatus].push(request);
    });
    return groups;
  }, [filteredRequests]);

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

  const getStatusCount = (status) => {
    if (status === 'ALL') return requests.length;
    return requests.filter(r => r.status === status).length;
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

  const renderCard = (request) => (
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
  );

  return (
    <div className="thena-board-container">
      <div className="thena-board-header">
        <h1>Thena Board</h1>
        <p className="subtitle">
          {filteredRequests.length} {filteredRequests.length === 1 ? 'request' : 'requests'}
          {selectedStatus !== 'ALL' && ` with status ${selectedStatus}`}
        </p>
      </div>

      {requests.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📬</div>
          <h3>No requests yet</h3>
          <p>Waiting for Thena webhooks with customer information...</p>
        </div>
      ) : (
        <div className="thena-board-layout">
          {/* Left Sidebar - Status Filter */}
          <div className="thena-sidebar">
            <h3 className="sidebar-title">Status Filter</h3>
            <div className="status-filters">
              <button
                className={`status-filter-btn ${selectedStatus === 'ALL' ? 'active' : ''}`}
                onClick={() => setSelectedStatus('ALL')}
              >
                <span className="filter-label">ALL</span>
                <span className="filter-count">{getStatusCount('ALL')}</span>
              </button>
              {statuses.map(status => (
                <button
                  key={status}
                  className={`status-filter-btn ${selectedStatus === status ? 'active' : ''}`}
                  onClick={() => setSelectedStatus(status)}
                  style={{
                    borderLeftColor: getStatusColor(status),
                    borderLeftWidth: '4px',
                    borderLeftStyle: 'solid'
                  }}
                >
                  <span className="filter-label">{status}</span>
                  <span className="filter-count">{getStatusCount(status)}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Main Content - Horizontal Flow by SubStatus */}
          <div className="thena-main-content">
            {filteredRequests.length === 0 ? (
              <div className="empty-state-inline">
                <p>No requests with status {selectedStatus}</p>
              </div>
            ) : (
              <div className="thena-pipeline">
                {Object.keys(groupedBySubStatus).sort().map(subStatus => (
                  <div key={subStatus} className="pipeline-column">
                    <div className="column-header">
                      <h3 className="column-title">{subStatus}</h3>
                      <span className="column-count">{groupedBySubStatus[subStatus].length}</span>
                    </div>
                    <div className="column-cards">
                      {groupedBySubStatus[subStatus].map(request => renderCard(request))}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default ThenaBoardView;
