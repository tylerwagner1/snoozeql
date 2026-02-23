import { useState, useEffect } from 'react'
import { TrendingDown, RefreshCw } from 'lucide-react'
import api from '../lib/api'
import type { RecommendationEnriched } from '../lib/api'
import { RecommendationCard } from '../components/RecommendationCard'
import { RecommendationModal } from '../components/RecommendationModal'
import toast from 'react-hot-toast'

const RecommendationsPage = () => {
  const [recommendations, setRecommendations] = useState<RecommendationEnriched[]>([])
  const [dismissedCount, setDismissedCount] = useState(0)
  const [selectedRecommendation, setSelectedRecommendation] = useState<RecommendationEnriched | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [confirmLoading, setConfirmLoading] = useState(false)

  useEffect(() => {
    const fetchRecommendations = async () => {
      try {
        const [pending, dismissed] = await Promise.all([
          api.getRecommendations('pending'),
          api.getRecommendations('dismissed')
        ])
        setRecommendations(pending || [])
        setDismissedCount(dismissed?.length || 0)
      } catch (err) {
        console.error('Failed to load recommendations:', err)
        toast.error('Failed to load recommendations')
      }
    }
    fetchRecommendations()
  }, [])

  const handleGenerate = async () => {
    setGenerating(true)
    try {
      const result = await api.generateRecommendations()
      toast.success(result.message)
      const updated = await api.getRecommendations('pending')
      setRecommendations(updated || [])
    } catch (err) {
      toast.error('Failed to generate recommendations')
    } finally {
      setGenerating(false)
    }
  }

  const handleDismiss = async (id: string) => {
    try {
      await api.dismissRecommendation(id)
      setRecommendations(prev => prev.filter(r => r.id !== id))
      setDismissedCount(prev => prev + 1)
      toast.success('Recommendation dismissed')
    } catch (err) {
      toast.error('Failed to dismiss recommendation')
    }
  }

  const handleConfirm = async (id: string) => {
    setConfirmLoading(true)
    try {
      await api.confirmRecommendation(id)
      setRecommendations(prev => prev.filter(r => r.id !== id))
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

  const pending = recommendations.filter(r => r.status === 'pending')

  if (generating) {
    return <div className="p-8 text-center text-slate-400">Generating recommendations...</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Recommendations</h1>
          <p className="text-sm text-slate-400 mt-1">
            {pending.length} pending · {dismissedCount} dismissed
          </p>
        </div>
        <button
          onClick={handleGenerate}
          disabled={generating}
          className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white rounded-lg text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-blue-500/20"
        >
          {generating ? (
            <>
              <RefreshCw className="h-4 w-4 animate-spin" />
              <span>Generating...</span>
            </>
          ) : (
            <>
              <RefreshCw className="h-4 w-4" />
              <span>Generate Recommendations</span>
            </>
          )}
        </button>
      </div>

      {pending.length > 0 ? (
        <div className="space-y-4">
          {recommendations.map((recommendation) => (
            <RecommendationCard
              key={recommendation.id}
              recommendation={recommendation}
              onOpenModal={handleOpenModal}
              onDismiss={handleDismiss}
            />
          ))}
        </div>
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
          <button
            onClick={handleGenerate}
            disabled={generating}
            className="mt-4 flex items-center gap-2 mx-auto px-4 py-2 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white rounded-lg text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg"
          >
            {generating ? (
              <>
                <RefreshCw className="h-4 w-4 animate-spin" />
                <span>Generating...</span>
              </>
            ) : (
              <>
                <RefreshCw className="h-4 w-4" />
                <span>Generate Recommendations</span>
              </>
            )}
          </button>
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
