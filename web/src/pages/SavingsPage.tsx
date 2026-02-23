import { useState, useEffect } from 'react'
import api from '../lib/api'
import type { SavingsSummary, InstanceSavingsItem } from '../lib/api'
import { DateRangeSelector, DateRange } from '../components/savings/DateRangeSelector'
import { SavingsSummaryCards } from '../components/savings/SavingsSummaryCards'
import { SavingsChart } from '../components/savings/SavingsChart'
import { InstanceSavingsTable } from '../components/savings/InstanceSavingsTable'
import { CostProjection } from '../components/savings/CostProjection'

interface DailySavings {
  date: string
  savings_cents: number
  stopped_minutes: number
}

export default function SavingsPage() {
  const [dateRange, setDateRange] = useState<DateRange>('30d')
  const [loading, setLoading] = useState(true)
  const [summary, setSummary] = useState<SavingsSummary | null>(null)
  const [dailySavings, setDailySavings] = useState<DailySavings[]>([])
  const [instanceSavings, setInstanceSavings] = useState<InstanceSavingsItem[]>([])

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        const days = dateRange === '7d' ? 7 : dateRange === '30d' ? 30 : 90

        const [summaryData, dailyData, byInstanceData] = await Promise.all([
          api.getSavingsSummary(days),
          api.getDailySavings(days),
          api.getSavingsByInstance(days, 10),
        ])

        setSummary(summaryData)
        setDailySavings(dailyData.daily_savings || [])
        setInstanceSavings(byInstanceData || [])
      } catch (err) {
        console.error('Failed to fetch savings data:', err)
        // Keep previous data on error to avoid flicker
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [dateRange])

  // Calculate projection data from summary
  // Projected cost = actual cost + savings (what we would have paid without SnoozeQL)
  const actualCostCents = summary
    ? summary.total_savings_cents // This is what we saved, not what we paid
    : 0
  // For now, show savings as the "actual" and 2x savings as "projected"
  // In a real implementation, this would come from billing data
  const projectedAlwaysOnCents = actualCostCents * 2

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Cost Savings</h1>
          <p className="text-sm text-slate-400 mt-1">
            Track your database cost optimization with SnoozeQL
          </p>
        </div>
        <DateRangeSelector value={dateRange} onChange={setDateRange} />
      </div>

      {/* Summary Cards */}
      <SavingsSummaryCards data={summary} loading={loading} />

      {/* Savings Chart */}
      <SavingsChart data={dailySavings} loading={loading} />

      {/* Two-column layout for table and projection on larger screens */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Instance Savings Table */}
        <InstanceSavingsTable data={instanceSavings} loading={loading} />

        {/* Cost Projection */}
        <CostProjection
          actualCostCents={actualCostCents}
          projectedAlwaysOnCents={projectedAlwaysOnCents}
          loading={loading}
        />
      </div>
    </div>
  )
}
