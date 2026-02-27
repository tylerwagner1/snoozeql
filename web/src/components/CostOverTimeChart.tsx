import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import { useState, useMemo } from 'react'
import type { Instance, Event } from '../lib/api'

interface CostDataPoint {
  date: string
  label: string
  actualCost: number
  potentialCost: number
  savings: number
}

interface CostOverTimeChartProps {
  instances: Instance[]
  events: Event[]
}

enum TimeRange {
  Daily = 'daily',
  Weekly = 'weekly'
}

export function CostOverTimeChart({ instances, events }: CostOverTimeChartProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>(TimeRange.Weekly)

  // Build instance cost lookup map
  const instanceCostMap = useMemo(() => 
    new Map(instances.map(i => [i.id, i.hourly_cost_cents])),
    [instances]
  )

  // Calculate total potential daily cost (if all instances ran 24/7)
  const totalPotentialDailyCost = useMemo(() => 
    instances.reduce((sum, inst) => sum + (inst.hourly_cost_cents / 100) * 24, 0),
    [instances]
  )

  // Build event timeline for calculating running hours per day
  const costData = useMemo(() => {
    const data: CostDataPoint[] = []
    const today = new Date()
    today.setHours(23, 59, 59, 999)
    
    const days = timeRange === TimeRange.Weekly ? 7 : 1
    
    // Build instance state timeline from events
    // Track when each instance was running vs stopped
    const instanceStates: Map<string, { running: boolean; since: Date }[]> = new Map()
    
    // Initialize all instances as running (assume running unless we have stop events)
    instances.forEach(inst => {
      const isCurrentlyRunning = inst.status === 'available' || inst.status === 'running' || inst.status === 'starting'
      instanceStates.set(inst.id, [{ running: isCurrentlyRunning, since: new Date(0) }])
    })
    
    // Sort events by timestamp ascending
    const sortedEvents = [...events].sort((a, b) => 
      new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    )
    
    // Process events to build state timeline
    sortedEvents.forEach(event => {
      const states = instanceStates.get(event.instance_id)
      if (!states) return
      
      const isStartEvent = event.event_type === 'start' || event.event_type === 'wake'
      const isStopEvent = event.event_type === 'stop' || event.event_type === 'sleep'
      
      if (isStartEvent || isStopEvent) {
        states.push({
          running: isStartEvent,
          since: new Date(event.created_at)
        })
      }
    })
    
    // Calculate cost for each day
    for (let dayOffset = days - 1; dayOffset >= 0; dayOffset--) {
      const dayStart = new Date(today)
      dayStart.setDate(today.getDate() - dayOffset)
      dayStart.setHours(0, 0, 0, 0)
      
      const dayEnd = new Date(dayStart)
      dayEnd.setHours(23, 59, 59, 999)
      
      let actualCostCents = 0
      
      // For each instance, calculate hours running on this day
      instances.forEach(inst => {
        const states = instanceStates.get(inst.id) || []
        const hourlyCost = inst.hourly_cost_cents
        
        // Find running hours for this day
        let runningHours = 0
        
        for (let hour = 0; hour < 24; hour++) {
          const hourTime = new Date(dayStart)
          hourTime.setHours(hour)
          
          // Find the state at this hour
          let wasRunning = true // Default to running
          for (const state of states) {
            if (state.since <= hourTime) {
              wasRunning = state.running
            } else {
              break
            }
          }
          
          if (wasRunning) {
            runningHours++
          }
        }
        
        actualCostCents += hourlyCost * runningHours
      })
      
      const actualCost = actualCostCents / 100
      const potentialCost = totalPotentialDailyCost
      const savings = potentialCost - actualCost
      
      data.push({
        date: dayStart.toISOString().split('T')[0],
        label: dayStart.toLocaleDateString('en-US', { 
          month: 'short', 
          day: 'numeric',
        }),
        actualCost,
        potentialCost,
        savings: savings > 0 ? savings : 0,
      })
    }
    
    return data
  }, [instances, events, timeRange, instanceCostMap, totalPotentialDailyCost])

  const currentHourlyCost = instances
    .filter(inst => 
      inst.status === 'available' || 
      inst.status === 'running' || 
      inst.status === 'starting'
    )
    .reduce((sum, inst) => sum + inst.hourly_cost_cents, 0) / 100

  const maxCost = costData.length > 0 
    ? Math.max(...costData.map(d => Math.max(d.actualCost, d.potentialCost)), 1)
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
              Today
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
                <linearGradient id="actualCostGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#06b6d4" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#06b6d4" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="potentialCostGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#94a3b8" stopOpacity={0.1} />
                  <stop offset="100%" stopColor="#94a3b8" stopOpacity={0} />
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
                formatter={(value: number, name: string) => {
                  const labels: Record<string, string> = {
                    actualCost: 'Actual Cost',
                    potentialCost: 'If Always On',
                    savings: 'Saved',
                  }
                  return [`$${value.toFixed(2)}`, labels[name] || name]
                }}
              />
              {/* Potential cost (faded background) */}
              <Area
                type="monotone"
                dataKey="potentialCost"
                stroke="#94a3b8"
                fill="url(#potentialCostGradient)"
                strokeWidth={1}
                strokeDasharray="4 4"
              />
              {/* Actual cost (main line) */}
              <Area
                type="monotone"
                dataKey="actualCost"
                stroke="#06b6d4"
                fill="url(#actualCostGradient)"
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
      
      {/* Legend */}
      <div className="flex items-center justify-center gap-6 mt-4 text-xs">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded bg-cyan-500"></div>
          <span className="text-slate-400">Actual Cost</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded bg-slate-500 opacity-50"></div>
          <span className="text-slate-400">If Always On</span>
        </div>
      </div>
    </div>
  )
}
