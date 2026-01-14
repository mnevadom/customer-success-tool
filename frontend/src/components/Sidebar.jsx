import React from 'react';

const Sidebar = ({ items, selectedId, onItemSelect, type }) => {
  if (!items || items.length === 0) {
    return (
      <div className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-title">
            {type === 'clients' ? 'Clients' : 'Dashboards'}
          </div>
        </div>
        <div className="loading-container">
          <div className="loading-text">No items found</div>
        </div>
      </div>
    );
  }

  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <div className="sidebar-title">
          {type === 'clients' ? 'Clients' : 'Dashboards'}
        </div>
      </div>
      <ul className="sidebar-list">
        {items.map((item) => (
          <li
            key={item.id}
            className={`sidebar-item ${selectedId === item.id ? 'active' : ''}`}
            onClick={() => onItemSelect(item.id)}
          >
            <div className="sidebar-item-name">{item.name}</div>
            {type === 'clients' && (
              <div className="sidebar-item-meta">
                <span className={`status-badge ${item.status.toLowerCase().replace(' ', '-')}`}>
                  {item.status}
                </span>
                <span>·</span>
                <span>{item.owner}</span>
              </div>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default Sidebar;
