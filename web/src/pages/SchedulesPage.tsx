import { useState, useEffect } from 'react'
import { Clock, Plus } from 'lucide-react'
import api from '../lib/api'
import type { Schedule } from '../lib/api'
import { ScheduleModal } from '../components/ScheduleModal'
import { cronToGrid, formatGridSummary } from '../lib/cronUtils'

const SchedulesPage = () => {
  const [schedules, setSchedules] = useState<Schedule[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingSchedule, setEditingSchedule] = useState<Schedule | null>(null)

  useEffect(() => {
    const fetchSchedules = async () => {
      try {
        const data = await api.getSchedules()
        // Ensure selectors is an array for each schedule (API might return null or empty)
        const schedules = Array.isArray(data) ? data : []
        const safeSchedules = schedules.map(sched => ({
          ...sched,
          selectors: sched.selectors || []
        }))
        setSchedules(safeSchedules)
      } catch (err) {
        setError('Failed to load schedules')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    fetchSchedules()
  }, [])

  const handleToggle = async (id: string, enabled: boolean) => {
    try {
      await api.updateSchedule(id, { enabled: !enabled })
      setSchedules(prev => prev.map(sched => sched.id === id ? { ...sched, enabled: !enabled } : sched))
    } catch (err) {
      console.error('Failed to toggle schedule:', err)
    }
  }

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this schedule?')) return
    try {
      await api.deleteSchedule(id)
      setSchedules(prev => prev.filter(sched => sched.id !== id))
    } catch (err) {
      console.error('Failed to delete schedule:', err)
    }
  }

  if (loading) return <div className="p-8 text-center text-slate-400">Loading schedules...</div>
  if (error) return <div className="p-8 text-center text-red-400">{error}</div>

  // Compute summary for display
  const getSummary = (schedule: Schedule) => {
    try {
      const grid = cronToGrid(schedule.sleep_cron, schedule.wake_cron)
      return formatGridSummary(grid)
    } catch {
      return { activeDays: schedule.sleep_cron, sleepHours: 'Custom CRON' }
    }
  }

  const handleCreateSchedule = () => {
    setEditingSchedule(null)  // Create mode
    setIsModalOpen(true)
  }

  const handleEditSchedule = (schedule: Schedule) => {
    setEditingSchedule(schedule)
    setIsModalOpen(true)
  }

  const handleModalSuccess = async () => {
    // Refresh schedules list after create/update
    const data = await api.getSchedules()
    const schedules = Array.isArray(data) ? data : []
    const safeSchedules = schedules.map(sched => ({
      ...sched,
      selectors: sched.selectors || []
    }))
    setSchedules(safeSchedules)
    setIsModalOpen(false)
  }

  return (
    <div className="space-y-6">
      <ScheduleModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={handleModalSuccess}
        schedule={editingSchedule}
      />
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Schedules</h1>
          <p className="text-sm text-slate-400 mt-1">{schedules.length} total schedules</p>
        </div>
        <button
          onClick={handleCreateSchedule}
          className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-600 to-blue-700 border border-transparent rounded-lg text-sm font-medium text-white hover:from-blue-500 hover:to-blue-600 shadow-lg shadow-blue-500/20 transition-all"
        >
          <Plus className="h-4 w-4" />
          Create Schedule
        </button>
      </div>

      <div className="bg-slate-800/50 shadow-lg border border-slate-700 rounded-xl overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-slate-900/50">
            <tr>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Name</th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Active Days</th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Sleep Hours</th>
              <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
              <th className="px-6 py-4 text-right text-xs font-semibold text-slate-400 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-slate-800/30 divide-y divide-slate-700">
            {schedules.map(schedule => {
              const summary = getSummary(schedule)
              return (
                <tr key={schedule.id} className="hover:bg-slate-700/50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10 bg-gradient-to-br from-purple-600 to-indigo-700 rounded-lg flex items-center justify-center shadow-lg shadow-purple-500/20">
                        <span className="text-white font-bold truncate max-w-[30px]">
                          {schedule.name.substring(0, 2).toUpperCase()}
                        </span>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-medium text-white">{schedule.name}</div>
                        {schedule.description && (
                          <div className="text-xs text-slate-400 truncate max-w-[200px]">{schedule.description}</div>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <Clock className="h-4 w-4 text-slate-500 mr-2" />
                      <span className="text-sm text-slate-300">{summary.activeDays}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <span className="text-sm text-slate-300">{summary.sleepHours}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <label className="inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={schedule.enabled}
                        onChange={() => handleToggle(schedule.id, schedule.enabled)}
                        className="sr-only peer"
                      />
                      <div className="relative w-11 h-6 bg-slate-700 peer-focus:outline-none rounded-full peer dark:bg-slate-600 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600 shadow-inner"></div>
                      <span className="ml-3 text-sm font-medium text-slate-300">{schedule.enabled ? 'Enabled' : 'Disabled'}</span>
                    </label>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="flex justify-end space-x-2">
                      <button
                        onClick={() => handleEditSchedule(schedule)}
                        className="px-3 py-1.5 bg-gradient-to-r from-blue-600 to-blue-700 text-white text-xs font-medium rounded-lg transition-all shadow-lg shadow-blue-500/20 hover:from-blue-500 hover:to-blue-600"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(schedule.id)}
                        className="px-3 py-1.5 bg-gradient-to-r from-red-600 to-red-700 text-white text-xs font-medium rounded-lg transition-all shadow-lg shadow-red-500/20 hover:from-red-500 hover:to-red-600"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {schedules.length === 0 && (
          <div className="p-12 text-center">
            <div className="inline-block p-4 bg-slate-700/50 rounded-full mb-4">
              <Clock className="h-12 w-12 text-slate-500" />
            </div>
            <p className="text-lg font-medium text-slate-300">No schedules yet</p>
            <p className="text-sm text-slate-400 mt-2 mb-6">
              Create your first schedule to start saving on database costs
            </p>
            <button
              onClick={handleCreateSchedule}
              className="px-6 py-2.5 bg-gradient-to-r from-blue-600 to-blue-700 text-white rounded-lg font-medium hover:from-blue-500 hover:to-blue-600 shadow-lg shadow-blue-500/20"
            >
              Create your first schedule
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

export default SchedulesPage
