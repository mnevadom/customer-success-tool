import React from 'react';
import '../styles/OverviewView.css';

const OverviewView = () => {
  const stats = {
    totalClients: 28,
    activeClients: 25,
    atRiskClients: 3,
    criticalAlerts: 2,
    highAlerts: 3,
    upcomingRenewals: 5,
    avgSatisfaction: 87,
    totalARR: '$3,250,000'
  };

  const recentActivity = [
    {
      id: 1,
      type: 'alert',
      icon: '🚨',
      message: 'Ramp reached account creation limit',
      time: '2 hours ago',
      severity: 'critical'
    },
    {
      id: 2,
      type: 'renewal',
      icon: '📅',
      message: 'iCapital Network renewal coming up in 1 month',
      time: '5 hours ago',
      severity: 'medium'
    },
    {
      id: 3,
      type: 'success',
      icon: '✅',
      message: 'Mercury SSL certificate renewal completed',
      time: '1 day ago',
      severity: 'success'
    },
    {
      id: 4,
      type: 'alert',
      icon: '⚠️',
      message: 'Monday.com approaching service usage limit (95%)',
      time: '1 day ago',
      severity: 'high'
    },
    {
      id: 5,
      type: 'meeting',
      icon: '👥',
      message: 'Scheduled renewal discussion with Stripe',
      time: '2 days ago',
      severity: 'info'
    }
  ];

  const topClients = [
    { name: 'iCapital Network', arr: '$175,000', status: 'Active', health: 95 },
    { name: 'Stripe', arr: '$150,000', status: 'At Risk', health: 65 },
    { name: 'Monday.com', arr: '$125,000', status: 'Active', health: 85 },
    { name: 'Plaid', arr: '$120,000', status: 'Active', health: 90 },
    { name: 'Robinhood', arr: '$110,000', status: 'Active', health: 88 }
  ];

  const getHealthColor = (health) => {
    if (health >= 90) return '#4CAF50';
    if (health >= 75) return '#00D1CA';
    if (health >= 60) return '#FF9900';
    return '#DC3545';
  };

  const getSeverityColor = (severity) => {
    const colors = {
      critical: '#DC3545',
      high: '#FF9900',
      medium: '#FFC107',
      success: '#4CAF50',
      info: '#17A2B8'
    };
    return colors[severity] || '#6c757d';
  };

  return (
    <div className="overview-container">
      <div className="overview-header">
        <h1 className="overview-title">Overview Dashboard</h1>
        <div className="overview-subtitle">Customer Success at a Glance</div>
      </div>

      {/* Key Metrics */}
      <div className="metrics-grid">
        <div className="metric-card">
          <div className="metric-icon">👥</div>
          <div className="metric-content">
            <div className="metric-value">{stats.totalClients}</div>
            <div className="metric-label">Total Clients</div>
            <div className="metric-detail">
              <span className="metric-success">{stats.activeClients} Active</span>
              <span className="metric-danger">{stats.atRiskClients} At Risk</span>
            </div>
          </div>
        </div>

        <div className="metric-card">
          <div className="metric-icon">💰</div>
          <div className="metric-content">
            <div className="metric-value">{stats.totalARR}</div>
            <div className="metric-label">Total ARR</div>
            <div className="metric-detail">
              <span className="metric-trend-up">↗ +12% this quarter</span>
            </div>
          </div>
        </div>

        <div className="metric-card">
          <div className="metric-icon">⚠️</div>
          <div className="metric-content">
            <div className="metric-value">{stats.criticalAlerts + stats.highAlerts}</div>
            <div className="metric-label">Active Alerts</div>
            <div className="metric-detail">
              <span className="metric-danger">{stats.criticalAlerts} Critical</span>
              <span className="metric-warning">{stats.highAlerts} High</span>
            </div>
          </div>
        </div>

        <div className="metric-card">
          <div className="metric-icon">📅</div>
          <div className="metric-content">
            <div className="metric-value">{stats.upcomingRenewals}</div>
            <div className="metric-label">Upcoming Renewals</div>
            <div className="metric-detail">
              <span className="metric-info">Next 90 days</span>
            </div>
          </div>
        </div>

        <div className="metric-card">
          <div className="metric-icon">😊</div>
          <div className="metric-content">
            <div className="metric-value">{stats.avgSatisfaction}%</div>
            <div className="metric-label">Avg Satisfaction</div>
            <div className="metric-detail">
              <span className="metric-trend-up">↗ +3% from last month</span>
            </div>
          </div>
        </div>
      </div>

      <div className="overview-content">
        {/* Top Clients */}
        <div className="overview-section">
          <h2 className="section-title">Top Clients by ARR</h2>
          <div className="clients-list">
            {topClients.map((client, index) => (
              <div key={index} className="client-row">
                <div className="client-rank">{index + 1}</div>
                <div className="client-info">
                  <div className="client-name">{client.name}</div>
                  <div className="client-arr">{client.arr}</div>
                </div>
                <div className="client-health">
                  <div className="health-bar">
                    <div
                      className="health-fill"
                      style={{
                        width: `${client.health}%`,
                        backgroundColor: getHealthColor(client.health)
                      }}
                    />
                  </div>
                  <div className="health-score" style={{ color: getHealthColor(client.health) }}>
                    {client.health}%
                  </div>
                </div>
                <div className={`client-status ${client.status === 'At Risk' ? 'status-risk' : 'status-active'}`}>
                  {client.status}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Activity */}
        <div className="overview-section">
          <h2 className="section-title">Recent Activity</h2>
          <div className="activity-list">
            {recentActivity.map((activity) => (
              <div key={activity.id} className="activity-item">
                <div
                  className="activity-icon"
                  style={{ backgroundColor: getSeverityColor(activity.severity) + '20' }}
                >
                  {activity.icon}
                </div>
                <div className="activity-content">
                  <div className="activity-message">{activity.message}</div>
                  <div className="activity-time">{activity.time}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="quick-actions">
        <h2 className="section-title">Quick Actions</h2>
        <div className="actions-grid">
          <button className="action-button">
            <span className="action-icon">📊</span>
            <span className="action-label">View All Reports</span>
          </button>
          <button className="action-button">
            <span className="action-icon">📧</span>
            <span className="action-label">Send Client Update</span>
          </button>
          <button className="action-button">
            <span className="action-icon">📅</span>
            <span className="action-label">Schedule Meeting</span>
          </button>
          <button className="action-button">
            <span className="action-icon">🔔</span>
            <span className="action-label">Manage Alerts</span>
          </button>
        </div>
      </div>
    </div>
  );
};

export default OverviewView;
