import { test, expect, Page } from '@playwright/test';

const APP_BASE_URL = 'http://app.localtest.me';

// Helper function to login
async function login(page: Page) {
  await page.goto(`${APP_BASE_URL}/login`);
  await expect(page.getByRole('heading', { name: 'MeshPump' })).toBeVisible();

  // Step 1: Fill in credentials and submit to initiate OAuth2 flow
  await page.getByPlaceholder('Email').fill('admin@example.com');
  await page.getByPlaceholder('Password').fill('Admin123!');
  await page.getByRole('button', { name: 'Sign in' }).click();

  // Wait for OAuth2 redirect back to login page with login_challenge
  await page.waitForURL(/\/login\?login_challenge=/, { timeout: 15000 });

  // Step 2: Fill credentials again (form is reset after redirect)
  await page.getByPlaceholder('Email').fill('admin@example.com');
  await page.getByPlaceholder('Password').fill('Admin123!');
  await page.getByRole('button', { name: 'Sign in' }).click();

  // Wait for OAuth2 flow to complete - redirects to /workspaces
  await page.waitForURL(/\/workspaces/, { timeout: 15000 });
}

test.describe('RBAC Roles Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should display roles list with system roles', async ({ page }) => {
    // Navigate to roles page
    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles`);

    // Verify page title
    await expect(page.getByRole('heading', { name: 'Roles' })).toBeVisible();

    // Verify Create Role button is visible
    await expect(page.getByRole('button', { name: 'Create Role' })).toBeVisible();

    // Verify system roles are displayed (at least super_admin should exist)
    await expect(page.getByText('super_admin')).toBeVisible();
    await expect(page.getByText('System', { exact: true }).first()).toBeVisible();
  });

  test('should display both global and workspace roles', async ({ page }) => {
    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles`);

    // Wait for roles to load
    await expect(page.getByText('super_admin')).toBeVisible();

    // Verify both global and workspace roles are visible
    await expect(page.getByText('super_admin')).toBeVisible(); // global role
    await expect(page.getByText('workspace_owner')).toBeVisible(); // workspace role
    await expect(page.getByText('workspace_admin')).toBeVisible(); // workspace role
    await expect(page.getByText('workspace_member')).toBeVisible(); // workspace role
  });

  test('should view system role (read-only)', async ({ page }) => {
    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles`);

    // Wait for roles to load
    await expect(page.getByText('super_admin')).toBeVisible();

    // Click view button on super_admin row (icon button with title="View")
    const superAdminRow = page.getByRole('row').filter({ hasText: 'super_admin' });
    await superAdminRow.locator('button[title="View"]').click();

    // Verify we're on the role edit page
    await expect(page.getByRole('heading', { name: 'View Role' })).toBeVisible();

    // Verify system role warning is displayed
    await expect(page.getByText('System Role')).toBeVisible();
    await expect(
      page.getByText('This is a system-defined role and cannot be modified')
    ).toBeVisible();

    // Verify form fields are disabled
    await expect(page.getByPlaceholder('custom_role')).toBeDisabled();
    await expect(page.getByPlaceholder('Custom Editor')).toBeDisabled();
  });

  test('should create a new custom role', async ({ page }) => {
    const roleName = `test_role_${Date.now()}`;

    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles/new`);

    // Verify we're on the create role page
    await expect(page.getByRole('heading', { name: 'Create Role' })).toBeVisible();

    // Fill in role details (use placeholders to avoid ambiguity)
    await page.getByPlaceholder('custom_role').fill(roleName);
    await page.getByPlaceholder('Custom Editor').fill('Test Role for E2E');

    // Wait for permissions table to load (verify it's visible)
    await expect(page.getByRole('table')).toBeVisible();
    await expect(page.getByRole('cell', { name: 'datasource' })).toBeVisible();

    // Submit the form
    await page.getByRole('button', { name: 'Create' }).click();

    // Wait for success message
    await expect(page.getByText('Role created successfully')).toBeVisible({ timeout: 10000 });

    // Verify redirect happened
    await expect(page).toHaveURL(/\/w\/internal\/rbac\/roles$/);

    // Verify the new role appears in the list
    await expect(page.getByText(roleName)).toBeVisible();
  });

  test('should create a role and select permissions', async ({ page }) => {
    const roleName = `perm_role_${Date.now()}`;

    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles/new`);

    // Fill in role details
    await page.getByPlaceholder('custom_role').fill(roleName);
    await page.getByPlaceholder('Custom Editor').fill('Role with Permissions');

    // Wait for permissions table to load
    await expect(page.getByRole('cell', { name: 'datasource' })).toBeVisible();

    // Select datasource:read permission
    const datasourceRow = page.getByRole('row').filter({ hasText: 'datasource' });
    const checkbox = datasourceRow.getByRole('checkbox').first();
    await checkbox.check();

    // Verify checkbox is checked
    await expect(checkbox).toBeChecked();

    // Submit the form
    await page.getByRole('button', { name: 'Create' }).click();

    // Wait for success or error - longer timeout for API with permissions
    await expect(page.getByText('Role created successfully')).toBeVisible({ timeout: 15000 });

    // Verify the role was created
    await expect(page.getByText(roleName)).toBeVisible();
  });

  test('should edit an existing custom role', async ({ page }) => {
    // First create a role to edit
    const roleName = `edit_test_${Date.now()}`;

    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles/new`);

    await page.getByPlaceholder('custom_role').fill(roleName);
    await page.getByPlaceholder('Custom Editor').fill('Role to Edit');
    await page.getByRole('button', { name: 'Create' }).click();

    // Wait for success and redirect
    await expect(page.getByText('Role created successfully')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(roleName)).toBeVisible({ timeout: 5000 });

    // Click edit on the new role (icon button with title="Edit")
    const roleRow = page.getByRole('row').filter({ hasText: roleName });
    await roleRow.locator('button[title="Edit"]').click();

    // Verify we're on the edit page
    await expect(page.getByRole('heading', { name: 'Edit Role' })).toBeVisible({ timeout: 5000 });

    // Wait for form to be populated
    await expect(page.getByPlaceholder('Custom Editor')).toHaveValue('Role to Edit', { timeout: 5000 });

    // Update display name
    await page.getByPlaceholder('Custom Editor').clear();
    await page.getByPlaceholder('Custom Editor').fill('Updated Role Name');

    // Submit the form
    await page.getByRole('button', { name: 'Update' }).click();

    // Verify success message
    await expect(page.getByText('Role updated successfully')).toBeVisible({ timeout: 10000 });

    // Verify redirect happened
    await expect(page).toHaveURL(/\/w\/internal\/rbac\/roles$/);
  });

  test('should delete a custom role', async ({ page }) => {
    // First create a role to delete
    const roleName = `delete_test_${Date.now()}`;

    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles/new`);

    await page.getByPlaceholder('custom_role').fill(roleName);
    await page.getByPlaceholder('Custom Editor').fill('Role to Delete');
    await page.getByRole('button', { name: 'Create' }).click();

    await page.waitForURL(/\/w\/internal\/rbac\/roles$/);
    await expect(page.getByText(roleName)).toBeVisible();

    // Click delete on the new role (icon button with title="Delete")
    const roleRow = page.getByRole('row').filter({ hasText: roleName });
    await roleRow.locator('button[title="Delete"]').click();

    // Confirm deletion in popconfirm
    await page.locator('.ant-popconfirm').getByRole('button', { name: 'Delete' }).click();

    // Verify success message
    await expect(page.getByText('Role deleted')).toBeVisible();

    // Verify role is removed from the list
    await expect(page.getByText(roleName)).not.toBeVisible();
  });

  test('should display permissions grouped by resource', async ({ page }) => {
    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles/new`);

    // Wait for permissions table to load
    await expect(page.getByRole('table')).toBeVisible();

    // Verify resource groups are displayed (as table cells)
    const resources = [
      'datasource',
      'pipeline',
      'storage',
      'contact',
      'message',
      'scheduler',
      'member',
      'workspace',
    ];

    for (const resource of resources) {
      await expect(page.getByRole('cell', { name: resource })).toBeVisible();
    }

    // Verify permission columns
    await expect(page.getByRole('columnheader', { name: 'Resource' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Read' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Write' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Create' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Delete' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Admin' })).toBeVisible();
  });

  test('should show different permissions for global vs workspace scope', async ({ page }) => {
    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles/new`);

    // Default is workspace scope - verify workspace resources
    await expect(page.getByRole('cell', { name: 'datasource' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'pipeline' })).toBeVisible();

    // Switch to global scope using the Select component
    await page.locator('#scope').click();
    await page.locator('.ant-select-item-option').filter({ hasText: 'Global' }).click();

    // Wait for permissions to reload
    await page.waitForTimeout(500);

    // Verify global resources are shown
    await expect(page.getByRole('cell', { name: 'user' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'role' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'rbac' })).toBeVisible();
  });

  test('should validate role name format', async ({ page }) => {
    await page.goto(`${APP_BASE_URL}/w/internal/rbac/roles/new`);

    // Try invalid name (uppercase)
    await page.getByPlaceholder('custom_role').fill('InvalidName');
    await page.getByPlaceholder('Custom Editor').fill('Test');
    await page.getByRole('button', { name: 'Create' }).click();

    // Verify validation error
    await expect(
      page.getByText('Name must start with lowercase letter')
    ).toBeVisible();

    // Fix the name
    await page.getByPlaceholder('custom_role').clear();
    await page.getByPlaceholder('custom_role').fill('valid_name');

    // Error should disappear
    await expect(
      page.getByText('Name must start with lowercase letter')
    ).not.toBeVisible();
  });
});
