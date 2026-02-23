import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import { formatCurrency } from '../../lib/formatters'

interface DailySavings {
  date: string
  savings_cents: number
  stopped_minutes: number
}

interface SavingsChartProps {
  data: DailySavings[]
  loading: boolean
}

export function SavingsChart({ data, loading }: SavingsChartProps) {
  if (loading) {
    return (
      <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
        <div className="h-4 bg-slate-700 rounded w-40 mb-4 animate-pulse" />
        <div className="h-64 bg-slate-700/50 rounded animate-pulse" />
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="bg-slate-800/50 rounded-xl p-8 border border-slate-700 text-center">
        <p className="text-slate-400">No savings data yet.</p>
        <p className="text-sm text-slate-500 mt-2">
          Savings are calculated when instances are stopped via schedules or manual actions.
        </p>
      </div>
    )
  }

  // Transform cents to dollars for display
  const chartData = data.map((d) => ({
    date: formatDate(d.date),
    savings: d.savings_cents / 100,
    savingsCents: d.savings_cents,
    stoppedHours: Math.round(d.stopped_minutes / 60 * 10) / 10,
  }))

  return (
    <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
      <h2 className="text-lg font-semibold text-white mb-4">Savings Over Time</h2>
      <div className="h-64">
        <ResponsiveContainer>
          <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="savingsGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#10b981" stopOpacity={0.3} />
                <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} stroke="#334155" strokeOpacity={0.2} />
            <XAxis
              dataKey="date"
              tick={{ fill: '#94a3b8', fontSize: 10 }}
              tickLine={false}
              axisLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fill: '#94a3b8', fontSize: 10 }}
              tickLine={false}
              axisLine={false}
              tickFormatter={(v) => `$${v}`}
              width={50}
            />
            <Tooltip
              contentStyle={{
                background: '#1e293b',
                border: '1px solid #334155',
                borderRadius: '0.5rem',
              }}
              labelStyle={{ color: '#f8fafc', marginBottom: '0.5rem' }}
              formatter={(value: number, name: string) => {
                if (name === 'savings') {
                  return [formatCurrency(value * 100), 'Savings']
                }
                return [value, name]
              }}
            />
            <Area
              type="monotone"
              dataKey="savings"
              stroke="#10b981"
              fill="url(#savingsGradient)"
              strokeWidth={2}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
