import { useState, useEffect, useMemo } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import toast from 'react-hot-toast'
import api from '../lib/api'
import type { Instance } from '../lib/api'
import { ConfirmDialog } from '../components/ConfirmDialog'

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

  // Selection state
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [showConfirmDialog, setShowConfirmDialog] = useState<'sleep' | 'wake' | null>(null)
  const [bulkLoading, setBulkLoading] = useState(false)

  // Selection handlers - toggleSelect and clearSelection don't depend on filteredAndSortedInstances
  const toggleSelect = (id: string) => {
    setSelectedIds(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const clearSelection = () => {
    setSelectedIds(new Set())
  }

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

  // Fetch instances from API
  const fetchInstances = async () => {
    try {
      const data = await api.getInstances()
      setInstances(data)
      setError(null)
    } catch (err) {
      setError('Failed to load instances')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchInstances()
  }, [])

  // Refresh instances every 5 seconds to show real-time status
  useEffect(() => {
    const interval = setInterval(() => {
      fetchInstances()
    }, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleStart = async (id: string) => {
    try {
      await api.startInstance(id)
      toast.success('Started instance')
      // Will be updated by 5-second refresh
    } catch (err: any) {
      console.error('Failed to start instance:', err)
      toast.error(err.message || 'Failed to start instance')
    }
  }

  const handleStop = async (id: string) => {
    try {
      await api.stopInstance(id)
      toast.success('Stopped instance')
      // Will be updated by 5-second refresh
    } catch (err: any) {
      console.error('Failed to stop instance:', err)
      toast.error(err.message || 'Failed to stop instance')
    }
  }

  // Bulk operation handlers
  const handleBulkSleep = async () => {
    setBulkLoading(true)
    try {
      const idsToStop = stoppableSelected.map(i => i.id)
      const result = await api.bulkStopInstances(idsToStop)
      
      if (result.success.length > 0) {
        toast.success(`Stopped ${result.success.length} instance(s)`)
      }
      if (result.failed.length > 0) {
        toast.error(`Failed to stop ${result.failed.length} instance(s)`)
      }
      
      clearSelection()
      setShowConfirmDialog(null)
      // Will be updated by 5-second refresh
    } catch (err: any) {
      toast.error(`Failed to stop instances: ${err.message || 'unknown error'}`)
      console.error(err)
    } finally {
      setBulkLoading(false)
    }
  }

  const handleBulkWake = async () => {
    setBulkLoading(true)
    try {
      const idsToStart = startableSelected.map(i => i.id)
      const result = await api.bulkStartInstances(idsToStart)
      
      if (result.success.length > 0) {
        toast.success(`Started ${result.success.length} instance(s)`)
      }
      if (result.failed.length > 0) {
        toast.error(`Failed to start ${result.failed.length} instance(s)`)
      }
      
      clearSelection()
      setShowConfirmDialog(null)
      // Will be updated by 5-second refresh
    } catch (err: any) {
      toast.error(`Failed to start instances: ${err.message || 'unknown error'}`)
      console.error(err)
    } finally {
      setBulkLoading(false)
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

  // Computed selection values that depend on filteredAndSortedInstances
  const selectAll = () => {
    setSelectedIds(new Set(filteredAndSortedInstances.map(i => i.id)))
  }

  const allSelected = filteredAndSortedInstances.length > 0 && 
    filteredAndSortedInstances.every(i => selectedIds.has(i.id))

  // Get selected instances for dialog
  const selectedInstances = instances.filter(i => selectedIds.has(i.id))
  const stoppableSelected = selectedInstances.filter(i => 
    i.status === 'available' || i.status === 'running'
  )
  const startableSelected = selectedInstances.filter(i => i.status === 'stopped')

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

      {/* Bulk action buttons */}
      {selectedIds.size > 0 && (
        <div className="flex items-center gap-3 bg-slate-800/50 px-4 py-2 rounded-lg border border-slate-700">
          <span className="text-sm text-slate-400">{selectedIds.size} selected</span>
          <button
            onClick={() => setShowConfirmDialog('sleep')}
            disabled={stoppableSelected.length === 0}
            className="px-3 py-1.5 bg-yellow-600 hover:bg-yellow-500 disabled:bg-slate-600 disabled:cursor-not-allowed text-white text-xs font-medium rounded-lg transition-all shadow-lg shadow-yellow-500/20"
          >
            Sleep Selected ({stoppableSelected.length})
          </button>
          <button
            onClick={() => setShowConfirmDialog('wake')}
            disabled={startableSelected.length === 0}
            className="px-3 py-1.5 bg-green-600 hover:bg-green-500 disabled:bg-slate-600 disabled:cursor-not-allowed text-white text-xs font-medium rounded-lg transition-all shadow-lg shadow-green-500/20"
          >
            Wake Selected ({startableSelected.length})
          </button>
          <button
            onClick={clearSelection}
            className="px-3 py-1.5 bg-slate-700 hover:bg-slate-600 text-white text-xs font-medium rounded-lg transition-all"
          >
            Clear
          </button>
        </div>
      )}

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
              <th className="px-6 py-4 text-left">
                <input
                  type="checkbox"
                  checked={allSelected}
                  onChange={(e) => e.target.checked ? selectAll() : clearSelection()}
                  className="rounded border-slate-600 bg-slate-700 text-blue-500 focus:ring-blue-500 focus:ring-offset-slate-800"
                />
              </th>
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
                  <input
                    type="checkbox"
                    checked={selectedIds.has(instance.id)}
                    onChange={() => toggleSelect(instance.id)}
                    className="rounded border-slate-600 bg-slate-700 text-blue-500 focus:ring-blue-500 focus:ring-offset-slate-800"
                  />
                </td>
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
                        onClick={() => handleStart(instance.name)}
                        className="px-3 py-1.5 bg-green-600 hover:bg-green-500 text-white text-xs font-medium rounded-lg transition-all shadow-lg shadow-green-500/20"
                      >
                        Start
                      </button>
                    ) : (
                      <button
                        onClick={() => handleStop(instance.name)}
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

      {/* Confirmation Dialogs */}
      <ConfirmDialog
        isOpen={showConfirmDialog === 'sleep'}
        onClose={() => setShowConfirmDialog(null)}
        onConfirm={handleBulkSleep}
        title="Confirm Sleep Operation"
        message={`Are you sure you want to sleep ${stoppableSelected.length} database instance(s)? This will stop them and they won't be accessible until woken.`}
        confirmText={`Sleep ${stoppableSelected.length} Instance(s)`}
        confirmVariant="warning"
        loading={bulkLoading}
      />

      <ConfirmDialog
        isOpen={showConfirmDialog === 'wake'}
        onClose={() => setShowConfirmDialog(null)}
        onConfirm={handleBulkWake}
        title="Confirm Wake Operation"
        message={`Are you sure you want to wake ${startableSelected.length} database instance(s)? This will start them and resume billing.`}
        confirmText={`Wake ${startableSelected.length} Instance(s)`}
        confirmVariant="success"
        loading={bulkLoading}
      />
    </div>
  )
}

export default InstancesPage
