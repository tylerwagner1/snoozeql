import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import api from '../lib/api'
import type { Instance } from '../lib/api'

const InstanceDetailPage = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [instance, setInstance] = useState<Instance | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    
    const fetchInstance = async () => {
      try {
        const data = await api.getInstance(id)
        setInstance(data)
      } catch (err) {
        setError('Failed to load instance details')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    fetchInstance()
  }, [id])

  const handleStart = async () => {
    if (!id) return
    try {
      await api.startInstance(id)
      setInstance(prev => prev ? { ...prev, status: 'starting' } : null)
    } catch (err) {
      console.error('Failed to start instance:', err)
    }
  }

  const handleStop = async () => {
    if (!id) return
    try {
      await api.stopInstance(id)
      setInstance(prev => prev ? { ...prev, status: 'stopping' } : null)
    } catch (err) {
      console.error('Failed to stop instance:', err)
    }
  }

  if (loading) return <div className="p-8 text-center text-gray-500">Loading instance...</div>
  if (error) return <div className="p-8 text-center text-red-500">{error}</div>
  if (!instance) return <div className="p-8 text-center text-gray-500">Instance not found</div>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h1 className="text-3xl font-bold text-gray-900">{instance.name}</h1>
          <span className={`px-3 py-1 text-sm rounded-full font-medium capitalize ${
            instance.status === 'running' ? 'bg-green-100 text-green-800' :
            instance.status === 'stopped' ? 'bg-gray-100 text-gray-800' :
            'bg-blue-100 text-blue-800'
          }`}>
            {instance.status}
          </span>
        </div>
        <div className="flex space-x-3">
          <button
            onClick={() => navigate(-1)}
            className="px-4 py-2 bg-white border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Back
          </button>
          {instance.status === 'stopped' ? (
            <button
              onClick={handleStart}
              className="px-4 py-2 bg-green-600 border border-transparent rounded-lg text-sm font-medium text-white hover:bg-green-700"
            >
              Start Instance
            </button>
          ) : (
            <button
              onClick={handleStop}
              className="px-4 py-2 bg-red-600 border border-transparent rounded-lg text-sm font-medium text-white hover:bg-red-700"
            >
              Stop Instance
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white shadow-sm border rounded-lg p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Configuration</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-gray-500 uppercase">Provider</p>
                <p className="text-sm font-medium text-gray-900">{instance.provider === 'aws' ? 'AWS' : 'GCP'}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500 uppercase">Region</p>
                <p className="text-sm font-medium text-gray-900">{instance.region}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500 uppercase">Engine</p>
                <p className="text-sm font-medium text-gray-900">{instance.engine}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500 uppercase">Instance Type</p>
                <p className="text-sm font-medium text-gray-900">{instance.instance_type}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500 uppercase">Created</p>
                <p className="text-sm font-medium text-gray-900">{new Date(instance.created_at).toLocaleDateString()}</p>
              </div>
            </div>
          </div>

          <div className="bg-white shadow-sm border rounded-lg p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Tags</h2>
            {Object.keys(instance.tags).length > 0 ? (
              <div className="space-y-2">
                {Object.entries(instance.tags).map(([key, value]) => (
                  <div key={key} className="flex items-center justify-between p-2 bg-gray-50 rounded-lg">
                    <span className="text-sm font-medium text-gray-900">{key}</span>
                    <span className="text-sm text-gray-600">{value}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500">No tags configured</p>
            )}
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-white shadow-sm border rounded-lg p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Stats</h2>
            <div className="space-y-4">
              <div className="p-4 bg-gray-50 rounded-lg">
                <p className="text-xs text-gray-500 uppercase">Status</p>
                <p className="text-sm font-medium text-gray-900 mt-1 capitalize">{instance.status}</p>
              </div>
              <div className="p-4 bg-gray-50 rounded-lg">
                <p className="text-xs text-gray-500 uppercase">Current Cost</p>
                <p className="text-sm font-medium text-gray-900 mt-1">
                  ${Math.random() * 0.5 + 0.1}/hr (estimated)
                </p>
              </div>
              <div className="p-4 bg-gray-50 rounded-lg">
                <p className="text-xs text-gray-500 uppercase">Idle Time</p>
                <p className="text-sm font-medium text-gray-900 mt-1">
                  {instance.status === 'stopped' ? 'N/A' : '2h 15m'}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white shadow-sm border rounded-lg p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Actions</h2>
            <div className="space-y-2">
               <button
                 onClick={() => navigate(`/audit-log?instance_id=${instance.id}`)}
                 className="w-full px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
               >
                 View Logs
               </button>
               <button
                 onClick={() => navigate('/schedules')}
                 className="w-full px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
               >
                 Configure Schedule
               </button>
               <button
                 onClick={() => navigate('/instances')}
                 className="w-full px-4 py-2 text-sm font-medium text-red-700 bg-white border border-red-300 rounded-lg hover:bg-red-50"
               >
                 Delete Instance
               </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default InstanceDetailPage
