import { useState, useEffect } from 'react'
import { Activity, Zap, TrendingDown, Clock, Search, Filter, Plus } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import api, { Instance, CloudAccount } from '../lib/api'
import type { Recommendation, Event } from '../lib/api'

const Dashboard = () => {
  const navigate = useNavigate()
  const [filter, setFilter] = useState('all')
  const [search, setSearch] = useState('')
  const [instances, setInstances] = useState<Instance[]>([])
  const [recommendations, setRecommendations] = useState<Recommendation[]>([])
  const [cloudAccounts, setCloudAccounts] = useState<CloudAccount[]>([])
  const [events, setEvents] = useState<Event[]>([])

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [instancesData, recommendationsData, accountsData, eventsData] = await Promise.all([
          api.getInstances(),
          api.getRecommendations(),
          api.getCloudAccounts(),
          api.getEvents(10, 0)
        ])
        // Handle null responses from API by defaulting to empty arrays
        setInstances(instancesData || [])
        setRecommendations(recommendationsData || [])
        setCloudAccounts(accountsData || [])
        setEvents(eventsData || [])
      } catch (err) {
        console.error(err)
        // Default to empty arrays on error
        setInstances([])
        setRecommendations([])
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

  const generateCostData = () => {
    const data = []
    const today = new Date()
    
    for (let day = 0; day < 7; day++) {
      const currentDay = new Date(today)
      currentDay.setDate(today.getDate() - (6 - day))
      
      for (let hour = 0; hour < 24; hour++) {
        const date = new Date(currentDay)
        date.setHours(hour)
        
        let hourlyCost = 0
        instances.forEach(inst => {
          if (hour >= 9 && hour < 17) {
            hourlyCost += inst.hourly_cost_cents
          } else if (hour >= 22 || hour < 7) {
            hourlyCost += 0
          } else {
            hourlyCost += inst.hourly_cost_cents * 0.2
          }
        })
        
        data.push({
          label: hour === 0 ? '12AM' : hour < 12 ? `${hour}AM` : hour === 12 ? '12PM' : `${hour - 12}PM`,
          cost: hourlyCost
        })
      }
    }
    
    return data
  }

  const costData = generateCostData()
  const maxCost = Math.max(...costData.map(d => d.cost), 100)
  const totalSavings = instances.reduce((sum, inst) => sum + (inst.hourly_cost_cents / 100) * 24 * 7, 0)
  const runningCount = filteredInstances.filter(i => i.status === 'available' || i.status === 'running' || i.status === 'starting').length
  const sleepingCount = filteredInstances.filter(i => i.status === 'stopped' || i.status === 'stopping').length
  const pendingActions = recommendations.length

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

      <div className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700">
        <h2 className="text-lg font-semibold text-white mb-4">Cost Over Time (7 days)</h2>
        <div className="h-64 w-full flex items-end space-x-1 sm:space-x-2">
          {costData.map((d, i) => {
            const heightPercentage = (d.cost / maxCost) * 100
            return (
              <div key={i} className="flex-1 flex flex-col justify-end group relative">
                <div 
                  className="bg-gradient-to-t from-blue-600 via-cyan-500 to-cyan-400 rounded-t-sm transition-all duration-300 hover:from-blue-500 hover:via-cyan-400 hover:to-cyan-300"
                  style={{ height: `${Math.max(heightPercentage, 0.5)}%` }}
                >
                  <div className="absolute -top-8 left-1/2 -translate-x-1/2 bg-slate-900 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap border border-slate-700 z-10">
                    ${Math.round(d.cost / 100)}/hr
                  </div>
                </div>
                {i % 24 === 0 && (
                  <div className="text-[10px] text-center text-slate-400 mt-2 truncate w-full">
                    {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'][i / 24]}
                  </div>
                )}
              </div>
            )
          })}
        </div>
        <div className="flex justify-between mt-4 text-xs text-slate-500 px-2">
          <span>00:00</span>
          <span>06:00</span>
          <span>12:00</span>
          <span>18:00</span>
          <span>24:00</span>
        </div>
      </div>

      <div className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700">
        <h2 className="text-lg font-semibold text-white mb-4">Recent Activity</h2>
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
                      {event.instance_id}
                    </span>
                    <span className="text-xs text-slate-500">
                      by {event.triggered_by === 'manual' ? 'User' : event.triggered_by}
                    </span>
                  </div>
                </div>
                <div className="flex flex-col items-end">
                  <span className={`text-sm font-medium ${
                    event.event_type === 'stop' ? 'text-red-400' : 'text-green-400'
                  }`}>
                    {event.event_type === 'stop' ? 'Stopped' : 'Started'}
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

      <div className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">Recommendations ({pendingActions} pending)</h2>
        </div>
        {recommendations.length > 0 ? (
          <div className="space-y-4">
            {recommendations.slice(0, 3).map((rec, i) => (
              <div key={i} className="p-5 bg-slate-900/50 border border-slate-700 rounded-lg hover:border-slate-600 transition-all">
                <div className="flex flex-col lg:flex-row items-start lg:items-center justify-between gap-4">
                  <div className="flex-1">
                    <p className="font-semibold text-white">{rec.instance_name}</p>
                    <p className="text-sm text-slate-400 mt-1">{rec.pattern}</p>
                    <div className="flex items-center space-x-3 mt-3">
                      <span className="text-sm font-semibold text-green-400">{rec.estimated_daily_savings}</span>
                      <span className="text-xs bg-gradient-to-r from-blue-600 to-purple-600 px-2.5 py-1 rounded text-white font-medium">
                        Confidence: {rec.confidence}%
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8">
            <div className="inline-block p-4 bg-slate-700/50 rounded-full mb-4">
              <Activity className="h-8 w-8 text-slate-500" />
            </div>
            <p className="text-sm text-slate-400">No recommendations pending</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default Dashboard
