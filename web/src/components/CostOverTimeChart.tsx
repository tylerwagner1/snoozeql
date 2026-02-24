import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import { useState, useEffect } from 'react'
import type { Instance } from '../lib/api'

interface CostDataPoint {
  date: string
  label: string
  cost: number
  costDollars: number
}

interface CostOverTimeChartProps {
  instances: Instance[]
}

enum TimeRange {
  Daily = 'daily',
  Weekly = 'weekly'
}

export function CostOverTimeChart({ instances }: CostOverTimeChartProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>(TimeRange.Weekly)
  const [costData, setCostData] = useState<CostDataPoint[]>([])

  // Calculate estimated cost for a given time range
  // Uses the formula: sum(hourly_cost * 24) for each running instance
  const calculateEstimatedCost = (days: number): CostDataPoint[] => {
    const data: CostDataPoint[] = []
    const today = new Date()
    
    // Get all instances with their hourly costs
    const instanceCosts = instances
      .filter(inst => 
        inst.status === 'available' || 
        inst.status === 'running' || 
        inst.status === 'starting'
      )
      .map(inst => ({
        name: inst.name,
        cost: inst.hourly_cost_cents
      }))
    
    // Get total hourly cost
    const totalHourlyCostCents = instanceCosts.reduce((sum, i) => sum + i.cost, 0)
    
    for (let day = days - 1; day >= 0; day--) {
      const currentDay = new Date(today)
      currentDay.setDate(today.getDate() - day)
      
      // Estimated cost: total hourly cost * 24 hours
      const estimatedDailyCostCents = totalHourlyCostCents * 24
      
      data.push({
        date: currentDay.toISOString().split('T')[0],
        label: currentDay.toLocaleDateString('en-US', { 
          month: 'short', 
          day: 'numeric',
          weekday: day === days - 1 ? 'short' : undefined // Only show weekday for last day
        }),
        cost: estimatedDailyCostCents,
        costDollars: estimatedDailyCostCents / 100,
      })
    }
    
    return data
  }

  useEffect(() => {
    if (timeRange === TimeRange.Weekly) {
      setCostData(calculateEstimatedCost(7))
    } else {
      // For 24h, estimate based on current hourly cost
      const data: CostDataPoint[] = []
      const totalHourlyCostCents = instances
        .filter(inst => 
          inst.status === 'available' || 
          inst.status === 'running' || 
          inst.status === 'starting'
        )
        .reduce((sum, inst) => sum + inst.hourly_cost_cents, 0)
      
      const today = new Date()
      for (let hour = 0; hour < 24; hour++) {
        const hourDate = new Date(today)
        hourDate.setHours(hour)
        data.push({
          date: today.toISOString().split('T')[0],
          label: hourDate.toLocaleTimeString('en-US', { 
            hour: 'numeric',
            hour12: true,
            minute: '2-digit'
          }),
          cost: totalHourlyCostCents,
          costDollars: totalHourlyCostCents / 100,
        })
      }
      setCostData(data)
    }
  }, [instances, timeRange])

  const currentHourlyCost = instances
    .filter(inst => 
      inst.status === 'available' || 
      inst.status === 'running' || 
      inst.status === 'starting'
    )
    .reduce((sum, inst) => sum + inst.hourly_cost_cents, 0) / 100

  const maxCost = costData.length > 0 
    ? Math.max(...costData.map(d => d.costDollars), 1)
    : 1

  return (
    <div className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-white">Cost Over Time</h2>
        <div className="flex items-center space-x-2">
          <div className="text-sm text-slate-400">
            Current: <span className="text-cyan-400 font-medium">${currentHourlyCost.toFixed(2)}/hr</span>
          </div>
          <div className="border-l border-slate-600 h-6 mx-2"></div>
          <div className="flex bg-slate-900 rounded-lg p-1">
            <button
              onClick={() => setTimeRange(TimeRange.Daily)}
              className={`px-3 py-1 text-sm rounded-md transition-colors ${
                timeRange === TimeRange.Daily
                  ? 'bg-slate-700 text-white'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              24h
            </button>
            <button
              onClick={() => setTimeRange(TimeRange.Weekly)}
              className={`px-3 py-1 text-sm rounded-md transition-colors ${
                timeRange === TimeRange.Weekly
                  ? 'bg-slate-700 text-white'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              7 days
            </button>
          </div>
        </div>
      </div>
      
      {instances.length === 0 ? (
        <div className="h-64 flex items-center justify-center">
          <p className="text-slate-400">No instances discovered yet</p>
        </div>
      ) : (
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={costData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
              <defs>
                <linearGradient id="costGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#06b6d4" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#06b6d4" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid vertical={false} stroke="#334155" strokeOpacity={0.3} />
              <XAxis
                dataKey="label"
                tick={{ fill: '#94a3b8', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                tick={{ fill: '#94a3b8', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v) => `$${v.toFixed(0)}`}
                width={50}
                domain={[0, Math.ceil(maxCost * 1.1)]}
              />
              <Tooltip
                contentStyle={{
                  background: '#1e293b',
                  border: '1px solid #334155',
                  borderRadius: '0.5rem',
                }}
                labelStyle={{ color: '#f8fafc', marginBottom: '0.5rem' }}
                formatter={(value: number) => [`$${value.toFixed(2)}`, 'Daily Cost']}
              />
              <Area
                type="monotone"
                dataKey="costDollars"
                stroke="#06b6d4"
                fill="url(#costGradient)"
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  )
}
