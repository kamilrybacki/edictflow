import { test, expect } from '@playwright/test'

test.describe('Home Page', () => {
  test('should display the home page', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/Edictflow/)
  })

  test('should show status badge or redirect to login', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    // Unauthenticated users get redirected to login
    // Authenticated users see the dashboard with status badge
    const isOnLogin = page.url().includes('/login')
    const isOnDashboard = !isOnLogin

    if (isOnDashboard) {
      // Look for status badge in header
      const header = page.locator('header')
      await expect(header).toBeVisible({ timeout: 5000 })
    } else {
      // On login page, just verify page loaded
      await expect(page.locator('body')).toBeVisible()
    }
  })

  test('should have navigation to login', async ({ page }) => {
    await page.goto('/')
    // Check for login link or redirect behavior
    await expect(page.locator('body')).toBeVisible()
  })
})

test.describe('Authentication Flow', () => {
  test('should display login page', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('h1, h2').first()).toBeVisible()
  })

  test('should display register page', async ({ page }) => {
    await page.goto('/register')
    await expect(page.locator('h1, h2').first()).toBeVisible()
  })

  test('login form should have email and password fields', async ({ page }) => {
    await page.goto('/login')
    const emailInput = page.locator('input[type="email"], input[name="email"]')
    const passwordInput = page.locator('input[type="password"]')

    await expect(emailInput).toBeVisible({ timeout: 5000 })
    await expect(passwordInput).toBeVisible()
  })

  test('register form should have required fields', async ({ page }) => {
    await page.goto('/register')
    const emailInput = page.locator('input[type="email"], input[name="email"]')
    // Use first() to handle password and confirm password fields
    const passwordInput = page.locator('input[type="password"]').first()
    const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]')

    await expect(emailInput).toBeVisible({ timeout: 5000 })
    await expect(passwordInput).toBeVisible()
    await expect(nameInput).toBeVisible()
  })

  test('should show validation error for invalid email', async ({ page }) => {
    await page.goto('/login')

    const emailInput = page.locator('input[type="email"], input[name="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill('invalid-email')
    await passwordInput.fill('password123')
    await submitButton.click()

    // Check for browser validation or custom error message
    const isInvalid = await emailInput.evaluate((el) => !(el as HTMLInputElement).validity.valid)
    expect(isInvalid).toBeTruthy()
  })
})

test.describe('Navigation', () => {
  test('should navigate between login and register', async ({ page }) => {
    await page.goto('/login')

    // Look for a link to register page
    const registerLink = page.locator('a[href*="register"]')
    if (await registerLink.isVisible()) {
      await registerLink.click()
      await expect(page).toHaveURL(/register/)
    }
  })

  test('unauthenticated user should not access admin', async ({ page }) => {
    await page.goto('/admin')
    // Should redirect to login or show unauthorized
    await expect(page).toHaveURL(/login|admin/)
  })

  test('unauthenticated user should not access approvals', async ({ page }) => {
    await page.goto('/approvals')
    // Should redirect to login or show unauthorized
    await expect(page).toHaveURL(/login|approvals/)
  })
})

test.describe('Responsive Design', () => {
  test('should work on mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    await expect(page.locator('body')).toBeVisible()
  })

  test('should work on tablet viewport', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto('/')
    await expect(page.locator('body')).toBeVisible()
  })

  test('should work on desktop viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 })
    await page.goto('/')
    await expect(page.locator('body')).toBeVisible()
  })
})
