import React, { useState } from 'react';
import { useQuery, useMutation } from '@apollo/client';
import { GET_FEATURE_REQUESTS, CREATE_FEATURE_REQUEST } from '../graphql/queries';
import '../styles/FeatureRequests.css';

const FeatureRequestsView = () => {
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    customerName: '',
    customersNumberOfRequests: 1,
    jiraLink: '',
    jiraKey: '',
    priority: '',
  });

  const { loading, error, data, refetch } = useQuery(GET_FEATURE_REQUESTS, {
    pollInterval: 30000,
  });

  const [createFeatureRequest, { loading: creating }] = useMutation(CREATE_FEATURE_REQUEST, {
    onCompleted: () => {
      setShowForm(false);
      setFormData({
        name: '',
        description: '',
        customerName: '',
        customersNumberOfRequests: 1,
        jiraLink: '',
        jiraKey: '',
        priority: '',
      });
      refetch();
    },
    onError: (error) => {
      console.error('Error creating feature request:', error);
      alert('Failed to create feature request: ' + error.message);
    },
  });

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: name === 'customersNumberOfRequests' ? parseInt(value) || 1 : value,
    }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    if (!formData.name || !formData.description || !formData.customerName) {
      alert('Please fill in all required fields');
      return;
    }

    const variables = {
      name: formData.name,
      description: formData.description,
      customerName: formData.customerName,
      customersNumberOfRequests: formData.customersNumberOfRequests,
    };

    if (formData.jiraLink) variables.jiraLink = formData.jiraLink;
    if (formData.jiraKey) variables.jiraKey = formData.jiraKey;
    if (formData.priority) variables.priority = formData.priority;

    createFeatureRequest({ variables });
  };

  const formatDate = (dateString) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'pending':
        return 'status-badge status-pending';
      case 'in_progress':
        return 'status-badge status-in-progress';
      case 'completed':
        return 'status-badge status-completed';
      default:
        return 'status-badge';
    }
  };

  if (loading) {
    return (
      <div className="feature-requests-container">
        <div className="loading-container">
          <div className="loading-spinner"></div>
          <div className="loading-text">Loading feature requests...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="feature-requests-container">
        <div className="error-container">
          <div className="error-title">Error</div>
          <div>{error.message}</div>
        </div>
      </div>
    );
  }

  const featureRequests = data?.featureRequests || [];

  return (
    <div className="feature-requests-container">
      <div className="feature-requests-header">
        <h1>Feature Requests</h1>
        <button className="btn-primary" onClick={() => setShowForm(!showForm)}>
          {showForm ? 'Cancel' : '+ New Feature Request'}
        </button>
      </div>

      {showForm && (
        <div className="feature-request-form-card">
          <h2>Create New Feature Request</h2>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="name">
                Name <span className="required">*</span>
              </label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                placeholder="Feature name"
                required
              />
            </div>

            <div className="form-group">
              <label htmlFor="description">
                Description <span className="required">*</span>
              </label>
              <textarea
                id="description"
                name="description"
                value={formData.description}
                onChange={handleInputChange}
                placeholder="Detailed description of the feature request"
                rows="4"
                required
              />
            </div>

            <div className="form-row">
              <div className="form-group">
                <label htmlFor="customerName">
                  Customer Name <span className="required">*</span>
                </label>
                <input
                  type="text"
                  id="customerName"
                  name="customerName"
                  value={formData.customerName}
                  onChange={handleInputChange}
                  placeholder="Customer name"
                  required
                />
              </div>

              <div className="form-group">
                <label htmlFor="customersNumberOfRequests">Number of Requests</label>
                <input
                  type="number"
                  id="customersNumberOfRequests"
                  name="customersNumberOfRequests"
                  value={formData.customersNumberOfRequests}
                  onChange={handleInputChange}
                  min="1"
                />
              </div>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label htmlFor="jiraLink">Jira Link</label>
                <input
                  type="url"
                  id="jiraLink"
                  name="jiraLink"
                  value={formData.jiraLink}
                  onChange={handleInputChange}
                  placeholder="https://jira.example.com/browse/PROJ-123"
                />
              </div>

              <div className="form-group">
                <label htmlFor="jiraKey">Jira Key</label>
                <input
                  type="text"
                  id="jiraKey"
                  name="jiraKey"
                  value={formData.jiraKey}
                  onChange={handleInputChange}
                  placeholder="PROJ-123"
                />
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="priority">Priority</label>
              <select
                id="priority"
                name="priority"
                value={formData.priority}
                onChange={handleInputChange}
              >
                <option value="">Select priority</option>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>

            <div className="form-actions">
              <button
                type="button"
                className="btn-secondary"
                onClick={() => setShowForm(false)}
                disabled={creating}
              >
                Cancel
              </button>
              <button type="submit" className="btn-primary" disabled={creating}>
                {creating ? 'Creating...' : 'Create Feature Request'}
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="feature-requests-list">
        {featureRequests.length === 0 ? (
          <div className="empty-state">
            <p>No feature requests yet. Create one to get started!</p>
          </div>
        ) : (
          <div className="feature-requests-grid">
            {featureRequests.map((request) => (
              <div key={request.id} className="feature-request-card">
                <div className="feature-request-header">
                  <h3>{request.name}</h3>
                  <span className={getStatusBadgeClass(request.status)}>
                    {request.status}
                  </span>
                </div>

                <p className="feature-request-description">{request.description}</p>

                <div className="feature-request-meta">
                  <div className="meta-item">
                    <span className="meta-label">Customer:</span>
                    <span className="meta-value">{request.customerName}</span>
                  </div>
                  <div className="meta-item">
                    <span className="meta-label">Requests:</span>
                    <span className="meta-value">{request.customersNumberOfRequests}</span>
                  </div>
                  {request.priority && (
                    <div className="meta-item">
                      <span className="meta-label">Priority:</span>
                      <span className="meta-value priority-badge priority-{request.priority}">
                        {request.priority}
                      </span>
                    </div>
                  )}
                  {request.jiraLink && (
                    <div className="meta-item">
                      <span className="meta-label">Jira:</span>
                      <a
                        href={request.jiraLink}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="jira-link"
                      >
                        {request.jiraKey || 'View in Jira'} →
                      </a>
                    </div>
                  )}
                  <div className="meta-item">
                    <span className="meta-label">Created:</span>
                    <span className="meta-value">{formatDate(request.createdAt)}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default FeatureRequestsView;
