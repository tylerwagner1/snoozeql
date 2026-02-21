import { useState, useEffect } from 'react'
import api from '../lib/api'
import type { Event } from '../lib/api'

const AuditLogPage = () => {
  const [events, setEvents] = useState<Event[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<'all' | 'sleep' | 'wake'>('all')

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const data = await api.getEvents(100)
        setEvents(data || [])
      } catch (err) {
        setError('Failed to load audit log')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    fetchEvents()
  }, [])

  const filteredEvents = events.filter(e => {
    if (filter === 'all') return true
    return e.event_type === filter
  })

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleString()
  }

  const getEventIcon = (eventType: string) => {
    switch (eventType) {
      case 'sleep':
        return (
          <div className="p-2 bg-yellow-500/10 rounded-lg">
            <svg className="w-5 h-5 text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
          </div>
        )
      case 'wake':
        return (
          <div className="p-2 bg-green-500/10 rounded-lg">
            <svg className="w-5 h-5 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
          </div>
        )
      default:
        return (
          <div className="p-2 bg-slate-500/10 rounded-lg">
            <svg className="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        )
    }
  }

  if (loading) return <div className="p-8 text-center text-slate-400">Loading audit log...</div>
  if (error) return <div className="p-8 text-center text-red-400">{error}</div>

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Audit Log</h1>
          <p className="text-sm text-slate-400 mt-1">
            {filteredEvents.length} event{filteredEvents.length !== 1 ? 's' : ''}
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-2">
        <button
          onClick={() => setFilter('all')}
          className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
            filter === 'all'
              ? 'bg-blue-600 text-white'
              : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'
          }`}
        >
          All Events
        </button>
        <button
          onClick={() => setFilter('sleep')}
          className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
            filter === 'sleep'
              ? 'bg-yellow-600 text-white'
              : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'
          }`}
        >
          Sleep
        </button>
        <button
          onClick={() => setFilter('wake')}
          className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
            filter === 'wake'
              ? 'bg-green-600 text-white'
              : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'
          }`}
        >
          Wake
        </button>
      </div>

      {/* Events List */}
      <div className="bg-slate-800/50 shadow-lg border border-slate-700 rounded-xl overflow-hidden">
        {filteredEvents.length === 0 ? (
          <div className="p-12 text-center text-slate-400">
            <div className="inline-block p-4 bg-slate-700/50 rounded-full mb-4">
              <svg className="h-12 w-12 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
              </svg>
            </div>
            <p className="text-lg font-medium text-slate-300">No events found</p>
            <p className="text-sm mt-2">Operations will appear here once you start sleeping or waking instances</p>
          </div>
        ) : (
          <div className="divide-y divide-slate-700">
            {filteredEvents.map(event => (
              <div key={event.id} className="p-4 hover:bg-slate-700/30 transition-colors">
                <div className="flex items-start gap-4">
                  {getEventIcon(event.event_type)}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className={`px-2 py-0.5 text-xs font-medium rounded-full capitalize ${
                        event.event_type === 'sleep' 
                          ? 'bg-yellow-500/10 text-yellow-400 border border-yellow-500/30'
                          : event.event_type === 'wake'
                          ? 'bg-green-500/10 text-green-400 border border-green-500/30'
                          : 'bg-slate-500/10 text-slate-400 border border-slate-500/30'
                      }`}>
                        {event.event_type}
                      </span>
                      <span className="text-xs text-slate-500">
                        {formatDate(event.created_at)}
                      </span>
                    </div>
                    <p className="mt-1 text-sm text-slate-300">
                      Instance <span className="font-mono text-blue-400">{event.instance_id.slice(0, 8)}...</span>
                      {' '}changed from{' '}
                      <span className="font-medium text-slate-200">{event.previous_status}</span>
                      {' '}to{' '}
                      <span className="font-medium text-slate-200">{event.new_status}</span>
                    </p>
                    <p className="mt-1 text-xs text-slate-500">
                      Triggered by: <span className="text-slate-400">{event.triggered_by}</span>
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default AuditLogPage
