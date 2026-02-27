import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { X, Clock, Globe, Database } from 'lucide-react'
import api from '../lib/api'
import type { Schedule } from '../lib/api'

const ScheduleNewPage = () => {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const [form, setForm] = useState({
    name: '',
    description: '',
    timezone: 'America/New_York',
    sleep_cron: '0 22 * * *',
    wake_cron: '0 7 * * *',
    selectors: [
      { name: { pattern: '', type: 'exact' } }
    ],
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      const data: Omit<Schedule, 'id' | 'created_at' | 'updated_at'> = {
        name: form.name,
        description: form.description,
        timezone: form.timezone,
        sleep_cron: form.sleep_cron,
        wake_cron: form.wake_cron,
        selectors: form.selectors,
        enabled: true,
      }

      await api.createSchedule(data)
      setSuccess(true)
      setTimeout(() => navigate('/schedules'), 2000)
    } catch (err: any) {
      setError(err.message || 'Failed to create schedule')
    } finally {
      setLoading(false)
    }
  }

  const addSelector = () => {
    setForm(prev => ({
      ...prev,
      selectors: [...prev.selectors, { name: { pattern: '', type: 'exact' } }]
    }))
  }

  const removeSelector = (index: number) => {
    setForm(prev => ({
      ...prev,
      selectors: prev.selectors.filter((_, i) => i !== index)
    }))
  }

  const updateSelector = (index: number, field: string, value: string) => {
    setForm(prev => ({
      ...prev,
      selectors: prev.selectors.map((selector, i) => {
        if (i === index) {
          return {
            ...selector,
            name: {
              ...selector.name,
              [field]: value
            }
          }
        }
        return selector
      })
    }))
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-white">Create Schedule</h1>
          <p className="text-sm text-slate-400 mt-1">Configure when database instances should sleep and wake</p>
        </div>
        <button
          onClick={() => navigate('/schedules')}
          className="p-2 text-slate-400 hover:text-white transition-colors rounded-lg hover:bg-slate-800"
        >
          <X className="h-6 w-6" />
        </button>
      </div>

      {success && (
        <div className="mb-6 p-4 bg-green-500/10 border border-green-500/30 rounded-lg flex items-center gap-2 text-green-400">
          <div className="flex-shrink-0 p-1 bg-green-500 rounded-full">
            <svg className="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <span>Schedule created successfully! Redirecting...</span>
        </div>
      )}

      {error && (
        <div className="mb-6 p-4 bg-red-500/10 border border-red-500/30 rounded-lg flex items-center gap-2 text-red-400">
          <div className="flex-shrink-0 p-1 bg-red-500 rounded-full">
            <svg className="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700">
        <div className="space-y-6">
          {/* Basic Info */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
              <div className="p-1.5 bg-blue-500/10 rounded-lg">
                <Database className="h-5 w-5 text-blue-400" />
              </div>
              Basic Information
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">Schedule Name</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={e => setForm(prev => ({ ...prev, name: e.target.value }))}
                  placeholder="e.g., Nightly Sleep Schedule"
                  required
                  className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
                />
                <p className="text-xs text-slate-500 mt-1">
                  A friendly name to identify this schedule
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">Description (Optional)</label>
                <textarea
                  value={form.description}
                  onChange={e => setForm(prev => ({ ...prev, description: e.target.value }))}
                  placeholder="Optional description of what this schedule does"
                  className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>
          </div>

          {/* Time Settings */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
              <div className="p-1.5 bg-purple-500/10 rounded-lg">
                <Clock className="h-5 w-5 text-purple-400" />
              </div>
              Timing Configuration
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">Timezone</label>
                <select
                  value={form.timezone}
                  onChange={e => setForm(prev => ({ ...prev, timezone: e.target.value }))}
                  className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
                >
                  <option value="America/New_York">America/New_York (EST/EDT)</option>
                  <option value="America/Chicago">America/Chicago (CST/CDT)</option>
                  <option value="America/Denver">America/Denver (MST/MDT)</option>
                  <option value="America/Los_Angeles">America/Los_Angeles (PST/PDT)</option>
                  <option value="UTC">UTC (Coordinated Universal Time)</option>
                  <option value="Europe/London">Europe/London (GMT/BST)</option>
                  <option value="Asia/Tokyo">Asia/Tokyo (JST)</option>
                  <option value="Australia/Sydney">Australia/Sydney (AEST/AEDT)</option>
                </select>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">Wake Cron (When to start)</label>
                  <input
                    type="text"
                    value={form.wake_cron}
                    onChange={e => setForm(prev => ({ ...prev, wake_cron: e.target.value }))}
                    placeholder="0 7 * * *"
                    required
                    className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-green-500 font-mono text-sm"
                  />
                  <p className="text-xs text-slate-500 mt-1">
                    <a href="https://crontab.guru/" target="_blank" rel="noopener noreferrer" className="text-green-400 hover:underline">
                      Cron generator
                    </a>
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">Sleep Cron (When to stop)</label>
                  <input
                    type="text"
                    value={form.sleep_cron}
                    onChange={e => setForm(prev => ({ ...prev, sleep_cron: e.target.value }))}
                    placeholder="0 22 * * *"
                    required
                    className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-purple-500 font-mono text-sm"
                  />
                  <p className="text-xs text-slate-500 mt-1">
                    <a href="https://crontab.guru/" target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:underline">
                      Cron generator
                    </a>
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Selectors */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
              <div className="p-1.5 bg-green-500/10 rounded-lg">
                <Globe className="h-5 w-5 text-green-400" />
              </div>
              Instance Selection
            </h2>
            <p className="text-sm text-slate-400 mb-4">
              Select which database instances this schedule applies to
            </p>

            {form.selectors.map((selector, index) => (
              <div key={index} className="p-4 bg-slate-900/50 border border-slate-700 rounded-lg mb-4">
                <div className="flex items-center gap-2 mb-3">
                  <span className="text-sm font-medium text-slate-400">Selector {index + 1}</span>
                  <button
                    type="button"
                    onClick={() => removeSelector(index)}
                    className="text-red-400 hover:text-red-300 ml-auto"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
                <div className="space-y-3">
                  <div>
                    <label className="block text-xs font-medium text-slate-400 mb-1">Instance Name Pattern</label>
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={selector.name?.pattern || ''}
                        onChange={e => updateSelector(index, 'pattern', e.target.value)}
                        placeholder="e.g., prod-*"
                        className="flex-1 px-3 py-2 bg-slate-950 border border-slate-700 rounded text-sm text-white focus:outline-none focus:border-green-500 font-mono"
                      />
                      <select
                        value={selector.name?.type || 'exact'}
                        onChange={e => updateSelector(index, 'type', e.target.value)}
                        className="px-3 py-2 bg-slate-950 border border-slate-700 rounded text-sm text-white focus:outline-none focus:border-green-500"
                      >
                        <option value="exact">Exact match</option>
                        <option value="contains">Contains</option>
                        <option value="prefix">Starts with</option>
                        <option value="suffix">Ends with</option>
                        <option value="regex">Regex pattern</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>
            ))}

            <button
              type="button"
              onClick={addSelector}
              className="w-full py-2 border-2 border-dashed border-slate-700 text-slate-400 hover:text-slate-300 hover:border-slate-600 rounded-lg transition-colors"
            >
              + Add Another Selector
            </button>
          </div>
        </div>

        <div className="mt-8 pt-6 border-t border-slate-700">
          <div className="flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={() => navigate('/schedules')}
              className="px-6 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg font-medium transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !form.name}
              className="px-6 py-2.5 bg-gradient-to-r from-green-600 to-green-700 hover:from-green-500 hover:to-green-600 text-white rounded-lg font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-green-500/20"
            >
              {loading ? 'Creating...' : 'Create Schedule'}
            </button>
          </div>
        </div>
      </form>
    </div>
  )
}

export default ScheduleNewPage
