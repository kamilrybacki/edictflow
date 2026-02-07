import { Page } from '@playwright/test'

// Test users from seed data
export const TEST_ADMIN = {
  email: 'alex.rivera@test.local',
  password: 'Test1234',
  name: 'Alex Rivera',
}

export async function login(page: Page, email: string = TEST_ADMIN.email, password: string = TEST_ADMIN.password) {
  await page.goto('/login')
  await page.waitForLoadState('networkidle')

  await page.locator('input[type="email"], input[name="email"]').fill(email)
  await page.locator('input[type="password"]').fill(password)
  await page.locator('button[type="submit"]').click()

  await page.waitForURL('/', { timeout: 10000 })
}

export async function clearAuth(page: Page) {
  await page.context().clearCookies()
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
}

export async function selectTeam(page: Page, teamName: string) {
  const team = page.locator(`li:has-text("${teamName}"), button:has-text("${teamName}")`).first()
  if (await team.isVisible({ timeout: 3000 }).catch(() => false)) {
    await team.click()
    await page.waitForTimeout(500)
  }
}
