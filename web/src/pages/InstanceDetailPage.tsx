import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import api from '../lib/api'
import type { Instance, HourlyMetric } from '../lib/api'
import { MetricsChart } from '../components/MetricsChart'

// Metric Card Component
const MetricCard = ({ label, value, unit, min, max, samples }: { 
  label: string; 
  value?: number; 
  unit: string;
  min?: number;
  max?: number;
  samples?: number;
}) => {
  const hasData = value !== undefined && value !== null;
  
  if (!hasData) return null;

  return (
    <div className="p-4 bg-gray-50 rounded-lg border border-gray-100">
      <p className="text-xs text-gray-500 uppercase">{label}</p>
      <div className="flex items-baseline mt-1 space-x-1">
        <p className="text-2xl font-bold text-gray-900">{value.toFixed(1)}</p>
        <p className="text-sm text-gray-600">{unit}</p>
      </div>
      <div className="flex items-center mt-2 space-x-3 text-xs text-gray-500">
        {min !== undefined && (
          <span>Min: {min.toFixed(1)}{unit}</span>
        )}
        {max !== undefined && (
          <span>Max: {max.toFixed(1)}{unit}</span>
        )}
        {samples !== undefined && (
          <span className="text-gray-400">({samples} samples)</span>
        )}
      </div>
    </div>
  );
};

// Helper functions for extracting metric values
const getMetricValue = (metrics: HourlyMetric[], metricName: string): number | undefined => {
  const metric = metrics.find(m => m.metric_name.toLowerCase() === metricName.toLowerCase());
  return metric ? metric.avg_value : undefined;
};

const getMetricMin = (metrics: HourlyMetric[], metricName: string): number | undefined => {
  const metric = metrics.find(m => m.metric_name.toLowerCase() === metricName.toLowerCase());
  return metric ? metric.min_value : undefined;
};

const getMetricMax = (metrics: HourlyMetric[], metricName: string): number | undefined => {
  const metric = metrics.find(m => m.metric_name.toLowerCase() === metricName.toLowerCase());
  return metric ? metric.max_value : undefined;
};

const getMetricSamples = (metrics: HourlyMetric[], metricName: string): number | undefined => {
  const metric = metrics.find(m => m.metric_name.toLowerCase() === metricName.toLowerCase());
  return metric ? metric.sample_count : undefined;
};

const InstanceDetailPage = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [instance, setInstance] = useState<Instance | null>(null)
  const [metrics, setMetrics] = useState<HourlyMetric[]>([])
  const [idleTime, setIdleTime] = useState<string>('--')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [collecting, setCollecting] = useState(false)

  // Check if metrics are stale (no data or older than 30 min)
  const isMetricsStale = () => {
    if (metrics.length === 0) return true
    const latestMetric = metrics.reduce((latest, m) => {
      const mTime = new Date(m.hour).getTime()
      return mTime > latest ? mTime : latest
    }, 0)
    const thirtyMinutesAgo = Date.now() - (30 * 60 * 1000)
    return latestMetric < thirtyMinutesAgo
  }

  useEffect(() => {
    if (!id) return
    
    const fetchInstance = async () => {
      try {
        const [instanceData, metricsData] = await Promise.all([
          api.getInstance(id),
          api.getInstanceMetrics(id)
        ])
        setInstance(instanceData)
        setMetrics(metricsData)
        
        // Calculate idle time from metrics (hours since last active)
        if (metricsData.length > 0) {
          // Get the oldest metric hour and calculate time since then
          const oldestHour = metricsData.reduce((min, m) => {
            const mTime = new Date(m.hour).getTime()
            return mTime < min ? mTime : min
          }, new Date().getTime())
          
          const hoursSinceIdle = Math.floor((Date.now() - oldestHour) / (1000 * 60 * 60))
          if (hoursSinceIdle < 24) {
            setIdleTime(`${hoursSinceIdle}h ${Math.floor((hoursSinceIdle * 60) % 60)}m`)
          } else if (hoursSinceIdle < 168) { // less than a week
            setIdleTime(`${Math.floor(hoursSinceIdle / 24)}d ${hoursSinceIdle % 24}h`)
          } else {
            setIdleTime(`${Math.floor(hoursSinceIdle / 24)}d`)
          }
        } else {
          setIdleTime('--')
        }
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

   // State for modal
   const [showMetrics, setShowMetrics] = useState(false)
   const [metricsForModal, setMetricsForModal] = useState<HourlyMetric[]>([])

  const handleCollectMetrics = async () => {
    if (!id) return
    setCollecting(true)
    try {
      await api.collectInstanceMetrics(id)
      // Refresh metrics after collection
      const metricsData = await api.getInstanceMetrics(id)
      setMetrics(metricsData)
      // Show the metrics in a modal with the fresh data
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

  const MetricModal = ({ metrics }: { metrics: HourlyMetric[] }) => {
    if (!showMetrics) return null
    
    // Group metrics by metric_name
    const metricsByType: Record<string, HourlyMetric> = {}
    metrics.forEach(m => {
      metricsByType[m.metric_name] = m
    })

    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
        <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
          <div className="p-6 border-b">
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-semibold text-gray-900">Current Metrics</h3>
               <button
                 onClick={() => {
                   setShowMetrics(false)
                   setMetricsForModal([])
                 }}
                 className="text-gray-500 hover:text-gray-700"
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
                <p className="text-gray-500">No metrics data available</p>
              </div>
            ) : (
              <div className="grid grid-cols-2 gap-4">
                 {metricsByType['CPUUtilization'] && (
                   <MetricCard
                     label="CPU Utilization"
                     value={metricsByType['CPUUtilization'].avg_value}
                     unit="%"
                     min={metricsByType['CPUUtilization'].min_value}
                     max={metricsByType['CPUUtilization'].max_value}
                     samples={metricsByType['CPUUtilization'].sample_count}
                   />
                 )}
                 {metricsByType['FreeableMemory'] && (
                   <MetricCard
                     label="Memory Available"
                     value={metricsByType['FreeableMemory'].avg_value}
                     unit="%"
                     min={metricsByType['FreeableMemory'].min_value}
                     max={metricsByType['FreeableMemory'].max_value}
                     samples={metricsByType['FreeableMemory'].sample_count}
                   />
                 )}
                 {metricsByType['ReadIOPS'] && (
                   <MetricCard
                     label="Read IOPS"
                     value={metricsByType['ReadIOPS'].avg_value}
                     unit="IOPS"
                     min={metricsByType['ReadIOPS'].min_value}
                     max={metricsByType['ReadIOPS'].max_value}
                     samples={metricsByType['ReadIOPS'].sample_count}
                   />
                 )}
                 {metricsByType['WriteIOPS'] && (
                   <MetricCard
                     label="Write IOPS"
                     value={metricsByType['WriteIOPS'].avg_value}
                     unit="IOPS"
                     min={metricsByType['WriteIOPS'].min_value}
                     max={metricsByType['WriteIOPS'].max_value}
                     samples={metricsByType['WriteIOPS'].sample_count}
                   />
                 )}
                 {metricsByType['DatabaseConnections'] && (
                   <MetricCard
                     label="Database Connections"
                     value={metricsByType['DatabaseConnections'].avg_value}
                     unit=""
                     min={metricsByType['DatabaseConnections'].min_value}
                     max={metricsByType['DatabaseConnections'].max_value}
                     samples={metricsByType['DatabaseConnections'].sample_count}
                   />
                 )}
               </div>
            )}
          </div>
          
          <div className="p-4 border-t bg-gray-50">
            <p className="text-xs text-gray-500 text-center">
              Metrics are collected every 15 minutes
            </p>
          </div>
        </div>
      </div>
    )
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

          <div className="bg-white shadow-sm border rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Metrics</h2>
              {isMetricsStale() && (
                <span className="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 rounded-full">
                  Metrics unavailable
                </span>
              )}
            </div>
            {metrics.length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <MetricCard 
                  label="CPU Utilization" 
                   value={getMetricValue(metrics, 'CPUUtilization')}
                   unit="%"
                   min={getMetricMin(metrics, 'CPUUtilization')}
                   max={getMetricMax(metrics, 'CPUUtilization')}
                   samples={getMetricSamples(metrics, 'CPUUtilization')}
                 />
                 <MetricCard 
                   label="Memory Available" 
                   value={getMetricValue(metrics, 'FreeableMemory')}
                   unit="%"
                   min={getMetricMin(metrics, 'FreeableMemory')}
                   max={getMetricMax(metrics, 'FreeableMemory')}
                   samples={getMetricSamples(metrics, 'FreeableMemory')}
                 />
                 <MetricCard 
                   label="Database Connections" 
                   value={getMetricValue(metrics, 'DatabaseConnections')}
                   unit=""
                   min={getMetricMin(metrics, 'DatabaseConnections')}
                   max={getMetricMax(metrics, 'DatabaseConnections')}
                   samples={getMetricSamples(metrics, 'DatabaseConnections')}
                 />
                 <MetricCard 
                   label="Read IOPS" 
                   value={getMetricValue(metrics, 'ReadIOPS')}
                  unit=""
                  min={getMetricMin(metrics, 'readiops')}
                   max={getMetricMax(metrics, 'ReadIOPS')}
                   samples={getMetricSamples(metrics, 'ReadIOPS')}
                 />
                 <MetricCard 
                   label="Write IOPS" 
                   value={getMetricValue(metrics, 'WriteIOPS')}
                   unit=""
                   min={getMetricMin(metrics, 'WriteIOPS')}
                   max={getMetricMax(metrics, 'WriteIOPS')}
                   samples={getMetricSamples(metrics, 'WriteIOPS')}
                 />
                 <MetricCard 
                   label="Total IOPS" 
                   value={getMetricValue(metrics, 'ReadIOPS')}
                   unit=""
                   min={getMetricMin(metrics, 'ReadIOPS')}
                   max={getMetricMax(metrics, 'ReadIOPS')}
                   samples={getMetricSamples(metrics, 'ReadIOPS')}
                 />
              </div>
            ) : (
              <p className="text-sm text-gray-500">No metrics data available yet. Metrics are collected every 15 minutes.</p>
            )}
          </div>

          {/* Metrics History Chart */}
          {id && (
            <div className="mt-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Metrics History</h2>
              <MetricsChart instanceId={id} />
            </div>
          )}
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
                  {instance.status === 'stopped' ? 'N/A' : idleTime}
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
                    onClick={handleCollectMetrics}
                    disabled={collecting}
                    className="w-full px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    title={collecting ? 'Collecting metrics...' : 'Click to test metrics connection'}
                  >
                    {collecting ? 'Collecting...' : 'Test Metrics'}
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
          <MetricModal metrics={metricsForModal} />
        </div>
      </div>
    )
 }

export default InstanceDetailPage
