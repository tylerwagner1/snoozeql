import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import api from '../lib/api'
import type { Instance, HourlyMetric } from '../lib/api'
import { MetricsChart } from '../components/MetricsChart'

const InstanceDetailPage = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [instance, setInstance] = useState<Instance | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [collecting, setCollecting] = useState(false)

  // State for modal
  const [showMetrics, setShowMetrics] = useState(false)
  const [metricsForModal, setMetricsForModal] = useState<HourlyMetric[]>([])

  useEffect(() => {
    if (!id) return
    
    const fetchInstance = async () => {
      try {
        const instanceData = await api.getInstance(id)
        setInstance(instanceData)
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

  const handleCollectMetrics = async () => {
    if (!id) return
    setCollecting(true)
    try {
      await api.collectInstanceMetrics(id)
      // Refresh metrics after collection
      const metricsData = await api.getInstanceMetrics(id)
      // Show success toast
      toast.success('Metrics collected successfully')
      setMetricsForModal(metricsData)
      setShowMetrics(true)
    } catch (err: any) {
      const message = err?.response?.data?.message || 'Failed to collect metrics'
      console.error('Failed to collect metrics:', message)
      toast.error(message)
    } finally {
      setCollecting(false)
    }
  }

  // Simple modal to show collected metrics
  const MetricModal = ({ metrics }: { metrics: HourlyMetric[] }) => {
    if (!showMetrics) return null
    
    // Group metrics by metric_name
    const metricsByType: Record<string, HourlyMetric> = {}
    metrics.forEach(m => {
      metricsByType[m.metric_name] = m
    })

    const MetricRow = ({ label, metric }: { label: string; metric?: HourlyMetric }) => {
      if (!metric) return null
      return (
        <div className="flex justify-between items-center py-2 border-b border-border last:border-0">
          <span className="text-muted-foreground">{label}</span>
          <span className="font-medium text-foreground">
            {metric.avg_value.toFixed(1)}
            {label.includes('CPU') || label.includes('Memory') ? '%' : ''}
          </span>
        </div>
      )
    }

    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
        <div className="bg-card rounded-lg shadow-xl max-w-md w-full border border-border">
          <div className="p-6 border-b border-border">
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-semibold text-foreground">Current Metrics</h3>
              <button
                onClick={() => {
                  setShowMetrics(false)
                  setMetricsForModal([])
                }}
                className="text-muted-foreground hover:text-foreground"
              >
                <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
          
          <div className="p-6">
            {metrics.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-muted-foreground">No metrics data available</p>
              </div>
            ) : (
              <div className="space-y-1">
                <MetricRow label="CPU Utilization" metric={metricsByType['CPUUtilization']} />
                <MetricRow label="Memory Available" metric={metricsByType['FreeableMemory']} />
                <MetricRow label="Database Connections" metric={metricsByType['DatabaseConnections']} />
              </div>
            )}
          </div>
          
          <div className="p-4 border-t border-border bg-muted/50 rounded-b-lg">
            <p className="text-xs text-muted-foreground text-center">
              Metrics are collected every 15 minutes
            </p>
          </div>
        </div>
      </div>
    )
  }

  if (loading) return <div className="p-8 text-center text-muted-foreground">Loading instance...</div>
  if (error) return <div className="p-8 text-center text-red-500">{error}</div>
  if (!instance) return <div className="p-8 text-center text-muted-foreground">Instance not found</div>

  return (
    <div className="space-y-6">
      {/* Header with name and status */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h1 className="text-3xl font-bold text-foreground">{instance.name}</h1>
          <span className={`px-3 py-1 text-sm rounded-full font-medium capitalize ${
            instance.status === 'running' ? 'bg-green-500/20 text-green-400' :
            instance.status === 'stopped' ? 'bg-gray-500/20 text-gray-400' :
            'bg-blue-500/20 text-blue-400'
          }`}>
            {instance.status}
          </span>
        </div>
        <div className="flex space-x-3">
          <button
            onClick={() => navigate(-1)}
            className="px-4 py-2 bg-card border border-border rounded-lg text-sm font-medium text-foreground hover:bg-muted"
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

      {/* Main content grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left column - Configuration, Tags, Metrics History */}
        <div className="lg:col-span-2 space-y-6">
          {/* Configuration Card */}
          <div className="bg-card shadow-sm border border-border rounded-lg p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">Configuration</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-muted-foreground uppercase">Provider</p>
                <p className="text-sm font-medium text-foreground">{instance.provider === 'aws' ? 'AWS' : 'GCP'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground uppercase">Region</p>
                <p className="text-sm font-medium text-foreground">{instance.region}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground uppercase">Engine</p>
                <p className="text-sm font-medium text-foreground">{instance.engine}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground uppercase">Instance Type</p>
                <p className="text-sm font-medium text-foreground">{instance.instance_type}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground uppercase">Created</p>
                <p className="text-sm font-medium text-foreground">{new Date(instance.created_at).toLocaleDateString()}</p>
              </div>
            </div>
          </div>

          {/* Tags Card */}
          <div className="bg-card shadow-sm border border-border rounded-lg p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">Tags</h2>
            {Object.keys(instance.tags).length > 0 ? (
              <div className="space-y-2">
                {Object.entries(instance.tags).map(([key, value]) => (
                  <div key={key} className="flex items-center justify-between p-2 bg-muted rounded-lg">
                    <span className="text-sm font-medium text-foreground">{key}</span>
                    <span className="text-sm text-muted-foreground">{value}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No tags configured</p>
            )}
          </div>

          {/* Metrics History Chart */}
          {id && (
            <div className="bg-card shadow-sm border border-border rounded-lg p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">Metrics History</h2>
              <MetricsChart instanceId={id} />
            </div>
          )}
        </div>

        {/* Right column - Actions */}
        <div className="space-y-6">
          <div className="bg-card shadow-sm border border-border rounded-lg p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">Actions</h2>
            <div className="space-y-2">
              <button
                onClick={() => navigate(`/audit-log?instance_id=${instance.id}`)}
                className="w-full px-4 py-2 text-sm font-medium text-foreground bg-card border border-border rounded-lg hover:bg-muted"
              >
                View Logs
              </button>
              <button
                onClick={() => navigate('/schedules')}
                className="w-full px-4 py-2 text-sm font-medium text-foreground bg-card border border-border rounded-lg hover:bg-muted"
              >
                Configure Schedule
              </button>
              <button
                onClick={handleCollectMetrics}
                disabled={collecting}
                className="w-full px-4 py-2 text-sm font-medium text-foreground bg-card border border-border rounded-lg hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
                title={collecting ? 'Collecting metrics...' : 'Click to test metrics connection'}
              >
                {collecting ? 'Collecting...' : 'Test Metrics'}
              </button>
              <button
                onClick={() => navigate('/instances')}
                className="w-full px-4 py-2 text-sm font-medium text-red-400 bg-card border border-red-500/30 rounded-lg hover:bg-red-500/10"
              >
                Delete Instance
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <MetricModal metrics={metricsForModal} />
    </div>
  )
}

export default InstanceDetailPage
