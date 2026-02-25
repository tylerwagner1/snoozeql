import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import api from '../lib/api'

type MetricTab = 'cpu' | 'memory' | 'connections'
type TimeRange = '1h' | '6h' | '24h' | '7d'

interface MetricsChartProps {
  instanceId: string
}

// Map tab to CloudWatch metric name
const metricNameMap: Record<MetricTab, string> = {
  cpu: 'CPUUtilization',
  memory: 'FreeableMemory',
  connections: 'DatabaseConnections',
}

// Format X-axis labels based on time range
const formatXAxis = (timestamp: number, range: TimeRange): string => {
  const date = new Date(timestamp)
  if (range === '1h' || range === '6h') {
    return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
  }
  if (range === '24h') {
    return date.toLocaleTimeString('en-US', { hour: 'numeric' })
  }
  // 7d
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

// Get time domain based on range
const getTimeDomain = (range: TimeRange): [number, number] => {
  const now = Date.now()
  const durations: Record<TimeRange, number> = {
    '1h': 60 * 60 * 1000,
    '6h': 6 * 60 * 60 * 1000,
    '24h': 24 * 60 * 60 * 1000,
    '7d': 7 * 24 * 60 * 60 * 1000,
  }
  return [now - durations[range], now]
}

// Get appropriate tick count based on range
const getTickCount = (range: TimeRange): number => {
  switch (range) {
    case '1h': return 6  // Every 10 min
    case '6h': return 6  // Every hour
    case '24h': return 8 // Every 3 hours
    case '7d': return 7  // Every day
  }
}

export function MetricsChart({ instanceId }: MetricsChartProps) {
  const [activeTab, setActiveTab] = useState<MetricTab>('cpu')
  const [timeRange, setTimeRange] = useState<TimeRange>('24h')

  const { data, isLoading, error } = useQuery({
    queryKey: ['metrics-history', instanceId, timeRange],
    queryFn: () => api.getMetricsHistory(instanceId, timeRange),
  })

  // Filter data for active metric and convert to timestamp
  const chartData = useMemo(() => {
    if (!data) return []
    return data
      .filter((m: { metric_name: string }) => m.metric_name === metricNameMap[activeTab])
      .map((m: { hour: string; avg_value: number }) => ({
        timestamp: new Date(m.hour).getTime(),
        value: m.avg_value,
      }))
      .sort((a, b) => a.timestamp - b.timestamp)
  }, [data, activeTab])

  // Get domain for the time axis
  const timeDomain = useMemo(() => getTimeDomain(timeRange), [timeRange])

  const isPercentage = activeTab === 'cpu' || activeTab === 'memory'

  const tabs = [
    { key: 'cpu' as const, label: 'CPU' },
    { key: 'memory' as const, label: 'Memory' },
    { key: 'connections' as const, label: 'Connections' },
  ]

  const timeRanges: TimeRange[] = ['1h', '6h', '24h', '7d']

  return (
    <div className="bg-white shadow-sm border rounded-lg p-6">
      {/* Header with tabs and time range */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex border-b border-gray-200">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.key
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
        
        <div className="flex bg-gray-100 rounded-lg p-1">
          {timeRanges.map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-3 py-1 text-sm rounded-md transition-colors ${
                timeRange === range
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              {range}
            </button>
          ))}
        </div>
      </div>

      {/* Chart area */}
      {isLoading ? (
        <div className="h-[150px] flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
        </div>
      ) : error ? (
        <div className="h-[150px] flex items-center justify-center">
          <p className="text-red-500 text-sm">Failed to load metrics</p>
        </div>
      ) : (
        <div className="h-[150px] relative">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData} margin={{ top: 5, right: 5, left: 0, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e5e7eb" />
              <XAxis
                dataKey="timestamp"
                type="number"
                scale="time"
                domain={timeDomain}
                tickFormatter={(ts) => formatXAxis(ts, timeRange)}
                tick={{ fill: '#9ca3af', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                tickCount={getTickCount(timeRange)}
              />
              <YAxis
                domain={isPercentage ? [0, 100] : [0, 'auto']}
                tickFormatter={(v) => isPercentage ? `${v}%` : String(v)}
                tick={{ fill: '#9ca3af', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                width={40}
              />
              <Tooltip
                cursor={{ strokeDasharray: '3 3' }}
                contentStyle={{
                  background: '#fff',
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.5rem',
                  fontSize: '12px',
                  padding: '8px 12px',
                }}
                labelStyle={{
                  color: '#374151',
                  fontWeight: 500,
                  marginBottom: '4px',
                }}
                labelFormatter={(ts) => new Date(ts).toLocaleString()}
                formatter={(value: number) => [
                  isPercentage ? `${value.toFixed(1)}%` : value.toFixed(1),
                  activeTab === 'cpu' ? 'CPU' : activeTab === 'memory' ? 'Memory' : 'Connections'
                ]}
              />
              {chartData.length > 0 && (
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke="#3b82f6"
                  dot={false}
                  strokeWidth={2}
                  connectNulls={false}
                  activeDot={{ r: 4, fill: '#3b82f6', stroke: '#fff', strokeWidth: 2 }}
                />
              )}
            </LineChart>
          </ResponsiveContainer>
          {chartData.length === 0 && (
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
              <p className="text-gray-500 text-sm">No data available</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
