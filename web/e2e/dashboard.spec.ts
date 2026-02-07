import { test, expect } from '@playwright/test'
import { login, selectTeam } from './fixtures/auth'

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('displays core UI elements: branding, teams, notifications, stats', async ({ page }) => {
    // Branding
    await expect(page.locator('text=Edictflow').first()).toBeVisible()

    // Teams section
    await expect(page.locator('h3:has-text("Teams")').first()).toBeVisible({ timeout: 5000 })

    // Notification bell in header
    const header = page.locator('header')
    await expect(header.locator('button:has(svg)').first()).toBeVisible({ timeout: 3000 })

    // Stats cards
    await expect(page.locator('text=Pending Approvals').first()).toBeVisible({ timeout: 5000 })
    await expect(page.locator('text=Active Rules').first()).toBeVisible()
  })

  test('team management: view, create, select', async ({ page }) => {
    // View teams
    await expect(page.locator('h3:has-text("Teams")').first()).toBeVisible({ timeout: 5000 })

    // Create team
    const teamName = `E2E Team ${Date.now()}`
    await page.locator('button[title="Create new team"]').click()
    await expect(page.locator('text="Create New Team"')).toBeVisible({ timeout: 3000 })

    await page.locator('input[placeholder="Enter team name"]').fill(teamName)

    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/teams') && (resp.status() === 200 || resp.status() === 201),
      { timeout: 10000 }
    )
    await page.locator('button:has-text("Create Team")').click()
    await responsePromise

    await expect(page.locator(`text="${teamName}"`)).toBeVisible({ timeout: 5000 })

    // Select team
    await selectTeam(page, 'Test Team')
    await expect(page.locator('main')).toBeVisible()
  })

  test('responsive design: mobile, tablet, desktop', async ({ page }) => {
    const viewports = [
      { width: 375, height: 667, name: 'mobile' },
      { width: 768, height: 1024, name: 'tablet' },
      { width: 1440, height: 900, name: 'desktop' },
    ]

    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height })
      await page.waitForTimeout(500)

      await expect(page.locator('body')).toBeVisible()
      const content = await page.locator('body').textContent()
      expect(content?.length).toBeGreaterThan(0)
    }
  })
})
