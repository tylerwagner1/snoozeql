import { useState, useEffect } from 'react'
import { TrendingDown } from 'lucide-react'
import api from '../lib/api'
import type { RecommendationEnriched, RecommendationGroup } from '../lib/api'
import { RecommendationsTable } from '../components/RecommendationsTable'
import { RecommendationModal } from '../components/RecommendationModal'
import toast from 'react-hot-toast'

const RecommendationsPage = () => {
  const [groups, setGroups] = useState<RecommendationGroup[]>([])
  const [dismissedCount, setDismissedCount] = useState(0)
  const [selectedRecommendation, setSelectedRecommendation] = useState<RecommendationEnriched | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [confirmLoading, setConfirmLoading] = useState(false)

  useEffect(() => {
    const fetchRecommendations = async () => {
      try {
        const [pendingResponse, dismissedResponse] = await Promise.all([
          api.getRecommendations('pending'),
          api.getRecommendations('dismissed')
        ])
        setGroups(pendingResponse?.groups || [])
        // Count dismissed from all groups
        const dismissedRecs = dismissedResponse?.groups?.flatMap(g => g.recommendations) || []
        setDismissedCount(dismissedRecs.length)
      } catch (err) {
        console.error('Failed to load recommendations:', err)
        toast.error('Failed to load recommendations')
      }
    }
    fetchRecommendations()
  }, [])

  const handleDismiss = async (ids: string | string[]) => {
    const idList = Array.isArray(ids) ? ids : [ids]
    try {
      await Promise.all(idList.map(id => api.dismissRecommendation(id)))
      // Remove from groups and update
      setGroups(prev => {
        const newGroups = prev.map(g => ({
          ...g,
          recommendations: g.recommendations.filter(r => !idList.includes(r.id)),
          instance_count: g.recommendations.filter(r => !idList.includes(r.id)).length,
          total_daily_savings: g.recommendations
            .filter(r => !idList.includes(r.id))
            .reduce((sum, r) => sum + r.estimated_daily_savings, 0)
        })).filter(g => g.instance_count > 0)
        return newGroups
      })
      // Only count as dismissed if not already dismissed
      const newDismissed = idList.length
      setDismissedCount(prev => prev + newDismissed)
      if (idList.length === 1) {
        toast.success('Recommendation dismissed')
      } else {
        toast.success(`${idList.length} recommendations dismissed`)
      }
    } catch (err) {
      toast.error('Failed to dismiss recommendations')
    }
  }

  const handleConfirm = async (id: string) => {
    setConfirmLoading(true)
    try {
      await api.confirmRecommendation(id)
      // Remove from groups
      setGroups(prev => {
        const newGroups = prev.map(g => ({
          ...g,
          recommendations: g.recommendations.filter(r => r.id !== id),
          instance_count: g.recommendations.filter(r => r.id !== id).length,
          total_daily_savings: g.recommendations
            .filter(r => r.id !== id)
            .reduce((sum, r) => sum + r.estimated_daily_savings, 0)
        })).filter(g => g.instance_count > 0)
        return newGroups
      })
      setModalOpen(false)
      setSelectedRecommendation(null)
      toast.success('Schedule created from recommendation!')
    } catch (err) {
      toast.error('Failed to create schedule')
    } finally {
      setConfirmLoading(false)
    }
  }

  const handleOpenModal = (rec: RecommendationEnriched) => {
    setSelectedRecommendation(rec)
    setModalOpen(true)
  }

  const pendingCount = groups.reduce((sum, g) => sum + g.instance_count, 0)

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Recommendations</h1>
          <p className="text-sm text-slate-400 mt-1">
            {pendingCount} pending · {dismissedCount} dismissed
          </p>
        </div>
      </div>

      {pendingCount > 0 ? (
        <RecommendationsTable
          groups={groups}
          onOpenModal={handleOpenModal}
          onDismiss={handleDismiss}
        />
      ) : (
        <div className="p-12 text-center text-slate-400">
          <div className="inline-block p-4 bg-slate-700/50 rounded-full mb-4">
            <TrendingDown className="h-12 w-12 text-slate-500" />
          </div>
          {dismissedCount > 0 ? (
            <p className="text-lg font-medium text-slate-300">No pending recommendations. {dismissedCount} dismissed recommendations.</p>
          ) : (
            <>
              <p className="text-lg font-medium text-slate-300">No pending recommendations</p>
              <p className="text-sm mt-2">Need 24+ hours of activity data to generate recommendations.</p>
            </>
          )}

        </div>
      )}

      <RecommendationModal
        isOpen={modalOpen}
        onClose={() => { setModalOpen(false); setSelectedRecommendation(null); }}
        recommendation={selectedRecommendation}
        onConfirm={handleConfirm}
        loading={confirmLoading}
      />
    </div>
  )
}

export default RecommendationsPage
