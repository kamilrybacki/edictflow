import { test, expect } from '@playwright/test'
import { login, clearAuth, TEST_ADMIN } from './fixtures/auth'

test.describe('Authentication', () => {
  test('complete login flow: display form -> validate -> authenticate -> redirect', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    // Form should be visible with required fields
    await expect(page.locator('input[type="email"], input[name="email"]')).toBeVisible()
    await expect(page.locator('input[type="password"]')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()

    // Invalid credentials should show error
    await page.locator('input[type="email"], input[name="email"]').fill('wrong@email.com')
    await page.locator('input[type="password"]').fill('wrongpassword')
    await page.locator('button[type="submit"]').click()
    await page.waitForTimeout(2000)

    const hasError = await page.locator('text=/invalid|error|failed/i').isVisible().catch(() => false)
    const stillOnLogin = page.url().includes('/login')
    expect(hasError || stillOnLogin).toBeTruthy()

    // Valid credentials should redirect to dashboard
    await page.locator('input[type="email"], input[name="email"]').fill(TEST_ADMIN.email)
    await page.locator('input[type="password"]').fill(TEST_ADMIN.password)
    await page.locator('button[type="submit"]').click()
    await page.waitForURL('/', { timeout: 10000 })

    await expect(page.locator('h1').first()).toBeVisible()
  })

  test('logout flow', async ({ page }) => {
    await login(page)

    const logoutButton = page.locator('button:has-text("Logout")').first()
    await expect(logoutButton).toBeVisible({ timeout: 5000 })
    await logoutButton.click()

    await expect(page).toHaveURL(/login/, { timeout: 5000 })
  })

  test('session expiry redirects to login', async ({ page }) => {
    await login(page)
    await clearAuth(page)

    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)

    await expect(page).toHaveURL(/login/, { timeout: 5000 })
  })
})

test.describe('Access Control', () => {
  test('unauthenticated users are redirected from protected routes', async ({ page }) => {
    const protectedRoutes = ['/', '/approvals', '/admin', '/admin/users', '/admin/roles', '/admin/audit']

    for (const route of protectedRoutes) {
      await page.goto(route)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const isOnLogin = page.url().includes('/login')
      const hasSignInMessage = await page.locator('text=/sign in|Please sign in/i').isVisible().catch(() => false)

      expect(isOnLogin || hasSignInMessage).toBeTruthy()
    }
  })

  test('authenticated admin can access all admin routes', async ({ page }) => {
    await login(page)

    const adminRoutes = ['/admin', '/admin/users', '/admin/roles', '/admin/audit']

    for (const route of adminRoutes) {
      await page.goto(route)
      await page.waitForLoadState('networkidle')

      const isOnAdmin = page.url().includes('/admin')
      expect(isOnAdmin).toBeTruthy()
    }
  })
})
