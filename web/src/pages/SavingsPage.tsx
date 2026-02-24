import { useState, useEffect } from 'react'
import api from '../lib/api'
import type { SavingsSummary, InstanceSavingsItem } from '../lib/api'
import { DateRangeSelector, DateRange } from '../components/savings/DateRangeSelector'
import { SavingsSummaryCards } from '../components/savings/SavingsSummaryCards'
import { SavingsChart } from '../components/savings/SavingsChart'
import { SavingsTable } from '../components/savings/SavingsTable'

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
  const [topSavers, setTopSavers] = useState<InstanceSavingsItem[]>([])

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        const days = dateRange === '7d' ? 7 : dateRange === '30d' ? 30 : 90

        const [summaryData, dailyData, topSaversData] = await Promise.all([
          api.getSavingsSummary(days),
          api.getDailySavings(days),
          api.getSavingsByInstance(days, 5),
        ])

        setSummary(summaryData)
        setDailySavings(dailyData.daily_savings || [])
        setTopSavers(topSaversData || [])
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
      <SavingsChart data={dailySavings} loading={loading} />

      {/* Top Savings Table */}
      <SavingsTable data={topSavers} loading={loading} />
    </div>
  )
}
