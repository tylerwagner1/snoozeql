import { useState, useEffect } from 'react'
import { TrendingDown, Timer, Activity } from 'lucide-react'
import api from '../lib/api'
import type { Recommendation } from '../lib/api'

const RecommendationsPage = () => {
  const [recommendations, setRecommendations] = useState<Recommendation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchRecommendations = async () => {
      try {
        const data = await api.getRecommendations()
        setRecommendations(data)
      } catch (err) {
        setError('Failed to load recommendations')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    fetchRecommendations()
  }, [])

  const handleApply = async (id: string) => {
    try {
      await api.applyRecommendation(id)
      setRecommendations(prev => prev.map(rec => rec.id === id ? { ...rec, status: 'applied' } : rec))
    } catch (err) {
      console.error('Failed to apply recommendation:', err)
    }
  }

  const handleIgnore = async (id: string) => {
    try {
      await api.ignoreRecommendation(id)
      setRecommendations(prev => prev.map(rec => rec.id === id ? { ...rec, status: 'ignored' } : rec))
    } catch (err) {
      console.error('Failed to ignore recommendation:', err)
    }
  }

  if (loading) return <div className="p-8 text-center text-slate-400">Loading recommendations...</div>
  if (error) return <div className="p-8 text-center text-red-400">{error}</div>

  const pending = recommendations.filter(r => r.status === 'pending')
  const applied = recommendations.filter(r => r.status === 'applied')
  const ignored = recommendations.filter(r => r.status === 'ignored')

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Recommendations</h1>
          <p className="text-sm text-slate-400 mt-1">{recommendations.length} total recommendations</p>
        </div>
        <div className="flex space-x-2">
          <span className="px-3 py-1.5 bg-blue-500/10 text-blue-400 border border-blue-500/30 rounded-full text-sm font-medium">
            Pending: {pending.length}
          </span>
          <span className="px-3 py-1.5 bg-green-500/10 text-green-400 border border-green-500/30 rounded-full text-sm font-medium">
            Applied: {applied.length}
          </span>
          <span className="px-3 py-1.5 bg-slate-500/10 text-slate-400 border border-slate-500/30 rounded-full text-sm font-medium">
            Ignored: {ignored.length}
          </span>
        </div>
      </div>

      <div className="space-y-4">
        {recommendations.map(recommendation => (
          <div key={recommendation.id} className={`bg-slate-800/50 shadow-lg border border-slate-700 rounded-xl p-6 transition-all ${
            recommendation.status === 'applied' ? 'opacity-60' : 'hover:border-slate-600'
          }`}>
            <div className="flex items-start justify-between mb-4">
              <div className="flex-1">
                <div className="flex items-center space-x-3 mb-2">
                  <h3 className="text-lg font-semibold text-white">{recommendation.instance_name}</h3>
                  <span className={`px-2.5 py-1 text-xs rounded-full border ${
                    recommendation.status === 'pending' ? 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30' :
                    recommendation.status === 'applied' ? 'bg-green-500/10 text-green-400 border-green-500/30' :
                    'bg-slate-500/10 text-slate-400 border-slate-500/30'
                  }`}>
                    {recommendation.status}
                  </span>
                </div>
                <div className="flex items-center text-sm text-slate-300 mb-3">
                  <Timer className="h-4 w-4 mr-2 text-slate-500" />
                  <span className="font-mono">{recommendation.pattern}</span>
                </div>
                
                <div className="flex flex-col sm:flex-row items-start sm:items-center space-y-3 sm:space-y-0 sm:space-x-6">
                  <div className="flex items-center space-x-2">
                    <TrendingDown className="h-4 w-4 text-green-500" />
                    <span className="text-xs text-slate-400 uppercase">Estimated Savings</span>
                    <span className="text-sm font-bold text-green-400">{recommendation.estimated_daily_savings}</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Activity className="h-4 w-4 text-slate-500" />
                    <span className="text-xs text-slate-400 uppercase">Confidence</span>
                    <div className="flex items-center space-x-2">
                      <div className="w-24 h-2 bg-slate-700 rounded-full overflow-hidden">
                        <div 
                          className={`h-full rounded-full ${
                            recommendation.confidence >= 80 ? 'bg-green-500' :
                            recommendation.confidence >= 60 ? 'bg-yellow-500' : 'bg-orange-500'
                          }`}
                          style={{ width: `${recommendation.confidence}%` }}
                        />
                      </div>
                      <span className="text-xs text-slate-300">{recommendation.confidence}%</span>
                    </div>
                  </div>
                </div>

                <div className="mt-4 p-4 bg-slate-900/50 border border-slate-700 rounded-lg">
                  <p className="text-xs text-slate-400 uppercase mb-2">Activity Pattern</p>
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-slate-300">Active hours:</span>
                      <span className="font-mono text-white">
                        {recommendation.activity_pattern.active_hours.length > 0
                          ? recommendation.activity_pattern.active_hours.join(', ')
                          : 'None detected'}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-slate-300">Inactive hours:</span>
                      <span className="font-mono text-white">
                        {recommendation.activity_pattern.inactive_hours.length > 0
                          ? recommendation.activity_pattern.inactive_hours.join(', ')
                          : 'None detected'}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="ml-4 flex flex-col space-y-2">
                {recommendation.status === 'pending' && (
                  <>
                    <button
                      onClick={() => handleApply(recommendation.id)}
                      className="px-4 py-2 bg-gradient-to-r from-green-600 to-green-700 border border-transparent rounded-lg text-sm font-medium text-white hover:from-green-500 hover:to-green-600 shadow-lg shadow-green-500/20 transition-all"
                    >
                      Apply
                    </button>
                    <button
                      onClick={() => handleIgnore(recommendation.id)}
                      className="px-4 py-2 bg-gradient-to-r from-red-600 to-red-700 border border-transparent rounded-lg text-sm font-medium text-white hover:from-red-500 hover:to-red-600 shadow-lg shadow-red-500/20 transition-all"
                    >
                      Ignore
                    </button>
                  </>
                )}
                {recommendation.status === 'applied' && (
                  <span className="px-4 py-2 bg-green-500/10 text-green-400 border border-green-500/30 rounded-lg text-sm font-medium text-center">
                    Applied
                  </span>
                )}
                {recommendation.status === 'ignored' && (
                  <span className="px-4 py-2 bg-slate-500/10 text-slate-400 border border-slate-500/30 rounded-lg text-sm font-medium text-center">
                    Ignored
                  </span>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {recommendations.length === 0 && (
        <div className="p-12 text-center text-slate-400">
          <div className="inline-block p-4 bg-slate-700/50 rounded-full mb-4">
            <TrendingDown className="h-12 w-12 text-slate-500" />
          </div>
          <p className="text-lg font-medium text-slate-300">No recommendations available</p>
          <p className="text-sm mt-2">Check back later for AI-powered suggestions</p>
        </div>
      )}
    </div>
  )
}

export default RecommendationsPage
