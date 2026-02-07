import { test, expect } from '@playwright/test'
import { login, selectTeam } from './fixtures/auth'

test.describe('Rule Lifecycle', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await selectTeam(page, 'Test Team')
  })

  test('complete rule creation flow: open editor, fill fields, save', async ({ page }) => {
    // Open editor
    await page.locator('button:has-text("New Rule")').click()
    await expect(page.locator('h2:has-text("Create New Rule")')).toBeVisible({ timeout: 3000 })

    // Verify form fields exist
    await expect(page.locator('label:has-text("Name")')).toBeVisible()
    await expect(page.locator('label:has-text("Content")')).toBeVisible()
    await expect(page.locator('label:has-text("Target Layer")')).toBeVisible()

    // Admin should see scope selector
    await expect(page.locator('label:has-text("Scope")')).toBeVisible()

    // Fill required fields
    const ruleName = `E2E Test Rule ${Date.now()}`
    await page.locator('input[placeholder*="TypeScript Best Practices"]').fill(ruleName)
    await page.locator('textarea[placeholder*="Rule Title"]').fill('# Test Rule\n\nCreated by E2E test.')
    await page.locator('input[placeholder*="security, best-practices"]').fill('e2e-test')

    // Save
    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/rules') && (resp.status() === 200 || resp.status() === 201),
      { timeout: 10000 }
    )
    await page.locator('button:has-text("Create Rule")').click()
    await responsePromise

    // Verify saved
    await expect(page.locator('h2:has-text("Create New Rule")')).not.toBeVisible({ timeout: 5000 })
    await expect(page.locator(`text="${ruleName}"`)).toBeVisible({ timeout: 5000 })
  })

  test('rule validation: empty fields prevent submission', async ({ page }) => {
    await page.locator('button:has-text("New Rule")').click()
    await expect(page.locator('h2:has-text("Create New Rule")')).toBeVisible({ timeout: 3000 })

    // Try to submit empty
    await page.locator('button:has-text("Create Rule")').click()
    await page.waitForTimeout(500)

    // Modal should still be visible
    await expect(page.locator('h2:has-text("Create New Rule")')).toBeVisible()
  })

  test('cancel rule creation', async ({ page }) => {
    await page.locator('button:has-text("New Rule")').click()
    await expect(page.locator('h2:has-text("Create New Rule")')).toBeVisible({ timeout: 3000 })

    await page.locator('input[placeholder*="TypeScript Best Practices"]').fill('Rule to cancel')
    await page.locator('button:has-text("Cancel")').click()

    await expect(page.locator('h2:has-text("Create New Rule")')).not.toBeVisible({ timeout: 3000 })
  })

  test('global rule creation (admin only): scope selector, force option', async ({ page }) => {
    await page.locator('button:has-text("New Rule")').click()
    await expect(page.locator('h2:has-text("Create New Rule")')).toBeVisible({ timeout: 3000 })

    // Select global scope
    await page.locator('input[value="global"][name="scope"]').click()

    // Force checkbox should appear
    await expect(page.locator('text=Force on all teams')).toBeVisible({ timeout: 3000 })

    // Fill and save global rule
    const ruleName = `E2E Global Rule ${Date.now()}`
    await page.locator('input[placeholder*="TypeScript Best Practices"]').fill(ruleName)
    await page.locator('textarea[placeholder*="Rule Title"]').fill('# Global Rule\n\nApplies to all teams.')

    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/rules') && (resp.status() === 200 || resp.status() === 201),
      { timeout: 10000 }
    )
    await page.locator('button:has-text("Create Global Rule")').click()
    await responsePromise

    await expect(page.locator('h2:has-text("Create New Rule")')).not.toBeVisible({ timeout: 5000 })
  })
})

test.describe('Rule Approval', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('approvals page: display pending rules, approve/reject actions', async ({ page }) => {
    await page.goto('/approvals')
    await page.waitForLoadState('networkidle')

    // Should see approvals page
    await expect(page.locator('h1:has-text("Pending Approvals")')).toBeVisible({ timeout: 5000 })

    // If pending rules exist, test actions
    const pendingRule = page.locator('button:has-text("pending")').first()
    if (await pendingRule.isVisible({ timeout: 3000 }).catch(() => false)) {
      await pendingRule.click()
      await page.waitForTimeout(500)

      // Should see approve/reject buttons and comment field
      await expect(page.locator('button:has-text("Approve")')).toBeVisible({ timeout: 3000 })
      await expect(page.locator('button:has-text("Reject")')).toBeVisible()
      await expect(page.locator('textarea[placeholder*="comment" i]')).toBeVisible()
    }
  })

  test('approve a pending rule', async ({ page }) => {
    await page.goto('/approvals')
    await page.waitForLoadState('networkidle')

    const pendingRule = page.locator('button:has-text("pending")').first()
    if (await pendingRule.isVisible({ timeout: 3000 }).catch(() => false)) {
      await pendingRule.click()
      await page.waitForTimeout(500)

      await page.locator('textarea[placeholder*="comment" i]').fill('Approved via E2E test')

      const responsePromise = page.waitForResponse(
        (resp) => resp.url().includes('/approve') && (resp.status() === 200 || resp.status() === 201),
        { timeout: 10000 }
      )
      await page.locator('button:has-text("Approve")').click()
      await responsePromise
    }
  })
})
