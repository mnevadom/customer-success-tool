import React, { useState } from 'react';
import '../styles/KanbanBoard.css';

const KanbanBoard = ({ dashboardId }) => {
  const [columns, setColumns] = useState({
    open: {
      id: 'open',
      title: 'Open',
      cards: [
        {
          id: 'card-1',
          customerName: 'iCapital Network',
          assignedTo: 'Mario',
          createdAt: '2026-01-10',
          description: 'Setup new environment for production'
        },
        {
          id: 'card-2',
          customerName: 'Upwork',
          assignedTo: 'Jona',
          createdAt: '2026-01-12',
          description: 'Review authentication implementation'
        },
        {
          id: 'card-3',
          customerName: 'Carta',
          assignedTo: 'Mario',
          createdAt: '2026-01-13',
          description: 'Performance optimization needed'
        }
      ]
    },
    inProgress: {
      id: 'inProgress',
      title: 'In Progress',
      cards: [
        {
          id: 'card-4',
          customerName: 'Ramp',
          assignedTo: 'Ramiro',
          createdAt: '2026-01-08',
          description: 'Database migration in progress'
        },
        {
          id: 'card-5',
          customerName: 'SoFi',
          assignedTo: 'Jona',
          createdAt: '2026-01-09',
          description: 'API integration work'
        }
      ]
    },
    waitingForUs: {
      id: 'waitingForUs',
      title: 'Waiting For Us',
      cards: [
        {
          id: 'card-6',
          customerName: 'Webull',
          assignedTo: 'Mario',
          createdAt: '2026-01-07',
          description: 'Needs code review from our team'
        },
        {
          id: 'card-7',
          customerName: 'Robinhood',
          assignedTo: 'Ramiro',
          createdAt: '2026-01-11',
          description: 'Waiting for deployment approval'
        }
      ]
    },
    waitingCustomer: {
      id: 'waitingCustomer',
      title: 'Waiting Customer',
      cards: [
        {
          id: 'card-8',
          customerName: 'Plaid',
          assignedTo: 'Jona',
          createdAt: '2026-01-05',
          description: 'Awaiting customer feedback on feature'
        },
        {
          id: 'card-9',
          customerName: 'Stripe',
          assignedTo: 'Ramiro',
          createdAt: '2026-01-06',
          description: 'Customer testing in progress'
        }
      ]
    },
    done: {
      id: 'done',
      title: 'Done',
      cards: [
        {
          id: 'card-10',
          customerName: 'Mercury',
          assignedTo: 'Mario',
          createdAt: '2026-01-03',
          description: 'SSL certificate renewal completed'
        },
        {
          id: 'card-11',
          customerName: 'Brex',
          assignedTo: 'Jona',
          createdAt: '2026-01-04',
          description: 'Security audit passed'
        }
      ]
    }
  });

  const [draggedCard, setDraggedCard] = useState(null);
  const [draggedFromColumn, setDraggedFromColumn] = useState(null);

  const handleDragStart = (e, card, columnId) => {
    setDraggedCard(card);
    setDraggedFromColumn(columnId);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  };

  const handleDrop = (e, targetColumnId) => {
    e.preventDefault();

    if (!draggedCard || !draggedFromColumn) return;

    // Don't do anything if dropped in the same column
    if (draggedFromColumn === targetColumnId) {
      setDraggedCard(null);
      setDraggedFromColumn(null);
      return;
    }

    // Remove card from source column
    const sourceColumn = { ...columns[draggedFromColumn] };
    sourceColumn.cards = sourceColumn.cards.filter(card => card.id !== draggedCard.id);

    // Add card to target column
    const targetColumn = { ...columns[targetColumnId] };
    targetColumn.cards = [...targetColumn.cards, draggedCard];

    // Update state
    setColumns({
      ...columns,
      [draggedFromColumn]: sourceColumn,
      [targetColumnId]: targetColumn
    });

    setDraggedCard(null);
    setDraggedFromColumn(null);
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

  const getColumnColor = (columnId) => {
    const colors = {
      open: '#7A3FF2',
      inProgress: '#00D1CA',
      waitingForUs: '#FF9900',
      waitingCustomer: '#FF6B6B',
      done: '#4CAF50'
    };
    return colors[columnId] || '#7A3FF2';
  };

  if (!dashboardId) {
    return (
      <div className="content-placeholder">
        <div className="content-placeholder-icon">📋</div>
        <div className="content-placeholder-text">Select a dashboard to view the Kanban board</div>
      </div>
    );
  }

  return (
    <div className="kanban-container">
      <div className="kanban-header">
        <h1 className="kanban-title">Task Board</h1>
        <div className="kanban-stats">
          <span className="kanban-stat">
            Total Tasks: <strong>{Object.values(columns).reduce((sum, col) => sum + col.cards.length, 0)}</strong>
          </span>
        </div>
      </div>

      <div className="kanban-board">
        {Object.values(columns).map((column) => (
          <div
            key={column.id}
            className="kanban-column"
            onDragOver={handleDragOver}
            onDrop={(e) => handleDrop(e, column.id)}
          >
            <div className="kanban-column-header" style={{ borderLeftColor: getColumnColor(column.id) }}>
              <h3 className="kanban-column-title">{column.title}</h3>
              <span className="kanban-column-count">{column.cards.length}</span>
            </div>

            <div className="kanban-cards">
              {column.cards.map((card) => (
                <div
                  key={card.id}
                  className="kanban-card"
                  draggable
                  onDragStart={(e) => handleDragStart(e, card, column.id)}
                >
                  <div className="kanban-card-header">
                    <div className="kanban-card-customer">{card.customerName}</div>
                    <div className="kanban-card-date">{formatDate(card.createdAt)}</div>
                  </div>

                  <div className="kanban-card-description">
                    {card.description}
                  </div>

                  <div className="kanban-card-footer">
                    <div className="kanban-card-assignee">
                      <div className="kanban-card-avatar">
                        {card.assignedTo.charAt(0)}
                      </div>
                      <span>{card.assignedTo}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default KanbanBoard;
