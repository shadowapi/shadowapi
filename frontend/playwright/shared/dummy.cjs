/**
 * Dummy data for CRUD Playwright tests
 * Each entity has create data and updated data for testing edit flows
 */

const DUMMY = {
  user: {
    create: {
      email: 'testuser@shadowapi.local',
      password: 'Test1234#',
      first_name: 'Test',
      last_name: 'User',
    },
    update: {
      first_name: 'Updated',
      last_name: 'Person',
    },
  },

  storage_hostfiles: {
    create: {
      name: 'Test Storage',
      path: '/tmp/test-storage',
    },
    update: {
      name: 'Updated Storage',
      path: '/tmp/updated-storage',
    },
  },

  oauth2_credential: {
    create: {
      name: 'Test OAuth2',
      provider: 'GMAIL',
      client_id: 'test-client-id-123',
      secret: 'test-secret-456',
    },
    update: {
      name: 'Updated OAuth2',
    },
  },

  scheduler: {
    create: {
      schedule_type: 'cron',
      cron_expression: '0 */6 * * *',
      timezone: 'Asia/Tokyo',
    },
    update: {
      timezone: 'Europe/London',
    },
  },

  sync_policy: {
    create: {
      name: 'Test Sync Policy',
    },
    update: {
      name: 'Updated Sync Policy',
    },
  },
}

module.exports = { DUMMY }
