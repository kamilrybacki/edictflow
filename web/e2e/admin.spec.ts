import { test, expect } from '@playwright/test'
import { login } from './fixtures/auth'

test.describe('User Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/admin/users')
    await page.waitForLoadState('networkidle')
  })

  test('complete user management flow: list, assign role, deactivate', async ({ page }) => {
    // Users list with table
    await expect(page.locator('h1:has-text("Users")')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('table')).toBeVisible()
    await expect(page.locator('th:has-text("User")')).toBeVisible()
    await expect(page.locator('th:has-text("Status")')).toBeVisible()
    await expect(page.locator('text=/\\d+ users/')).toBeVisible()

    // Status badges
    const activeBadge = page.locator('span:has-text("Active")').first()
    expect(await activeBadge.isVisible({ timeout: 3000 }).catch(() => false)).toBeTruthy()

    // Assign role
    const assignRoleBtn = page.locator('button:has-text("Assign Role")').first()
    await expect(assignRoleBtn).toBeVisible()
    await assignRoleBtn.click()

    await expect(page.locator('h2:has-text("Assign Role")')).toBeVisible({ timeout: 3000 })
    await page.locator('button:has-text("Cancel")').click()
    await expect(page.locator('h2:has-text("Assign Role")')).not.toBeVisible({ timeout: 2000 })

    // Deactivate button visible
    const deactivateBtn = page.locator('button:has-text("Deactivate")').first()
    expect(await deactivateBtn.isVisible({ timeout: 3000 }).catch(() => false)).toBeTruthy()
  })
})

test.describe('Role Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/admin/roles')
    await page.waitForLoadState('networkidle')
  })

  test('complete role management flow: list, create, permissions', async ({ page }) => {
    // Roles list
    await expect(page.locator('h1:has-text("Roles")')).toBeVisible({ timeout: 5000 })
    const roleButtons = page.locator('.space-y-2 button')
    expect(await roleButtons.count()).toBeGreaterThan(0)
    await expect(page.locator('text=/\\d+ permissions/').first()).toBeVisible()

    // Create role modal
    await page.locator('button:has-text("Add Role")').click()
    await expect(page.locator('h2:has-text("Create New Role")')).toBeVisible({ timeout: 3000 })
    await expect(page.locator('label:has-text("Name")')).toBeVisible()
    await expect(page.locator('label:has-text("Description")')).toBeVisible()

    // Create a role
    const roleName = `E2E Role ${Date.now()}`
    await page.locator('input[placeholder="Role name"]').fill(roleName)
    await page.locator('input[placeholder="Role description"]').fill('Test role')

    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/roles') && resp.request().method() === 'POST',
      { timeout: 10000 }
    )
    await page.locator('button:has-text("Create")').last().click()
    await responsePromise

    await expect(page.locator('h2:has-text("Create New Role")')).not.toBeVisible({ timeout: 3000 })
    await expect(page.locator(`text="${roleName}"`)).toBeVisible({ timeout: 5000 })

    // View permissions
    await page.locator('.space-y-2 button').first().click()
    await expect(page.locator('text=Permissions')).toBeVisible({ timeout: 3000 })

    // Categories visible
    const categoryHeaders = page.locator('h3.text-sm.font-medium.uppercase')
    expect(await categoryHeaders.count()).toBeGreaterThan(0)

    // Permission checkboxes
    const permissionCheckbox = page.locator('input[type="checkbox"]').first()
    if (await permissionCheckbox.isVisible({ timeout: 2000 }).catch(() => false)) {
      const wasChecked = await permissionCheckbox.isChecked()

      const toggleResponse = page.waitForResponse(
        (resp) => resp.url().includes('/permissions') && (resp.status() === 200 || resp.status() === 201),
        { timeout: 10000 }
      )
      await permissionCheckbox.click()
      await toggleResponse

      expect(await permissionCheckbox.isChecked()).not.toBe(wasChecked)

      // Toggle back
      await permissionCheckbox.click()
    }
  })
})

test.describe('Audit Log', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/admin/audit')
    await page.waitForLoadState('networkidle')
  })

  test('complete audit log flow: display, filter, paginate', async ({ page }) => {
    // Display
    await expect(page.locator('h1:has-text("Audit Log")')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('table')).toBeVisible()
    await expect(page.locator('th:has-text("Time")')).toBeVisible()
    await expect(page.locator('th:has-text("Action")')).toBeVisible()
    await expect(page.locator('th:has-text("Entity")')).toBeVisible()

    // Filters exist
    const entityTypeSelect = page.locator('select').filter({ has: page.locator('option:has-text("rule")') }).first()
    await expect(entityTypeSelect).toBeVisible()

    const actionSelect = page.locator('select').filter({ has: page.locator('option:has-text("created")') }).first()
    await expect(actionSelect).toBeVisible()

    // Apply filter
    await entityTypeSelect.selectOption('rule')
    await page.waitForTimeout(1000)

    // Reset filter
    await entityTypeSelect.selectOption('')
    await page.waitForTimeout(500)

    // Pagination
    await expect(page.locator('button:has-text("Previous")')).toBeVisible()
    await expect(page.locator('button:has-text("Next")')).toBeVisible()
  })
})
