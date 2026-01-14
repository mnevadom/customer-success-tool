import React from 'react';

const TopNav = ({ activeTab, onTabChange }) => {
  return (
    <div className="top-nav">
      <div className="top-nav-title">Customer Success</div>
      <div className="top-nav-tabs">
        <button
          className={`top-nav-tab ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => onTabChange('overview')}
        >
          OVERVIEW
        </button>
        <button
          className={`top-nav-tab ${activeTab === 'information' ? 'active' : ''}`}
          onClick={() => onTabChange('information')}
        >
          CUSTOMERS
        </button>
        <button
          className={`top-nav-tab ${activeTab === 'dashboards' ? 'active' : ''}`}
          onClick={() => onTabChange('dashboards')}
        >
          DASHBOARDS
        </button>
        <button
          className={`top-nav-tab ${activeTab === 'warnings' ? 'active' : ''}`}
          onClick={() => onTabChange('warnings')}
        >
          ⚠️ WARNINGS
        </button>
      </div>
    </div>
  );
};

export default TopNav;
