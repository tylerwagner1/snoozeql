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
  ongoingCost: number | null
  loading: boolean
}

export function SavingsChart({ data, ongoingCost, loading }: SavingsChartProps) {
  if (loading) {
    return (
    <div className="bg-slate-800 rounded-xl p-6 border border-slate-700">
        <div className="h-4 bg-slate-700 rounded w-36 mb-4 animate-pulse" />
        <div className="grid grid-cols-2 gap-6 mb-6">
          <div className="h-24 bg-slate-700 rounded-lg animate-pulse" />
          <div className="h-24 bg-slate-700 rounded-lg animate-pulse" />
        </div>
        <div className="h-20 bg-slate-700 rounded-lg animate-pulse" />
      </div>
    )
  }

  // Current ongoing cost in dollars
  const currentOngoingCost = ongoingCost ? ongoingCost / 100 : 0

  // Handle empty data - show current ongoing cost
  if (data.length === 0) {
    return (
      <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
        <h2 className="text-lg font-semibold text-white mb-4">Cost Over Time</h2>
        <div className="mb-4 p-4 bg-green-900/20 rounded-lg border border-green-500/30">
          <div className="flex justify-between items-center">
            <span className="text-sm text-slate-300">Current ongoing cost:</span>
            <span className="text-2xl font-bold text-green-400">{formatCurrency(ongoingCost!)}</span>
          </div>
          <p className="text-xs text-slate-400 mt-1">Cost of running {ongoingCost !== null && ongoingCost > 0 ? "instances" : "no instances"} per hour</p>
        </div>
        <div className="bg-slate-700/50 rounded-lg p-4 text-center">
          <p className="text-slate-400">No historical savings data yet.</p>
          <p className="text-sm text-slate-500 mt-2">
            Savings will appear when instances are stopped via schedules or manual actions.
          </p>
        </div>
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
      <h2 className="text-lg font-semibold text-white mb-4">Cost Over Time</h2>
      {currentOngoingCost > 0 && (
        <div className="mb-4 p-3 bg-slate-700/50 rounded-lg">
          <div className="flex justify-between items-center">
            <span className="text-sm text-slate-300">Current ongoing cost:</span>
            <span className="text-lg font-semibold text-white">{formatCurrency(ongoingCost!)}</span>
          </div>
          <div className="text-xs text-slate-400 mt-1">Cost of running instances per hour</div>
        </div>
      )}
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
