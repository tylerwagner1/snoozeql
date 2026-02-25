import { useState } from 'react'
import { ChevronDown, ChevronUp, Users } from 'lucide-react'
import type { RecommendationEnriched } from '../lib/api'

interface RecommendationGroupProps {
  group: {
    pattern_description: string
    pattern_key: string
    total_daily_savings: number
    instance_count: number
    recommendations: RecommendationEnriched[]
  }
  onOpenModal: (rec: RecommendationEnriched) => void
  onDismiss: (id: string) => void
}

export function RecommendationGroup({ group, onOpenModal, onDismiss }: RecommendationGroupProps) {
  const [expanded, setExpanded] = useState(true) // Start expanded

  // Single-instance groups render as regular cards
  if (group.instance_count === 1) {
    return (
      <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
        {group.recommendations.map((rec) => (
          <div key={rec.id} className="p-4">
            <RecommendationCard
              recommendation={rec}
              onOpenModal={onOpenModal}
              onDismiss={onDismiss}
            />
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="bg-slate-800/30 rounded-xl border border-slate-700 overflow-hidden">
      {/* Group header */}
      <div
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-slate-700/30 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-3">
          <Users className="h-5 w-5 text-slate-400" />
          <span className="font-medium text-white">{group.pattern_description}</span>
          <span className="text-sm text-slate-400">
            {group.instance_count} instance{group.instance_count !== 1 ? 's' : ''}
          </span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-green-400 font-semibold">
            ${group.total_daily_savings.toFixed(2)}/day total
          </span>
          {expanded ? (
            <ChevronUp className="h-4 w-4 text-slate-400" />
          ) : (
            <ChevronDown className="h-4 w-4 text-slate-400" />
          )}
        </div>
      </div>

      {/* Instance list */}
      {expanded && (
        <div className="px-4 pb-4 space-y-2">
          {group.recommendations.map((rec) => (
            <RecommendationCard
              key={rec.id}
              recommendation={rec}
              onOpenModal={onOpenModal}
              onDismiss={onDismiss}
            />
          ))}
        </div>
      )}
    </div>
  )
}

// Import RecommendationCard inline to avoid circular import
function RecommendationCard({
  recommendation,
  onOpenModal,
  onDismiss,
}: {
  recommendation: RecommendationEnriched
  onOpenModal: (rec: RecommendationEnriched) => void
  onDismiss: (id: string) => void
}) {
  const [expanded, setExpanded] = useState(false)

  // Determine confidence label and color
  const getConfidence = (score: number) => {
    if (score >= 80) return { label: 'High', color: 'bg-green-500/10 text-green-400 border-green-500/30' }
    if (score >= 50) return { label: 'Medium', color: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30' }
    return { label: 'Low', color: 'bg-orange-500/10 text-orange-400 border-orange-500/30' }
  }

  const confidence = getConfidence(recommendation.confidence_score)

  // Determine idle pattern text
  const getIdlePatternText = () => {
    const { idle_start_hour, idle_end_hour, days_of_week } = recommendation.detected_pattern

    // Format idle hours
    let patternText
    if (idle_start_hour < idle_end_hour) {
      // Normal case - no overnight
      patternText = `Idle from ${idle_start_hour}:00 to ${idle_end_hour}:00`
    } else {
      // Overnight case
      patternText = `Idle from ${idle_start_hour}:00 to ${idle_end_hour < 24 ? idle_end_hour : 0}:00 (overnight)`
    }

    // Format days of week
    if (days_of_week && days_of_week.length > 0) {
      patternText += ` on ${days_of_week.join(', ')}`
    } else {
      patternText += ` on all days`
    }

    return patternText
  }

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
      {/* Summary row (always visible) */}
      <div
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-slate-700/30 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-3 overflow-hidden">
          <span className="px-2 py-1 text-xs bg-purple-500/10 text-purple-400 border border-purple-500/30 rounded font-medium whitespace-nowrap">
            AI Suggested
          </span>
          <span className="font-medium text-white truncate max-w-[200px]">{recommendation.instance_name}</span>
          <span
            className={`${confidence.color} px-2 py-1 text-xs rounded border whitespace-nowrap`}
          >
            {confidence.label} confidence
          </span>
        </div>

        <div className="flex items-center gap-4">
          <span className="text-green-400 font-semibold whitespace-nowrap">
            ${recommendation.estimated_daily_savings.toFixed(2)}/day
          </span>
          {expanded ? <ChevronUp className="h-4 w-4 text-slate-400" /> : <ChevronDown className="h-4 w-4 text-slate-400" />}
        </div>
      </div>

      {/* Expanded section */}
      {expanded && (
        <div className="px-4 pb-4 border-t border-slate-700 space-y-3">
          {/* Activity pattern summary */}
          <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700">
            <p className="text-sm text-slate-400 mb-2">
              {getIdlePatternText()}
            </p>
          </div>

          {/* Suggested schedule */}
          <div className="grid grid-cols-2 gap-3">
            <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700">
              <span className="text-xs text-slate-400 uppercase">Sleep at</span>
              <p className="text-xs text-slate-300 mt-1 truncate">{recommendation.suggested_schedule.sleep_cron}</p>
            </div>
            <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700">
              <span className="text-xs text-slate-400 uppercase">Wake at</span>
              <p className="text-xs text-slate-300 mt-1 truncate">{recommendation.suggested_schedule.wake_cron}</p>
            </div>
          </div>

          {/* Action buttons */}
          <div className="flex gap-3 pt-2">
            <button
              onClick={(e) => {
                e.stopPropagation()
                onOpenModal(recommendation)
              }}
              className="flex-1 px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg text-sm font-medium transition-colors"
            >
              View Details
            </button>
            <button
              onClick={(e) => {
                e.stopPropagation()
                onDismiss(recommendation.id)
              }}
              className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-slate-300 hover:text-white rounded-lg text-sm font-medium transition-colors"
            >
              <ChevronUp className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default RecommendationGroup
