import React from 'react';
import '../styles/WarningsView.css';

const WarningsView = () => {
  const warnings = [
    {
      id: 'warn-1',
      type: 'limit',
      severity: 'high',
      customerName: 'Monday.com',
      title: 'Service Limit Approaching',
      description: 'Customer is at 95% of their service usage limit (9,500 / 10,000 requests)',
      daysUntilCritical: 7,
      assignedTo: 'Mario',
      createdAt: '2026-01-14',
      action: 'Contact customer about upgrading plan'
    },
    {
      id: 'warn-2',
      type: 'renewal',
      severity: 'medium',
      customerName: 'iCapital Network',
      title: 'Contract Renewal Due Soon',
      description: 'Annual contract renewal is coming up in 1 month',
      daysUntilCritical: 30,
      assignedTo: 'Jona',
      createdAt: '2026-01-13',
      action: 'Schedule renewal discussion meeting'
    },
    {
      id: 'warn-3',
      type: 'limit',
      severity: 'critical',
      customerName: 'Ramp',
      title: 'Account Creation Limit Reached',
      description: 'Customer has reached 100% of account creation limit (500/500 accounts)',
      daysUntilCritical: 0,
      assignedTo: 'Ramiro',
      createdAt: '2026-01-14',
      action: 'Immediate action required - Enable additional accounts'
    },
    {
      id: 'warn-4',
      type: 'renewal',
      severity: 'high',
      customerName: 'Stripe',
      title: 'Renewal at Risk',
      description: 'Contract expires in 2 weeks. Customer has reported issues and satisfaction score is low',
      daysUntilCritical: 14,
      assignedTo: 'Mario',
      createdAt: '2026-01-12',
      action: 'Emergency meeting with customer success team'
    },
    {
      id: 'warn-5',
      type: 'performance',
      severity: 'medium',
      customerName: 'Plaid',
      title: 'Unusual API Usage Pattern',
      description: 'API usage has increased by 300% in the last 7 days',
      daysUntilCritical: 5,
      assignedTo: 'Jona',
      createdAt: '2026-01-13',
      action: 'Investigate usage pattern and contact customer'
    },
    {
      id: 'warn-6',
      type: 'limit',
      severity: 'high',
      customerName: 'SoFi',
      title: 'Storage Limit Warning',
      description: 'Data storage is at 88% capacity (440GB / 500GB)',
      daysUntilCritical: 10,
      assignedTo: 'Ramiro',
      createdAt: '2026-01-14',
      action: 'Discuss storage upgrade options'
    },
    {
      id: 'warn-7',
      type: 'renewal',
      severity: 'low',
      customerName: 'Robinhood',
      title: 'Renewal Coming Up',
      description: 'Contract renewal in 2 months',
      daysUntilCritical: 60,
      assignedTo: 'Mario',
      createdAt: '2026-01-10',
      action: 'Schedule check-in call'
    },
    {
      id: 'warn-8',
      type: 'performance',
      severity: 'critical',
      customerName: 'Webull',
      title: 'Service Degradation Detected',
      description: 'Customer experiencing 15% increase in API response times',
      daysUntilCritical: 1,
      assignedTo: 'Jona',
      createdAt: '2026-01-14',
      action: 'Technical investigation required'
    }
  ];

  const getSeverityColor = (severity) => {
    const colors = {
      critical: '#DC3545',
      high: '#FF9900',
      medium: '#FFC107',
      low: '#17A2B8'
    };
    return colors[severity] || '#6c757d';
  };

  const getSeverityIcon = (severity) => {
    const icons = {
      critical: '🚨',
      high: '⚠️',
      medium: '⚡',
      low: 'ℹ️'
    };
    return icons[severity] || '📢';
  };

  const getTypeLabel = (type) => {
    const labels = {
      limit: 'Limit Alert',
      renewal: 'Renewal',
      performance: 'Performance',
      security: 'Security'
    };
    return labels[type] || 'Alert';
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    const today = new Date();
    const diffTime = today - date;
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;

    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const groupedWarnings = {
    critical: warnings.filter(w => w.severity === 'critical'),
    high: warnings.filter(w => w.severity === 'high'),
    medium: warnings.filter(w => w.severity === 'medium'),
    low: warnings.filter(w => w.severity === 'low')
  };

  return (
    <div className="warnings-container">
      <div className="warnings-header">
        <h1 className="warnings-title">Warnings & Alerts</h1>
        <div className="warnings-stats">
          <span className="warnings-stat critical">
            <span className="stat-dot"></span>
            Critical: {groupedWarnings.critical.length}
          </span>
          <span className="warnings-stat high">
            <span className="stat-dot"></span>
            High: {groupedWarnings.high.length}
          </span>
          <span className="warnings-stat medium">
            <span className="stat-dot"></span>
            Medium: {groupedWarnings.medium.length}
          </span>
          <span className="warnings-stat low">
            <span className="stat-dot"></span>
            Low: {groupedWarnings.low.length}
          </span>
        </div>
      </div>

      <div className="warnings-content">
        {Object.entries(groupedWarnings).map(([severity, cards]) => (
          cards.length > 0 && (
            <div key={severity} className="warnings-section">
              <h2 className="warnings-section-title" style={{ color: getSeverityColor(severity) }}>
                {getSeverityIcon(severity)} {severity.charAt(0).toUpperCase() + severity.slice(1)} Priority
              </h2>
              <div className="warnings-grid">
                {cards.map((warning) => (
                  <div
                    key={warning.id}
                    className="warning-card"
                    style={{ borderLeftColor: getSeverityColor(warning.severity) }}
                  >
                    <div className="warning-card-header">
                      <div className="warning-card-type">
                        {getTypeLabel(warning.type)}
                      </div>
                      <div className="warning-card-severity" style={{ backgroundColor: getSeverityColor(warning.severity) }}>
                        {warning.severity.toUpperCase()}
                      </div>
                    </div>

                    <div className="warning-card-customer">
                      {warning.customerName}
                    </div>

                    <div className="warning-card-title">
                      {warning.title}
                    </div>

                    <div className="warning-card-description">
                      {warning.description}
                    </div>

                    {warning.daysUntilCritical !== null && (
                      <div className="warning-card-timeline">
                        {warning.daysUntilCritical === 0 ? (
                          <span className="timeline-urgent">⏰ Action needed now</span>
                        ) : warning.daysUntilCritical === 1 ? (
                          <span className="timeline-soon">⏰ Action needed in 1 day</span>
                        ) : (
                          <span className="timeline-normal">📅 {warning.daysUntilCritical} days remaining</span>
                        )}
                      </div>
                    )}

                    <div className="warning-card-action">
                      <strong>Action:</strong> {warning.action}
                    </div>

                    <div className="warning-card-footer">
                      <div className="warning-card-assignee">
                        <div className="warning-card-avatar">
                          {warning.assignedTo.charAt(0)}
                        </div>
                        <span>{warning.assignedTo}</span>
                      </div>
                      <div className="warning-card-date">
                        {formatDate(warning.createdAt)}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )
        ))}
      </div>
    </div>
  );
};

export default WarningsView;
