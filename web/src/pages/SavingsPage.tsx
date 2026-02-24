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
  const [ongoingCost, setOngoingCost] = useState<number | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        const days = dateRange === '7d' ? 7 : dateRange === '30d' ? 30 : 90

        const [summaryData, dailyData, byInstanceData, ongoingData] = await Promise.all([
          api.getSavingsSummary(days),
          api.getDailySavings(days),
          api.getSavingsByInstance(days, 10),
          api.getOngoingCost(),
        ])

        setSummary(summaryData)
        setDailySavings(dailyData.daily_savings || [])
        setInstanceSavings(byInstanceData || [])
        setOngoingCost(ongoingData.ongoing_cost_cents)
      } catch (err) {
        console.error('Failed to fetch savings data:', err)
        // Keep previous data on error to avoid flicker
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [dateRange])

  return (
    <div className="space-y-6 bg-slate-900 min-h-screen p-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Cost Over Time</h1>
          <p className="text-sm text-slate-400 mt-1">
            Track your database costs and savings with SnoozeQL
          </p>
        </div>
        <DateRangeSelector value={dateRange} onChange={setDateRange} />
      </div>

      {/* Summary Cards */}
      <SavingsSummaryCards data={summary} loading={loading} />

      {/* Savings Chart */}
      <SavingsChart data={dailySavings} ongoingCost={ongoingCost} loading={loading} />

      {/* Two-column layout for table and projection on larger screens */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Instance Savings Table */}
        <InstanceSavingsTable data={instanceSavings} loading={loading} />

        {/* Cost Projection */}
        <CostProjection
          ongoingCost={ongoingCost}
          loading={loading}
        />
      </div>
    </div>
  )
}
