import React from 'react';
import { useQuery } from '@apollo/client';
import { GET_DASHBOARD } from '../graphql/queries';
import KanbanBoard from './KanbanBoard';

const DashboardView = ({ dashboardId }) => {
  const { loading, error, data } = useQuery(GET_DASHBOARD, {
    variables: { id: dashboardId },
    skip: !dashboardId,
  });

  if (!dashboardId) {
    return (
      <div className="content-placeholder">
        <div className="content-placeholder-icon">📈</div>
        <div className="content-placeholder-text">Select a dashboard to view</div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <div className="loading-text">Loading dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error-container">
        <div className="error-title">Error loading dashboard</div>
        <div>{error.message}</div>
      </div>
    );
  }

  if (!data || !data.dashboard) {
    return (
      <div className="error-container">
        <div className="error-title">Dashboard not found</div>
        <div>The requested dashboard could not be found.</div>
      </div>
    );
  }

  const { dashboard } = data;

  // If dashboard name contains "Tasks" or "Kanban", show Kanban board
  if (dashboard.name.toLowerCase().includes('tasks') ||
      dashboard.name.toLowerCase().includes('kanban')) {
    return <KanbanBoard dashboardId={dashboardId} />;
  }

  const renderWidgetContent = (widget) => {
    switch (widget.type) {
      case 'KPI':
        return (
          <>
            <div className="widget-kpi-value">{widget.data.value}</div>
            {widget.data.trend && (
              <div className="widget-kpi-trend">↗ {widget.data.trend}</div>
            )}
          </>
        );
      case 'chart':
        return (
          <div className="widget-content">
            <div className="widget-kpi-value">{widget.data.value}</div>
            {widget.data.chartType && (
              <div className="widget-text-content">
                Chart type: {widget.data.chartType}
              </div>
            )}
          </div>
        );
      case 'text':
        return (
          <div className="widget-text-content">{widget.data.content}</div>
        );
      default:
        return (
          <div className="widget-content">
            {JSON.stringify(widget.data, null, 2)}
          </div>
        );
    }
  };

  return (
    <div className="dashboard-container">
      <h1 className="dashboard-title">{dashboard.name}</h1>
      <div className="widgets-grid">
        {dashboard.widgets.map((widget) => (
          <div key={widget.id} className="widget-card">
            <div className="widget-title">{widget.title}</div>
            <div className="widget-type">{widget.type}</div>
            {renderWidgetContent(widget)}
          </div>
        ))}
      </div>
    </div>
  );
};

export default DashboardView;
