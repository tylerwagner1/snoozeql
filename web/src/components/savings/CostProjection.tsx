import { AlertTriangle } from 'lucide-react'
import { formatCurrency } from '../../lib/formatters'

interface CostProjectionProps {
  ongoingCost: number | null
  loading: boolean
}

export function CostProjection({
  ongoingCost,
  loading,
}: CostProjectionProps) {
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

  // Handle case where no ongoing cost data available
  if (!ongoingCost || ongoingCost === 0) {
    return (
      <div className="bg-slate-800 rounded-xl p-8 border border-slate-700 text-center">
        <p className="text-slate-400">No instance data available.</p>
        <p className="text-sm text-slate-500 mt-2">
          Once instances are stopped and restarted with schedules, cost projections will appear.
        </p>
      </div>
    )
  }

  // Calculate costs
  // Projected "Always On" = ongoingCost * 2 (what we'd pay if running 24/7)
  const projectedAlwaysOnCents = ongoingCost * 2
  const actualCostCents = ongoingCost
  const savingsCents = projectedAlwaysOnCents - actualCostCents
  const savingsPercent = ((savingsCents / projectedAlwaysOnCents) * 100).toFixed(1)

  return (
    <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
      <h2 className="text-lg font-semibold text-white mb-4">Cost Comparison</h2>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 mb-6">
        {/* Projected "Always On" Cost */}
        <div className="p-4 bg-slate-700/30 rounded-lg">
          <p className="text-sm text-slate-400 mb-1">If Always Running</p>
          <p className="text-2xl font-bold text-slate-300">
            {formatCurrency(projectedAlwaysOnCents)}
          </p>
          <p className="text-xs text-slate-500 mt-1">24/7 uptime estimate</p>
        </div>

        {/* Actual Cost */}
        <div className="p-4 bg-green-900/20 rounded-lg border border-green-500/30">
          <p className="text-sm text-slate-400 mb-1">Actual Cost</p>
          <p className="text-2xl font-bold text-green-400">
            {formatCurrency(actualCostCents)}
          </p>
          <p className="text-xs text-green-500/70 mt-1">With SnoozeQL savings</p>
        </div>
      </div>

      {/* Savings Summary */}
      <div className="p-4 bg-green-900/30 rounded-lg border border-green-500/20 mb-4">
        <p className="text-sm text-slate-400">You save by stopping instances</p>
        <p className="text-3xl font-bold text-green-400">
          {formatCurrency(savingsCents)}{' '}
          <span className="text-lg text-green-500">({savingsPercent}%)</span>
        </p>
      </div>

      {/* SAV-05 Required Disclaimer */}
      <div className="bg-yellow-900/20 border border-yellow-500/30 rounded-lg p-4">
        <div className="flex items-start gap-3">
          <AlertTriangle className="h-5 w-5 text-yellow-400 flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-yellow-400 text-sm font-medium mb-2">
              Disclaimer: Estimates Only
            </p>
            <p className="text-yellow-400/80 text-sm">
              Cost projections are estimates based on instance hourly rates at the time of
              stop/start events. Actual cloud provider billing may differ due to:
            </p>
            <ul className="text-yellow-400/70 text-sm mt-2 ml-4 list-disc space-y-1">
              <li>Reserved instance pricing or savings plans</li>
              <li>Data transfer and storage charges</li>
              <li>Taxes and regional pricing variations</li>
              <li>Promotional credits or discounts</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
