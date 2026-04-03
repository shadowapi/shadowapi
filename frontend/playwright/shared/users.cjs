/**
 * Central test user registry for ShadowAPI Playwright tests
 *
 * Auth model: simple DB auth (email + password), no OIDC
 * Single instance serves one CRM (Reactima CRM)
 */

const ROLES = {
  ADMIN: 'admin',
}

const SHARED_PASSWORD = 'Admin123#'

const BASE_URL = 'https://shadowapi.local'

const USERS = {
  admin: {
    email: 'admin@shadowapi.local',
    password: SHARED_PASSWORD,
    role: ROLES.ADMIN,
  },
}

const getAdmin = () => USERS.admin

module.exports = { ROLES, SHARED_PASSWORD, BASE_URL, USERS, getAdmin }
