import React from 'react';
import { useQuery } from '@apollo/client';
import { GET_CLIENT } from '../graphql/queries';

const ClientDetail = ({ clientId }) => {
  const { loading, error, data } = useQuery(GET_CLIENT, {
    variables: { id: clientId },
    skip: !clientId,
  });

  if (!clientId) {
    return (
      <div className="content-placeholder">
        <div className="content-placeholder-icon">📊</div>
        <div className="content-placeholder-text">Select a client to view details</div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <div className="loading-text">Loading client details...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error-container">
        <div className="error-title">Error loading client</div>
        <div>{error.message}</div>
      </div>
    );
  }

  if (!data || !data.client) {
    return (
      <div className="error-container">
        <div className="error-title">Client not found</div>
        <div>The requested client could not be found.</div>
      </div>
    );
  }

  const { client } = data;
  const createdDate = new Date(client.createdAt).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
  const lastActivityDate = new Date(client.lastActivity).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
  const nextRenewalDate = new Date(client.nextRenewalDate).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });

  return (
    <div className="client-detail-card">
      <div className="client-detail-header">
        <div>
          <h1 className="client-detail-title">{client.name}</h1>
          <span className={`status-badge ${client.status.toLowerCase().replace(' ', '-')}`}>
            {client.status}
          </span>
        </div>
      </div>

      <div className="client-detail-grid">
        <div className="client-detail-section">
          <div className="client-detail-label">Total ARR</div>
          <div className="client-detail-value">{client.totalARR}</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Next Renewal Date</div>
          <div className="client-detail-value">{nextRenewalDate}</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Days Until Renewal</div>
          <div className="client-detail-value">{client.daysUntilRenewal} days</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Number of Units</div>
          <div className="client-detail-value">{client.numberOfUnits}</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Current Accounts Created</div>
          <div className="client-detail-value">
            {client.currentAccountsCreated > 0 ? client.currentAccountsCreated : 'N/A'}
          </div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Current MAU</div>
          <div className="client-detail-value">
            {client.currentMAU > 0 ? client.currentMAU.toLocaleString() : 'N/A'}
          </div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Install Type</div>
          <div className="client-detail-value">{client.installType}</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Region</div>
          <div className="client-detail-value">{client.region}</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">SA Owner</div>
          <div className="client-detail-value">{client.saOwner}</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Customer Success Owner</div>
          <div className="client-detail-value">{client.owner}</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Created</div>
          <div className="client-detail-value">{createdDate}</div>
        </div>

        <div className="client-detail-section">
          <div className="client-detail-label">Last Activity</div>
          <div className="client-detail-value">{lastActivityDate}</div>
        </div>
      </div>

      <div className="client-detail-section">
        <div className="client-detail-label">Tags</div>
        <div className="client-tags">
          {client.tags.map((tag, index) => (
            <span key={index} className="client-tag">
              {tag}
            </span>
          ))}
        </div>
      </div>

      <div className="client-detail-section">
        <div className="client-detail-label">Summary</div>
        <div className="client-detail-value">{client.summary}</div>
      </div>
    </div>
  );
};

export default ClientDetail;
