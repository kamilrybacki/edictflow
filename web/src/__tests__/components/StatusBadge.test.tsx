import { render, screen, waitFor } from '@testing-library/react'
import { StatusBadge } from '@/components/StatusBadge'
import { fetchServiceInfo } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  fetchServiceInfo: jest.fn(),
}))

const mockFetchServiceInfo = fetchServiceInfo as jest.MockedFunction<typeof fetchServiceInfo>

describe('StatusBadge', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('shows loading state initially', () => {
    mockFetchServiceInfo.mockImplementation(() => new Promise(() => {}))
    render(<StatusBadge />)
    expect(screen.getByText('Connecting...')).toBeInTheDocument()
  })

  it('shows connected state when API responds', async () => {
    mockFetchServiceInfo.mockResolvedValue({
      service: 'edictflow',
      version: '1.0.0',
      status: 'running',
    })

    render(<StatusBadge />)

    await waitFor(() => {
      expect(screen.getByText('edictflow v1.0.0')).toBeInTheDocument()
    })
  })

  it('shows disconnected state when API fails', async () => {
    mockFetchServiceInfo.mockResolvedValue(null)

    render(<StatusBadge />)

    await waitFor(() => {
      expect(screen.getByText('API Disconnected')).toBeInTheDocument()
    })
  })

  it('has correct indicator color for connected state', async () => {
    mockFetchServiceInfo.mockResolvedValue({
      service: 'edictflow',
      version: '1.0.0',
      status: 'running',
    })

    const { container } = render(<StatusBadge />)

    await waitFor(() => {
      const indicator = container.querySelector('.bg-green-500')
      expect(indicator).toBeInTheDocument()
    })
  })

  it('has correct indicator color for disconnected state', async () => {
    mockFetchServiceInfo.mockResolvedValue(null)

    const { container } = render(<StatusBadge />)

    await waitFor(() => {
      const indicator = container.querySelector('.bg-red-500')
      expect(indicator).toBeInTheDocument()
    })
  })
})
