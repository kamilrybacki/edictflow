import { renderHook, act, waitFor } from '@testing-library/react'
import { AuthProvider, useAuth } from '@/contexts/AuthContext'
import { login as apiLogin, register as apiRegister } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  login: jest.fn(),
  register: jest.fn(),
}))

const mockApiLogin = apiLogin as jest.MockedFunction<typeof apiLogin>
const mockApiRegister = apiRegister as jest.MockedFunction<typeof apiRegister>

const mockUser = {
  id: 'user-1',
  email: 'test@example.com',
  name: 'Test User',
  teamId: 'team-1',
  permissions: ['rules:read', 'users:read'],
  authProvider: 'local' as const,
  isActive: true,
  createdAt: '2024-01-01T00:00:00Z',
}

const mockToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEiLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9ucyI6WyJydWxlczpyZWFkIl0sImV4cCI6OTk5OTk5OTk5OX0.test'

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: jest.fn((key: string) => store[key] || null),
    setItem: jest.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: jest.fn((key: string) => {
      delete store[key]
    }),
    clear: jest.fn(() => {
      store = {}
    }),
  }
})()

Object.defineProperty(window, 'localStorage', { value: localStorageMock })

describe('AuthContext', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorageMock.clear()
  })

  it('starts with unauthenticated state', async () => {
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
    expect(result.current.token).toBeNull()
  })

  it('login updates auth state', async () => {
    mockApiLogin.mockResolvedValue({
      token: mockToken,
      user: mockUser,
    })

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    await act(async () => {
      await result.current.login({ email: 'test@example.com', password: 'password' })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user).toEqual(mockUser)
    expect(result.current.token).toBe(mockToken)
    expect(localStorageMock.setItem).toHaveBeenCalled()
  })

  it('logout clears auth state', async () => {
    mockApiLogin.mockResolvedValue({
      token: mockToken,
      user: mockUser,
    })

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    // Login first
    await act(async () => {
      await result.current.login({ email: 'test@example.com', password: 'password' })
    })

    expect(result.current.isAuthenticated).toBe(true)

    // Then logout
    act(() => {
      result.current.logout()
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
    expect(result.current.token).toBeNull()
    expect(localStorageMock.removeItem).toHaveBeenCalled()
  })

  it('register updates auth state', async () => {
    mockApiRegister.mockResolvedValue({
      token: mockToken,
      user: mockUser,
    })

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    await act(async () => {
      await result.current.register({
        email: 'test@example.com',
        name: 'Test User',
        password: 'password',
      })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user).toEqual(mockUser)
  })

  it('hasPermission returns correct value', async () => {
    mockApiLogin.mockResolvedValue({
      token: mockToken,
      user: mockUser,
    })

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    await act(async () => {
      await result.current.login({ email: 'test@example.com', password: 'password' })
    })

    expect(result.current.hasPermission('rules:read')).toBe(true)
    expect(result.current.hasPermission('rules:write')).toBe(false)
  })

  it('hasAnyPermission returns correct value', async () => {
    mockApiLogin.mockResolvedValue({
      token: mockToken,
      user: mockUser,
    })

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    await act(async () => {
      await result.current.login({ email: 'test@example.com', password: 'password' })
    })

    expect(result.current.hasAnyPermission('rules:read', 'admin:all')).toBe(true)
    expect(result.current.hasAnyPermission('admin:all', 'super:power')).toBe(false)
  })

  it('throws error when useAuth is used outside provider', () => {
    // Suppress console.error for this test
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => {
      renderHook(() => useAuth())
    }).toThrow('useAuth must be used within an AuthProvider')

    consoleSpy.mockRestore()
  })
})
