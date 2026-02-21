import { useState, useEffect, useMemo } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import api from '../lib/api'
import type { Instance } from '../lib/api'

const InstancesPage = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  
  const [instances, setInstances] = useState<Instance[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sortConfig, setSortConfig] = useState<{
    key: keyof Instance
    direction: 'asc' | 'desc'
  }>({ key: 'name', direction: 'asc' })
  const [filters, setFilters] = useState<{
    status: string
    provider: string
  }>({
    status: searchParams.get('status') || 'all',
    provider: searchParams.get('provider') || 'all'
  })

  const updateFilter = (key: 'status' | 'provider', value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }))
    const newParams = new URLSearchParams(searchParams)
    if (value === 'all') {
      newParams.delete(key)
    } else {
      newParams.set(key, value)
    }
    setSearchParams(newParams)
  }

  useEffect(() => {
    const fetchInstances = async () => {
      try {
        const data = await api.getInstances()
        setInstances(data)
      } catch (err) {
        setError('Failed to load instances')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    fetchInstances()
  }, [])

  const handleStart = async (id: string) => {
    try {
      await api.startInstance(id)
      setInstances(prev => prev.map(inst => inst.id === id ? { ...inst, status: 'starting' } : inst))
    } catch (err) {
      console.error('Failed to start instance:', err)
    }
  }

  const handleStop = async (id: string) => {
    try {
      await api.stopInstance(id)
      setInstances(prev => prev.map(inst => inst.id === id ? { ...inst, status: 'stopping' } : inst))
    } catch (err) {
      console.error('Failed to stop instance:', err)
    }
  }

  const handleSort = (key: keyof Instance) => {
    setSortConfig(prev => ({
      key,
      direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
    }))
  }

  const filteredAndSortedInstances = useMemo(() => {
    let filtered = instances
    
    if (filters.status !== 'all') {
      // Map URL status values to filter values
      // 'running' -> filter for 'available', 'running', 'starting'
      // 'stopped' -> filter for 'stopped', 'stopping'
      let statusFilter: string[] = []
      if (filters.status === 'running') {
        statusFilter = ['available', 'running', 'starting']
      } else if (filters.status === 'stopped') {
        statusFilter = ['stopped', 'stopping']
      } else {
        statusFilter = [filters.status]
      }
      filtered = filtered.filter(i => statusFilter.includes(i.status))
    }
    if (filters.provider !== 'all') {
      filtered = filtered.filter(i => i.provider.startsWith(filters.provider))
    }
    
    return [...filtered].sort((a, b) => {
      const aVal = a[sortConfig.key] ?? ''
      const bVal = b[sortConfig.key] ?? ''
      if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1
      if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1
      return 0
    })
  }, [instances, filters, sortConfig])

  if (loading) return <div className="p-8 text-center text-slate-400">Loading instances...</div>
  if (error) return <div className="p-8 text-center text-red-400">{error}</div>

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Database Instances</h1>
          <p className="text-sm text-slate-400 mt-1">
            {filteredAndSortedInstances.length} of {instances.length} instances
          </p>
        </div>
      </div>

      <div className="flex gap-4 mb-4">
        <select
          value={filters.status}
          onChange={(e) => updateFilter('status', e.target.value)}
          className="px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-white"
        >
          <option value="all">All Status</option>
          <option value="available">Running</option>
          <option value="stopped">Stopped</option>
          <option value="starting">Starting</option>
          <option value="stopping">Stopping</option>
        </select>
        
        <select
          value={filters.provider}
          onChange={(e) => updateFilter('provider', e.target.value)}
          className="px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-white"
        >
          <option value="all">All Providers</option>
          <option value="aws">AWS</option>
          <option value="gcp">GCP</option>
        </select>
      </div>

      <div className="bg-slate-800/50 shadow-lg border border-slate-700 rounded-xl overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-slate-900/50">
            <tr>
              <th 
                className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider cursor-pointer hover:text-white"
                onClick={() => handleSort('name')}
              >
                Name {sortConfig.key === 'name' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
              </th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Provider</th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Region</th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Engine</th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Type</th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
              <th className="px-6 py-4 text-right text-xs font-semibold text-slate-400 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-slate-800/30 divide-y divide-slate-700">
            {filteredAndSortedInstances.map(instance => (
              <tr key={instance.id} className="hover:bg-slate-700/50 transition-colors">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center">
                    <div className="flex-shrink-0 h-10 w-10 bg-gradient-to-br from-blue-600 to-cyan-700 rounded-lg flex items-center justify-center shadow-lg shadow-blue-500/20">
                      <span className="text-white font-bold truncate max-w-[30px]">
                        {instance.name.substring(0, 2).toUpperCase()}
                      </span>
                    </div>
                    <div className="ml-4">
                      <Link to={`/instances/${instance.id}`} className="text-sm font-medium text-white hover:text-blue-400 transition-colors">
                        {instance.name}
                      </Link>
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2.5 py-1 text-xs rounded-full font-medium ${
                    instance.provider.startsWith('aws') ? 'bg-green-500/10 text-green-400 border border-green-500/30' : 'bg-blue-500/10 text-blue-400 border border-blue-500/30'
                  }`}>
                    {instance.provider.startsWith('aws') ? 'AWS' : 'GCP'}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-300">{instance.region}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-white font-medium">{instance.engine}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-300">{instance.instance_type}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2.5 py-1 text-xs rounded-full font-medium capitalize ${
                    instance.status === 'running' || instance.status === 'starting' ? 'bg-green-500/10 text-green-400 border border-green-500/30' :
                    instance.status === 'stopped' ? 'bg-slate-500/10 text-slate-400 border border-slate-500/30' :
                    instance.status === 'stopping' ? 'bg-orange-500/10 text-orange-400 border border-orange-500/30 animate-pulse' :
                    'bg-yellow-500/10 text-yellow-400 border border-yellow-500/30'
                  }`}>
                    {instance.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right">
                  <div className="flex justify-end space-x-2">
                    {instance.status === 'stopped' ? (
                      <button
                        onClick={() => handleStart(instance.id)}
                        className="px-3 py-1.5 bg-green-600 hover:bg-green-500 text-white text-xs font-medium rounded-lg transition-all shadow-lg shadow-green-500/20"
                      >
                        Start
                      </button>
                    ) : (
                      <button
                        onClick={() => handleStop(instance.id)}
                        className="px-3 py-1.5 bg-red-600 hover:bg-red-500 text-white text-xs font-medium rounded-lg transition-all shadow-lg shadow-red-500/20"
                      >
                        Stop
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {instances.length === 0 && (
          <div className="p-12 text-center text-slate-400">
            <div className="inline-block p-4 bg-slate-700/50 rounded-full mb-4">
              <svg className="h-12 w-12 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" />
              </svg>
            </div>
            <p className="text-lg font-medium text-slate-300">No instances found</p>
            <p className="text-sm mt-2">Try discovering your database instances</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default InstancesPage
