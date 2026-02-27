import { Users, Eye, X } from 'lucide-react'
import type { RecommendationGroup } from '../lib/api'

interface RecommendationsTableProps {
  groups: RecommendationGroup[]
  onOpenModal: (rec: any) => void
  onDismiss: (ids: string | string[]) => Promise<void>
}

export function RecommendationsTable({ groups, onOpenModal, onDismiss }: RecommendationsTableProps) {
  const formatSavings = (amount: number) => {
    return `$${amount.toFixed(2)}/day`
  }

  const handleDismissGroup = async (group: RecommendationGroup) => {
    const ids = group.recommendations.map(r => r.id)
    await onDismiss(ids)
  }

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="bg-slate-900/50 border-b border-slate-700">
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                Schedule Pattern
              </th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                Wake/Sleep
              </th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                Instances Affected
              </th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                Est. Daily Savings
              </th>
              <th className="px-6 py-4 text-right text-xs font-semibold text-slate-400 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-700">
            {groups.map((group) => {
              // Get wake/sleep for display from first recommendation
              const firstRec = group.recommendations[0]
              const wakeTime = firstRec?.suggested_schedule?.wake_cron?.split(' ')[0]?.replace('0 ', '') || 'N/A'
              const sleepTime = firstRec?.suggested_schedule?.sleep_cron?.split(' ')[0]?.replace('0 ', '') || 'N/A'

              return (
                <tr
                  key={group.pattern_key}
                  className="hover:bg-slate-700/30 transition-colors group"
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-purple-500/10 rounded-lg">
                        <Users className="h-5 w-5 text-purple-400" />
                      </div>
                      <div>
                        <p className="font-medium text-white">{group.pattern_description}</p>
                        <p className="text-xs text-slate-500 mt-0.5">
                          {group.instance_count} instance{group.instance_count !== 1 ? 's' : ''}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center gap-2">
                        <span className="text-slate-400 text-xs uppercase">Wake</span>
                        <span className="text-slate-300 text-sm font-mono">{wakeTime}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-slate-400 text-xs uppercase">Sleep</span>
                        <span className="text-slate-300 text-sm font-mono">{sleepTime}</span>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-slate-700/50 text-slate-300 border border-slate-600">
                      {group.instance_count}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-green-400 font-semibold">
                      {formatSavings(group.total_daily_savings)}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="flex items-center justify-end gap-3">
                      <button
                        onClick={() => firstRec && onOpenModal(firstRec)}
                        className="flex items-center gap-2 px-3 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg text-sm font-medium transition-colors"
                      >
                        <Eye className="h-4 w-4" />
                        <span>View Details</span>
                      </button>
                      <button
                        onClick={() => handleDismissGroup(group)}
                        className="flex items-center gap-2 px-3 py-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 hover:text-red-300 rounded-lg text-sm font-medium transition-colors"
                      >
                        <X className="h-4 w-4" />
                        <span>Dismiss</span>
                      </button>
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default RecommendationsTable
