import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import type { Instance } from '../lib/api'

interface CostOverTimeChartProps {
  instances: Instance[]
}

interface CostDataPoint {
  date: string
  label: string
  cost: number
  costDollars: number
}

export function CostOverTimeChart({ instances }: CostOverTimeChartProps) {
  // Generate cost data for last 7 days
  const generateCostData = (): CostDataPoint[] => {
    const data: CostDataPoint[] = []
    const today = new Date()
    
    // Calculate total hourly cost from all running instances
    const totalHourlyCostCents = instances
      .filter(inst => inst.status === 'available' || inst.status === 'running' || inst.status === 'starting')
      .reduce((sum, inst) => sum + inst.hourly_cost_cents, 0)
    
    for (let day = 6; day >= 0; day--) {
      const currentDay = new Date(today)
      currentDay.setDate(today.getDate() - day)
      
      // Simulate cost variation based on time of day patterns
      // In a real implementation, this would come from actual historical data
      let dailyCost = 0
      
      // Business hours (9-17) = full cost, nights (22-7) = minimal, other times = 20%
      const businessHours = 8 // 9AM - 5PM
      const offHours = 9 // 10PM - 7AM  
      const transitionHours = 7 // remaining hours
      
      dailyCost = (totalHourlyCostCents * businessHours) + 
                  (totalHourlyCostCents * 0.2 * transitionHours) + 
                  (totalHourlyCostCents * 0 * offHours)
      
      data.push({
        date: currentDay.toISOString().split('T')[0],
        label: currentDay.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        cost: dailyCost,
        costDollars: dailyCost / 100,
      })
    }
    
    return data
  }

  const costData = generateCostData()
  const maxCost = Math.max(...costData.map(d => d.costDollars), 1)
  const currentHourlyCost = instances
    .filter(inst => inst.status === 'available' || inst.status === 'running' || inst.status === 'starting')
    .reduce((sum, inst) => sum + inst.hourly_cost_cents, 0) / 100

  return (
    <div className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-white">Cost Over Time (7 days)</h2>
        <div className="text-sm text-slate-400">
          Current: <span className="text-cyan-400 font-medium">${currentHourlyCost.toFixed(2)}/hr</span>
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
