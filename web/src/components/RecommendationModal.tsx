import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react'
import { X, Clock, TrendingDown } from 'lucide-react'
import clsx from 'clsx'
import type { RecommendationEnriched } from '../lib/api'
import { ActivityGraph } from './ActivityGraph'

interface RecommendationModalProps {
  isOpen: boolean
  onClose: () => void
  recommendation: RecommendationEnriched | null
  onConfirm: (id: string) => Promise<void>
  loading?: boolean
}

export function RecommendationModal({
  isOpen,
  onClose,
  recommendation,
  onConfirm,
  loading,
}: RecommendationModalProps) {
  if (!recommendation) return null

  // Determine confidence label
  const getConfidenceLabel = (score: number) => {
    if (score >= 80) return 'High'
    if (score >= 50) return 'Medium'
    return 'Low'
  }

  const confidenceLabel = getConfidenceLabel(recommendation.confidence_score)

  // Get idle window pattern description
  const formatHour = (hour: number) => {
    const period = hour >= 12 ? 'PM' : 'AM';
    const displayHour = hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
    return `${displayHour}:00 ${period}`;
  }

  const getPatternDescription = () => {
    const { idle_start_hour, idle_end_hour } = recommendation.detected_pattern;
    
    const wakeTime = formatHour(idle_end_hour);
    const sleepTime = formatHour(idle_start_hour);
    
    return `Detected low utilization outside ${wakeTime} – ${sleepTime}`;
  }

  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <DialogBackdrop
        transition
        className="fixed inset-0 bg-black/50 backdrop-blur-sm duration-200 ease-out data-[closed]:opacity-0"
      />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel
          transition
          className="max-w-lg w-full bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-2xl duration-200 ease-out data-[closed]:scale-95 data-[closed]:opacity-0"
        >
          <div className="flex items-center justify-between mb-6">
            <DialogTitle className="text-xl font-semibold text-white">
              Confirm Schedule Recommendation
            </DialogTitle>
            <button
              onClick={onClose}
              disabled={loading}
              className="text-slate-400 hover:text-white transition-colors disabled:opacity-50"
            >
              <X className="h-6 w-6" />
            </button>
          </div>

          {/* Instance info */}
          <div className="mb-4">
            <span className="px-2 py-1 text-xs bg-purple-500/10 text-purple-400 border border-purple-500/30 rounded mr-2">
              AI Suggested
            </span>
            <span className="text-lg font-semibold text-white">{recommendation.instance_name}</span>
          </div>

          {/* Activity pattern section */}
          <div className="mb-6 p-4 bg-slate-900/50 rounded-lg border border-slate-700">
            <h3 className="text-sm font-medium text-slate-300 mb-2">Detected Activity Pattern</h3>
            <ActivityGraph pattern={recommendation.detected_pattern} />
            <p className="text-xs text-slate-400 mt-2 text-center">
              {getPatternDescription()}
            </p>
          </div>

          {/* Suggested schedule */}
          <div className="mb-6 space-y-3">
            <h3 className="text-sm font-medium text-slate-300">Suggested Schedule</h3>
            <div className="grid grid-cols-2 gap-3">
              <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700">
                <div className="flex items-center gap-2 text-slate-400 mb-1">
              <Clock className="h-4 w-4" />
                <span className="text-xs uppercase">Wake at</span>
                </div>
                <p className="text-white font-mono text-sm">{recommendation.suggested_schedule.wake_cron}</p>
                <p className="text-xs text-slate-500 mt-1">End idle period</p>
              </div>
              <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700">
                <div className="flex items-center gap-2 text-slate-400 mb-1">
              <Clock className="h-4 w-4" />
                <span className="text-xs uppercase">Sleep at</span>
                </div>
                <p className="text-white font-mono text-sm">{recommendation.suggested_schedule.sleep_cron}</p>
                <p className="text-xs text-slate-500 mt-1">Start idle period</p>
              </div>
            </div>
          </div>

          {/* Savings summary */}
          <div className="mb-6 p-4 bg-green-500/10 rounded-lg border border-green-500/30">
            <div className="flex items-center gap-3">
              <TrendingDown className="h-5 w-5 text-green-400" />
              <div className="flex-1">
                <p className="text-sm text-green-400">Estimated daily savings</p>
                <p className="text-lg font-bold text-green-400">
                  ${recommendation.estimated_daily_savings.toFixed(2)}/day
                </p>
              </div>
            </div>
          </div>

          {/* Confidence text */}
          <div className="mb-6 text-sm text-slate-400">
            <span className="font-medium">{confidenceLabel} confidence</span> - based on recent activity patterns
          </div>

          {/* Action buttons */}
          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => onConfirm(recommendation.id)}
              disabled={loading}
              className={clsx(
                'px-4 py-2 text-white rounded-lg transition-all shadow-lg',
                'bg-green-600 hover:bg-green-500 shadow-green-500/20',
                'disabled:opacity-50 disabled:shadow-none'
              )}
            >
              {loading ? 'Creating schedule...' : 'Create Schedule'}
            </button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  )
}

export default RecommendationModal
