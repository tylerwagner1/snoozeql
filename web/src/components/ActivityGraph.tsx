import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, ReferenceLine, CartesianGrid } from 'recharts'

interface ActivityGraphProps {
  pattern: {
    idle_start_hour: number
    idle_end_hour: number
    avg_cpu?: number
  }
}

export function ActivityGraph({ pattern }: ActivityGraphProps) {
  // Generate 24-hour data
  const data = Array.from({ length: 24 }, (_, hour) => {
    // Determine if this hour is in the idle window
    let isIdle = false
    if (pattern.idle_start_hour < pattern.idle_end_hour) {
      // Normal case: e.g., 22:00 to 07:00
      isIdle = hour >= pattern.idle_start_hour && hour < pattern.idle_end_hour
    } else if (pattern.idle_start_hour === pattern.idle_end_hour) {
      // Edge case: same start and end hour - treat as 1-hour window
      isIdle = hour === pattern.idle_start_hour
    } else {
      // Overnight case: e.g., 22:00 to 06:00 (crosses midnight)
      isIdle = hour >= pattern.idle_start_hour || hour < pattern.idle_end_hour
    }

    // Use provided avg_cpu or estimate based on idle status
    // Active hours should show higher CPU than idle hours
    const idleCpu = pattern.avg_cpu ?? 1.0
    const activeCpu = idleCpu * 3 // Active hours typically 3x idle CPU
    const cpu = isIdle ? idleCpu : activeCpu

    return {
      hour: `${hour}:00`,
      cpu,
      isIdle,
    }
  })

  // Format hour label
  const formatXAxis = (tickItem: string) => {
    const hour = parseInt(tickItem.split(':')[0], 10)
    if (hour === 0) return '12AM'
    if (hour === 12) return '12PM'
    return hour < 12 ? `${hour}AM` : `${hour - 12}PM`
  }

  // Format tooltip
  const formatTooltip = (value: number) => {
    return `${value.toFixed(1)}% CPU`
  }

  return (
    <div className="h-32 w-full">
      <ResponsiveContainer>
        <AreaChart data={data} margin={{ top: 5, right: 10, left: -20, bottom: 0 }}>
          <defs>
            <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.3} />
              <stop offset="100%" stopColor="#3b82f6" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="idleGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#10b981" stopOpacity={0.3} />
              <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid
            vertical={false}
            stroke="#334155"
            strokeOpacity={0.2}
          />
          <XAxis
            dataKey="hour"
            tick={{ fill: '#94a3b8', fontSize: 10 }}
            tickLine={false}
            axisLine={false}
            interval={3} // Show every 3rd tick
            tickFormatter={formatXAxis}
          />
          <YAxis
            tick={{ fill: '#94a3b8', fontSize: 10 }}
            tickLine={false}
            axisLine={false}
            domain={[0, 30]}
            tickFormatter={(v) => `${v}%`}
          />
          <Tooltip
            contentStyle={{
              background: '#1e293b',
              border: 'none',
              borderRadius: '0.5rem',
            }}
            labelStyle={{ color: '#f8fafc', marginBottom: '0.5rem' }}
            itemStyle={{ color: '#e2e8f0' }}
            formatter={formatTooltip}
          />
          {/* Idle window area */}
          <Area
            type="monotone"
            dataKey="cpu"
            stroke="#10b981"
            fill="url(#idleGradient)"
            strokeWidth={2}
            dot={false}
            activeDot={false}
          />
          {/* Activity line (for contrast) */}
          <Area
            type="monotone"
            dataKey="cpu"
            stroke="#3b82f6"
            fill="url(#cpuGradient)"
            strokeWidth={1}
            strokeDasharray="4 4"
            opacity={0.5}
            dot={false}
            activeDot={false}
          />
          {/* Idle window markers */}
          <ReferenceLine
            x={`${pattern.idle_start_hour}:00`}
            stroke="#10b981"
            strokeDasharray="3 3"
            label={{
              value: 'Sleep',
              fill: '#10b981',
              fontSize: 10,
              position: 'top',
            }}
          />
          <ReferenceLine
            x={`${pattern.idle_end_hour}:00`}
            stroke="#f59e0b"
            strokeDasharray="3 3"
            label={{
              value: 'Wake',
              fill: '#f59e0b',
              fontSize: 10,
              position: 'top',
            }}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}

export default ActivityGraph
