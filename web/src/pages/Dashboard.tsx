import { useState, useEffect, useMemo } from 'react'
import { Activity, Zap, TrendingDown, Clock, Search, Filter, Plus, RefreshCw } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import api, { Instance, CloudAccount } from '../lib/api'
import type { RecommendationEnriched, RecommendationGroup, Event } from '../lib/api'
import { RecommendationCard } from '../components/RecommendationCard'
import { RecommendationModal } from '../components/RecommendationModal'
import { CostOverTimeChart } from '../components/CostOverTimeChart'
import toast from 'react-hot-toast'

const Dashboard = () => {
  const navigate = useNavigate()
  const [filter, setFilter] = useState('all')
  const [search, setSearch] = useState('')
  const [instances, setInstances] = useState<Instance[]>([])
  const [groups, setGroups] = useState<RecommendationGroup[]>([])
  const [cloudAccounts, setCloudAccounts] = useState<CloudAccount[]>([])
  const [events, setEvents] = useState<Event[]>([])
  const [selectedRecommendation, setSelectedRecommendation] = useState<RecommendationEnriched | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [confirmLoading, setConfirmLoading] = useState(false)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [instancesData, recommendationsResponse, accountsData, eventsData] = await Promise.all([
          api.getInstances(),
          api.getRecommendations('pending'),
          api.getCloudAccounts(),
          api.getEvents(10, 0)
        ])
        setInstances(instancesData || [])
        setGroups(recommendationsResponse?.groups || [])
        setCloudAccounts(accountsData || [])
        setEvents(eventsData || [])
      } catch (err) {
        console.error(err)
        setInstances([])
        setGroups([])
        setCloudAccounts([])
        setEvents([])
      }
    }
    fetchData()
  }, [])

  const filteredInstances = instances.filter(instance => {
    const matchesSearch = instance.name.toLowerCase().includes(search.toLowerCase())
    const matchesFilter = filter === 'all' || 
      (filter === 'running' && (instance.status === 'available' || instance.status === 'running' || instance.status === 'starting')) ||
      (filter === 'sleeping' && (instance.status === 'stopped' || instance.status === 'stopping')) ||
      (filter === 'aws' && instance.provider === 'aws') ||
      (filter === 'gcp' && instance.provider === 'gcp')
    return matchesSearch && matchesFilter
  })

  const totalSavings = instances.reduce((sum, inst) => sum + (inst.hourly_cost_cents / 100) * 24 * 7, 0)
  const runningCount = filteredInstances.filter(i => i.status === 'available' || i.status === 'running' || i.status === 'starting').length
  const sleepingCount = filteredInstances.filter(i => i.status === 'stopped' || i.status === 'stopping').length
  const pendingActions = groups.reduce((sum, g) => sum + g.instance_count, 0)

  // Create instance name lookup map for events
  const instanceNameMap = useMemo(() => 
    new Map(instances.map(i => [i.id, i.name])),
    [instances]
  )

  const handleGenerate = async () => {
    setGenerating(true)
    try {
      const result = await api.generateRecommendations()
      toast.success(result.message)
      const updated = await api.getRecommendations('pending')
      setGroups(updated?.groups || [])
    } catch (err: any) {
      const errorMessage = err?.response?.data?.message || err?.response?.data?.error || err?.message || 'Failed to generate recommendations'
      toast.error(errorMessage)
    } finally {
      setGenerating(false)
    }
  }

  const handleDismiss = async (id: string) => {
    try {
      await api.dismissRecommendation(id)
      // Remove from groups and update
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
      toast.success('Recommendation dismissed')
    } catch (err) {
      toast.error('Failed to dismiss recommendation')
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

  return (
    <div className="space-y-6">
      <div className="flex flex-col lg:flex-row items-start lg:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Dashboard</h1>
          <p className="text-sm text-slate-400 mt-1">
            Last updated: Just now
          </p>
        </div>
        <div className="flex flex-col sm:flex-row gap-3 w-full lg:w-auto">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              type="text"
              placeholder="Search instances..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9 pr-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 w-full sm:w-64"
            />
          </div>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500 min-w-[140px]"
          >
            <option value="all">All Instances</option>
            <option value="running">Running Only</option>
            <option value="sleeping">Sleeping Only</option>
            <option value="aws">AWS Only</option>
            <option value="gcp">GCP Only</option>
          </select>
          <div className="flex items-center gap-2 px-3 py-2 bg-slate-800 rounded-lg border border-slate-700 text-sm">
            <Filter className="h-4 w-4 text-slate-400" />
            <span className="text-slate-300">{filteredInstances.length} total</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-slate-800/50 rounded-xl p-5 shadow-lg border border-slate-700 hover:border-green-500/50 transition-all group">
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm text-slate-400 font-medium">Total Savings</p>
            <div className="p-2 bg-gradient-to-br from-green-500 to-emerald-600 rounded-lg group-hover:scale-105 transition-transform shadow-lg shadow-green-500/20">
              <TrendingDown className="h-5 w-5 text-white" />
            </div>
          </div>
          <p className="text-3xl font-bold text-white mb-1">${totalSavings.toFixed(2)}</p>
          <p className="text-sm text-green-400">+12% vs last week</p>
        </div>
        <div 
          className="bg-slate-800/50 rounded-xl p-5 shadow-lg border border-slate-700 hover:border-blue-500/50 transition-all group cursor-pointer"
          onClick={() => navigate('/instances?status=running')}
        >
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm text-slate-400 font-medium">Running Databases</p>
            <div className="p-2 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-lg group-hover:scale-105 transition-transform shadow-lg shadow-blue-500/20">
              <Activity className="h-5 w-5 text-white" />
            </div>
          </div>
          <p className="text-3xl font-bold text-white mb-1">{runningCount}</p>
          <p className="text-sm text-blue-400">Active & processing</p>
        </div>
        <div 
          className="bg-slate-800/50 rounded-xl p-5 shadow-lg border border-slate-700 hover:border-purple-500/50 transition-all group cursor-pointer"
          onClick={() => navigate('/instances?status=stopped')}
        >
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm text-slate-400 font-medium">Sleeping Databases</p>
            <div className="p-2 bg-gradient-to-br from-purple-500 to-pink-600 rounded-lg group-hover:scale-105 transition-transform shadow-lg shadow-purple-500/20">
              <Zap className="h-5 w-5 text-white" />
            </div>
          </div>
          <p className="text-3xl font-bold text-white mb-1">{sleepingCount}</p>
          <p className="text-sm text-purple-400">Cost optimized</p>
        </div>
        <div 
          className="bg-slate-800/50 rounded-xl p-5 shadow-lg border border-slate-700 hover:border-yellow-500/50 transition-all group cursor-pointer"
          onClick={() => navigate('/recommendations')}
        >
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm text-slate-400 font-medium">Pending Actions</p>
            <div className="p-2 bg-gradient-to-br from-yellow-500 to-orange-600 rounded-lg group-hover:scale-105 transition-transform shadow-lg shadow-yellow-500/20">
              <Clock className="h-5 w-5 text-white" />
            </div>
          </div>
          <p className="text-3xl font-bold text-white mb-1">{pendingActions}</p>
          <p className="text-sm text-yellow-400">Needs attention</p>
        </div>
      </div>

      {cloudAccounts.length === 0 && (
        <div className="bg-gradient-to-r from-blue-900/50 to-purple-900/50 rounded-xl p-6 border border-blue-500/30">
          <h2 className="text-xl font-semibold text-white mb-2">Get Started</h2>
          <p className="text-slate-300 mb-4">Connect your cloud accounts to discover and manage database instances.</p>
          <div className="flex gap-3">
            <button
              onClick={() => navigate('/cloud-accounts')}
              className="px-4 py-2 bg-gradient-to-r from-yellow-500 to-orange-600 text-white font-medium rounded-lg hover:from-yellow-400 hover:to-orange-500 transition-all shadow-lg"
            >
              Add AWS Account
            </button>
            <button
              onClick={() => navigate('/cloud-accounts')}
              className="px-4 py-2 bg-gradient-to-r from-blue-500 to-indigo-600 text-white font-medium rounded-lg hover:from-blue-400 hover:to-indigo-500 transition-all shadow-lg"
            >
              Add GCP Account
            </button>
          </div>
        </div>
      )}

      {cloudAccounts.length > 0 && (
        <div className="flex gap-4">
          <button
            onClick={() => navigate('/cloud-accounts')}
            className="text-sm text-blue-400 hover:text-blue-300 flex items-center gap-1"
          >
            <Plus className="h-4 w-4" />
            Add Account
          </button>
          <button
            onClick={() => navigate('/instances')}
            className="text-sm text-slate-400 hover:text-slate-300 flex items-center gap-1"
          >
            View All Instances →
          </button>
        </div>
      )}

      <CostOverTimeChart instances={instances} />

      <div className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700">
        <h2 className="text-lg font-semibold text-white mb-4">Recent Activity</h2>
        <div className="max-h-80 overflow-y-auto">
          {events.length > 0 ? (
            <div className="space-y-4">
              {events.map((event, i) => (
                <div key={i} className="flex items-center justify-between py-3 border-b border-slate-700 last:border-0 hover:bg-slate-800/50 px-2 rounded transition-colors">
                  <div className="flex items-center space-x-3">
                    <span className="text-sm text-slate-400 w-20">
                      {new Date(event.created_at).toLocaleString()}
                    </span>
                    <div className="flex flex-col">
                      <span className="text-sm text-slate-200 font-medium">
                        {instanceNameMap.get(event.instance_id) || event.instance_id.slice(0, 8) + '...'}
                      </span>
                      <span className="text-xs text-slate-500">
                        by {event.triggered_by === 'manual' ? 'User' : event.triggered_by}
                      </span>
                    </div>
                  </div>
                  <div className="flex flex-col items-end">
                    <span className={`text-sm font-medium ${
                      ['stop', 'sleep'].includes(event.event_type) ? 'text-red-400' : 'text-green-400'
                    }`}>
                      {['stop', 'sleep'].includes(event.event_type) ? 'Stopped' : 'Started'}
                    </span>
                    <span className="text-xs text-slate-500">
                      {event.previous_status} → {event.new_status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : instances.length > 0 ? (
            <div className="space-y-4">
              {instances.slice(0, 4).map((instance, i) => (
                <div key={i} className="flex items-center justify-between py-3 border-b border-slate-700 last:border-0 hover:bg-slate-800/50 px-2 rounded transition-colors">
                  <div className="flex items-center space-x-3">
                    <span className="text-sm text-slate-400 w-20">No activity yet</span>
                    <span className="text-sm text-slate-200 font-medium">{instance.name} ({instance.engine})</span>
                  </div>
                  <span className="text-sm text-slate-500">${(instance.hourly_cost_cents / 100).toFixed(2)}/hr</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <p className="text-sm text-slate-400">No instances discovered yet</p>
            </div>
          )}
        </div>
      </div>


      <div className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">AI Recommendations</h2>
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
                <span>Generate</span>
              </>
            )}
          </button>
        </div>
        
        {pendingActions > 0 ? (
          <div className="space-y-4">
            {/* Show first group's recommendations (up to 3) */}
            {groups.length > 0 ? groups[0].recommendations.slice(0, 3).map((rec: RecommendationEnriched) => (
              <RecommendationCard
                key={rec.id}
                recommendation={rec}
                onOpenModal={handleOpenModal}
                onDismiss={handleDismiss}
              />
            )) : null}
            <div className="text-center pt-2">
              <button
                onClick={() => navigate('/recommendations')}
                className="text-sm text-blue-400 hover:text-blue-300 font-medium"
              >
                View all recommendations →
              </button>
            </div>
          </div>
        ) : (
          <div className="text-center py-8">
            <p className="text-sm text-slate-400">Need 24+ hours of activity data to generate recommendations.</p>
            <button
              onClick={handleGenerate}
              disabled={generating}
              className="mt-4 flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white rounded-lg text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg"
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
      </div>

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

export default Dashboard
