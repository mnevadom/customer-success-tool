import React, { useState } from 'react';
import { useQuery } from '@apollo/client';
import TopNav from './components/TopNav';
import Sidebar from './components/Sidebar';
import ClientDetail from './components/ClientDetail';
import DashboardView from './components/DashboardView';
import { GET_CLIENTS, GET_DASHBOARDS } from './graphql/queries';
import './styles/theme.css';
import './styles/App.css';

function App() {
  const [activeTab, setActiveTab] = useState('information');
  const [selectedClientId, setSelectedClientId] = useState(null);
  const [selectedDashboardId, setSelectedDashboardId] = useState(null);

  const { loading: clientsLoading, error: clientsError, data: clientsData } = useQuery(
    GET_CLIENTS,
    {
      skip: activeTab !== 'information',
    }
  );

  const { loading: dashboardsLoading, error: dashboardsError, data: dashboardsData } = useQuery(
    GET_DASHBOARDS,
    {
      skip: activeTab !== 'dashboards',
    }
  );

  const handleTabChange = (tab) => {
    setActiveTab(tab);
    if (tab === 'information') {
      setSelectedDashboardId(null);
    } else {
      setSelectedClientId(null);
    }
  };

  const handleClientSelect = (clientId) => {
    setSelectedClientId(clientId);
  };

  const handleDashboardSelect = (dashboardId) => {
    setSelectedDashboardId(dashboardId);
  };

  const renderSidebar = () => {
    if (activeTab === 'information') {
      if (clientsLoading) {
        return (
          <div className="sidebar">
            <div className="sidebar-header">
              <div className="sidebar-title">Clients</div>
            </div>
            <div className="loading-container">
              <div className="loading-spinner"></div>
              <div className="loading-text">Loading clients...</div>
            </div>
          </div>
        );
      }

      if (clientsError) {
        return (
          <div className="sidebar">
            <div className="sidebar-header">
              <div className="sidebar-title">Clients</div>
            </div>
            <div className="error-container">
              <div className="error-title">Error</div>
              <div>{clientsError.message}</div>
            </div>
          </div>
        );
      }

      return (
        <Sidebar
          items={clientsData?.clients || []}
          selectedId={selectedClientId}
          onItemSelect={handleClientSelect}
          type="clients"
        />
      );
    } else {
      if (dashboardsLoading) {
        return (
          <div className="sidebar">
            <div className="sidebar-header">
              <div className="sidebar-title">Dashboards</div>
            </div>
            <div className="loading-container">
              <div className="loading-spinner"></div>
              <div className="loading-text">Loading dashboards...</div>
            </div>
          </div>
        );
      }

      if (dashboardsError) {
        return (
          <div className="sidebar">
            <div className="sidebar-header">
              <div className="sidebar-title">Dashboards</div>
            </div>
            <div className="error-container">
              <div className="error-title">Error</div>
              <div>{dashboardsError.message}</div>
            </div>
          </div>
        );
      }

      return (
        <Sidebar
          items={dashboardsData?.dashboards || []}
          selectedId={selectedDashboardId}
          onItemSelect={handleDashboardSelect}
          type="dashboards"
        />
      );
    }
  };

  const renderContent = () => {
    if (activeTab === 'information') {
      return <ClientDetail clientId={selectedClientId} />;
    } else {
      return <DashboardView dashboardId={selectedDashboardId} />;
    }
  };

  return (
    <div className="app-container">
      <TopNav activeTab={activeTab} onTabChange={handleTabChange} />
      <div className="main-layout">
        {renderSidebar()}
        <div className="content-area">{renderContent()}</div>
      </div>
    </div>
  );
}

export default App;
