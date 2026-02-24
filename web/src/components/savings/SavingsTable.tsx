import { formatCurrency, formatHours } from '../../lib/formatters'
import type { InstanceSavingsItem } from '../../lib/api'

interface SavingsTableProps {
  data: InstanceSavingsItem[]
  loading: boolean
}

export function SavingsTable({ data, loading }: SavingsTableProps) {
  if (loading) {
    return (
      <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
        <div className="h-4 bg-slate-700 rounded w-48 mb-4 animate-pulse" />
        <div className="space-y-3">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="h-12 bg-slate-700/50 rounded animate-pulse" />
          ))}
        </div>
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="bg-slate-800/50 rounded-xl p-8 border border-slate-700 text-center">
        <p className="text-slate-400">No instance savings data yet.</p>
        <p className="text-sm text-slate-500 mt-2">
          Stop instances to start accumulating savings.
        </p>
      </div>
    )
  }

  return (
    <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
      <h2 className="text-lg font-semibold text-white mb-4">Top Savings</h2>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="text-left text-sm text-slate-400 border-b border-slate-700">
              <th className="pb-3 font-medium">Instance</th>
              <th className="pb-3 font-medium">Provider</th>
              <th className="pb-3 font-medium">Region</th>
              <th className="pb-3 font-medium text-right">Hours Stopped</th>
              <th className="pb-3 font-medium text-right">Savings</th>
            </tr>
          </thead>
          <tbody>
            {data.map((instance, index) => (
              <tr
                key={instance.instance_id}
                className="border-b border-slate-700/50 hover:bg-slate-700/30 transition-colors"
              >
                <td className="py-3 text-sm">
                  <div className="flex items-center gap-2">
                    <span className="text-slate-500 text-xs w-5">#{index + 1}</span>
                    <span className="text-white font-medium">{instance.name}</span>
                  </div>
                </td>
                <td className="py-3 text-sm text-slate-300">
                  <span className="px-2 py-0.5 bg-slate-700 rounded text-xs uppercase">
                    {instance.provider}
                  </span>
                </td>
                <td className="py-3 text-sm text-slate-300">{instance.region}</td>
                <td className="py-3 text-sm text-slate-300 text-right">
                  {formatHours(instance.stopped_hours)}
                </td>
                <td className="py-3 text-sm text-green-400 text-right font-medium">
                  {formatCurrency(instance.savings_cents)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
