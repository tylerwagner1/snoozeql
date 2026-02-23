const API_BASE_URL = 'http://localhost:8080/api/v1'

export interface Instance {
  id: string
  cloud_account_id: string
  account_name?: string
  provider: string
  provider_id: string
  name: string
  region: string
  instance_type: string
  engine: string
  status: string
  managed: boolean
  tags: Record<string, string>
  hourly_cost_cents: number
  created_at: string
  updated_at: string
}

export interface Selector {
  name: { pattern: string; type: string }
  provider?: string
  region?: { pattern: string; type: string }
  engine?: { pattern: string; type: string }
  tags?: Record<string, { pattern: string; type: string }>
}

export interface Schedule {
  id: string
  name: string
  description: string
  selectors: Selector[]
  timezone: string
  sleep_cron: string
  wake_cron: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface Recommendation {
  id: string
  instance_name: string
  pattern: string
  suggested_action: 'start' | 'stop' | 'no_change'
  estimated_daily_savings: string
  confidence: number
  activity_pattern: {
    active_hours: number[]
    inactive_hours: number[]
    confidence: number
  }
  created_at: string
  status: 'pending' | 'ignored' | 'applied'
}

export interface CloudAccount {
  id: string
  name: string
  provider: string
  regions: string[]
  connection_status?: 'connected' | 'syncing' | 'failed' | 'unknown'
  last_sync_at?: string
  last_error?: string
  created_at: string
}

export interface Event {
  id: string
  instance_id: string
  event_type: string
  triggered_by: string
  previous_status: string
  new_status: string
  metadata?: Record<string, unknown>
  created_at: string
}

export interface BulkOperationResponse {
  success: string[]
  failed: Array<{
    instance_id: string
    error: string
  }>
}

export interface Stats {
  total_instances: number
  running_instances: number
  stopped_instances: number
  savings_7d: number
  pending_actions: number
}

const api = {
  async get<T>(path: string): Promise<T> {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      headers: {
        'Authorization': 'Bearer dev-key',
      },
    })
    if (!response.ok) {
      throw new Error(`API error: ${response.statusText}`)
    }
    return response.json()
  },

  async post<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer dev-key',
      },
      body: body ? JSON.stringify(body) : undefined,
    })
    if (!response.ok) {
      throw new Error(`API error: ${response.statusText}`)
    }
    return response.json()
  },

  async put<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer dev-key',
      },
      body: body ? JSON.stringify(body) : undefined,
    })
    if (!response.ok) {
      throw new Error(`API error: ${response.statusText}`)
    }
    return response.json()
  },

  async del(path: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'DELETE',
      headers: {
        'Authorization': 'Bearer dev-key',
      },
    })
    if (!response.ok) {
      throw new Error(`API error: ${response.statusText}`)
    }
  },

  // Instances
  getInstances: () => api.get<Instance[]>('/instances'),
  getInstance: (id: string) => api.get<Instance>(`/instances/${id}`),
  startInstance: (id: string) => api.post<{ success: boolean; instance_id: string; provider: string; status: string }>(`/instances/${id}/start`),
  stopInstance: (id: string) => api.post<{ success: boolean; instance_id: string; provider: string; status: string }>(`/instances/${id}/stop`),
  bulkStopInstances: (instanceIds: string[]) =>
    api.post<BulkOperationResponse>('/instances/bulk-stop', { instance_ids: instanceIds }),
  bulkStartInstances: (instanceIds: string[]) =>
    api.post<BulkOperationResponse>('/instances/bulk-start', { instance_ids: instanceIds }),

  // Schedules
  getSchedules: () => api.get<Schedule[]>('/schedules'),
  getSchedule: (id: string) => api.get<Schedule>(`/schedules/${id}`),
  createSchedule: (data: Omit<Schedule, 'id' | 'created_at' | 'updated_at'>) => api.post<Schedule>('/schedules', data),
  updateSchedule: (id: string, data: Partial<Schedule>) => api.put<Schedule>(`/schedules/${id}`, data),
  deleteSchedule: (id: string) => api.del(`/schedules/${id}`),
  
  // Schedule filter preview
  previewFilter: (selectors: Selector[], operator: 'and' | 'or' = 'and') =>
    api.post<{ matched_count: number; total_count: number; instances: Instance[] }>(
      '/schedules/preview-filter',
      { selectors, operator }
    ),

  // Recommendations
  getRecommendations: () => api.get<Recommendation[]>('/recommendations'),
  getRecommendation: (id: string) => api.get<Recommendation>(`/recommendations/${id}`),
  applyRecommendation: (id: string) => api.post<void>(`/recommendations/${id}/apply`),
  ignoreRecommendation: (id: string) => api.post<void>(`/recommendations/${id}/ignore`),

  // Stats
  getStats: () => api.get<Stats>('/stats'),

  // Cloud accounts
  getCloudAccounts: () => api.get<CloudAccount[]>('/cloud-accounts'),
  createCloudAccount: (data: { name: string; provider: string; regions?: string[]; credentials: { [key: string]: string } }) => api.post<CloudAccount>('/cloud-accounts', data),
  updateCloudAccount: (id: string, data: { name: string; regions?: string[]; credentials: { [key: string]: string } }) => api.put<{ success: boolean }>(`/cloud-accounts/${id}`, data),
  deleteCloudAccount: (id: string) => api.del(`/cloud-accounts/${id}`),

  // Discovery
  refreshInstances: () => api.post<{ success: boolean; message: string }>('/discovery/refresh'),

  // Events/Audit Log
  getEvents: (limit?: number, offset?: number) => {
    const params = new URLSearchParams()
    if (limit) params.set('limit', limit.toString())
    if (offset) params.set('offset', offset.toString())
    const query = params.toString()
    return api.get<Event[]>(`/events${query ? `?${query}` : ''}`)
  },
  getEventsByInstance: (instanceId: string) => api.get<Event[]>(`/instances/${instanceId}/events`),
}

export default api
