const express = require('express');
const jsforce = require('jsforce');
const cors = require('cors');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 9000;

// CORS configuration
app.use(cors());
app.use(express.json());

// Salesforce connection instance
let sfConnection = null;

// Salesforce credentials from environment variables
const SF_LOGIN_URL = process.env.SF_LOGIN_URL || 'https://login.salesforce.com';
const SF_USERNAME = process.env.SF_USERNAME;
const SF_PASSWORD = process.env.SF_PASSWORD;
const SF_SECURITY_TOKEN = process.env.SF_SECURITY_TOKEN;

/**
 * Initialize Salesforce connection
 */
async function initializeSalesforceConnection() {
  try {
    if (!SF_USERNAME || !SF_PASSWORD) {
      console.warn('⚠️  Salesforce credentials not configured. Service running in mock mode.');
      return null;
    }

    const conn = new jsforce.Connection({
      loginUrl: SF_LOGIN_URL,
    });

    // Login to Salesforce
    const passwordWithToken = SF_SECURITY_TOKEN
      ? SF_PASSWORD + SF_SECURITY_TOKEN
      : SF_PASSWORD;

    await conn.login(SF_USERNAME, passwordWithToken);
    console.log('✅ Connected to Salesforce successfully');
    console.log(`📊 Organization ID: ${conn.userInfo.organizationId}`);
    console.log(`👤 User ID: ${conn.userInfo.userId}`);

    return conn;
  } catch (error) {
    console.error('❌ Failed to connect to Salesforce:', error.message);
    return null;
  }
}

/**
 * Get Salesforce connection (lazy initialization)
 */
async function getSalesforceConnection() {
  if (!sfConnection) {
    sfConnection = await initializeSalesforceConnection();
  }
  return sfConnection;
}

/**
 * Map Salesforce Account to our Client format
 */
function mapSalesforceAccountToClient(account) {
  return {
    id: account.Id,
    name: account.Name,
    status: account.Customer_Status__c || 'Active', // Custom field
    owner: account.Owner?.Name || 'Unassigned',
    createdAt: account.CreatedDate,
    lastActivity: account.LastActivityDate || account.LastModifiedDate,
    tags: account.Tags__c ? account.Tags__c.split(';') : [], // Custom field (semicolon-separated)
    summary: account.Description || '',
    totalARR: account.Annual_Revenue__c ? `$${parseFloat(account.Annual_Revenue__c).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}` : '$0.00',
    nextRenewalDate: account.Renewal_Date__c || null,
    daysUntilRenewal: account.Days_Until_Renewal__c || 0,
    numberOfUnits: account.Number_of_Units__c || 0,
    currentAccountsCreated: account.Accounts_Created__c || 0,
    currentMAU: account.Monthly_Active_Users__c || 0,
    installType: account.Install_Type__c || 'Unknown',
    region: account.Region__c || 'Unknown',
    saOwner: account.SA_Owner__c || 'Unassigned',
  };
}

// ==================== API ROUTES ====================

/**
 * Health check endpoint
 */
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: 'salesforce-sync',
    time: new Date().toISOString(),
    salesforceConnected: sfConnection !== null,
  });
});

/**
 * Root endpoint
 */
app.get('/', (req, res) => {
  res.json({
    service: 'Salesforce Sync Service',
    version: '1.0.0',
    endpoints: {
      health: '/health',
      clients: '/api/clients',
      client: '/api/clients/:id',
      sync: '/api/sync',
      test: '/api/test-connection',
    },
  });
});

/**
 * Test Salesforce connection
 */
app.get('/api/test-connection', async (req, res) => {
  try {
    const conn = await getSalesforceConnection();

    if (!conn) {
      return res.status(503).json({
        success: false,
        message: 'Salesforce credentials not configured',
        mode: 'mock',
      });
    }

    // Test query
    const result = await conn.query('SELECT Id, Name FROM Account LIMIT 1');

    res.json({
      success: true,
      message: 'Connected to Salesforce successfully',
      organizationId: conn.userInfo.organizationId,
      recordsFound: result.totalSize,
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      message: 'Failed to connect to Salesforce',
      error: error.message,
    });
  }
});

/**
 * Get all clients from Salesforce
 * This endpoint fetches Account records and maps them to our Client format
 */
app.get('/api/clients', async (req, res) => {
  try {
    const conn = await getSalesforceConnection();

    if (!conn) {
      return res.status(503).json({
        success: false,
        message: 'Salesforce not connected. Please configure SF_USERNAME and SF_PASSWORD.',
        mode: 'mock',
        clients: [],
      });
    }

    // Query Salesforce Accounts
    // Adjust the SOQL query based on your Salesforce schema
    const result = await conn.query(`
      SELECT
        Id,
        Name,
        Description,
        CreatedDate,
        LastModifiedDate,
        LastActivityDate,
        Owner.Name,
        Customer_Status__c,
        Tags__c,
        Annual_Revenue__c,
        Renewal_Date__c,
        Days_Until_Renewal__c,
        Number_of_Units__c,
        Accounts_Created__c,
        Monthly_Active_Users__c,
        Install_Type__c,
        Region__c,
        SA_Owner__c
      FROM Account
      WHERE Customer_Status__c IN ('Active', 'At risk')
      ORDER BY Name
    `);

    const clients = result.records.map(mapSalesforceAccountToClient);

    res.json({
      success: true,
      count: clients.length,
      clients: clients,
      fetchedAt: new Date().toISOString(),
    });
  } catch (error) {
    console.error('Error fetching clients from Salesforce:', error);
    res.status(500).json({
      success: false,
      message: 'Failed to fetch clients from Salesforce',
      error: error.message,
    });
  }
});

/**
 * Get a specific client by ID from Salesforce
 */
app.get('/api/clients/:id', async (req, res) => {
  try {
    const conn = await getSalesforceConnection();
    const { id } = req.params;

    if (!conn) {
      return res.status(503).json({
        success: false,
        message: 'Salesforce not connected',
        mode: 'mock',
      });
    }

    const result = await conn.query(`
      SELECT
        Id,
        Name,
        Description,
        CreatedDate,
        LastModifiedDate,
        LastActivityDate,
        Owner.Name,
        Customer_Status__c,
        Tags__c,
        Annual_Revenue__c,
        Renewal_Date__c,
        Days_Until_Renewal__c,
        Number_of_Units__c,
        Accounts_Created__c,
        Monthly_Active_Users__c,
        Install_Type__c,
        Region__c,
        SA_Owner__c
      FROM Account
      WHERE Id = '${id}'
      LIMIT 1
    `);

    if (result.totalSize === 0) {
      return res.status(404).json({
        success: false,
        message: 'Client not found',
      });
    }

    const client = mapSalesforceAccountToClient(result.records[0]);

    res.json({
      success: true,
      client: client,
      fetchedAt: new Date().toISOString(),
    });
  } catch (error) {
    console.error('Error fetching client from Salesforce:', error);
    res.status(500).json({
      success: false,
      message: 'Failed to fetch client from Salesforce',
      error: error.message,
    });
  }
});

/**
 * Force sync/refresh Salesforce connection
 */
app.post('/api/sync', async (req, res) => {
  try {
    // Reset connection
    sfConnection = null;
    const conn = await getSalesforceConnection();

    if (!conn) {
      return res.status(503).json({
        success: false,
        message: 'Failed to reconnect to Salesforce',
      });
    }

    res.json({
      success: true,
      message: 'Salesforce connection refreshed',
      organizationId: conn.userInfo.organizationId,
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      message: 'Failed to refresh Salesforce connection',
      error: error.message,
    });
  }
});

/**
 * Endpoint to setup Streaming API (for future push notifications)
 * This is a placeholder for future implementation
 */
app.post('/api/streaming/subscribe', async (req, res) => {
  res.json({
    success: false,
    message: 'Streaming API not yet implemented. Coming soon!',
    note: 'This will use Salesforce Platform Events / Change Data Capture',
  });
});

// Start server
app.listen(PORT, '0.0.0.0', () => {
  console.log(`🚀 Salesforce Sync Service starting on port ${PORT}`);
  console.log(`📡 Environment: ${process.env.NODE_ENV || 'development'}`);
  console.log(`🔗 Salesforce Login URL: ${SF_LOGIN_URL}`);

  if (SF_USERNAME) {
    console.log(`👤 Salesforce Username: ${SF_USERNAME}`);
    console.log('⏳ Initializing Salesforce connection...');
    initializeSalesforceConnection().then((conn) => {
      if (conn) {
        console.log('✅ Salesforce connection ready');
      } else {
        console.log('⚠️  Running in mock mode - configure credentials to enable Salesforce integration');
      }
    });
  } else {
    console.log('⚠️  No Salesforce credentials configured - service running in mock mode');
    console.log('💡 Set SF_USERNAME, SF_PASSWORD, and SF_SECURITY_TOKEN environment variables');
  }
});

// Graceful shutdown
process.on('SIGINT', () => {
  console.log('\n👋 Shutting down Salesforce Sync Service...');
  process.exit(0);
});

process.on('SIGTERM', () => {
  console.log('\n👋 Shutting down Salesforce Sync Service...');
  process.exit(0);
});
