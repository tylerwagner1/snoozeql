import { useState, useEffect } from 'react';
import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react';
import { XMarkIcon } from 'lucide-react';
import clsx from 'clsx';
import { WeeklyScheduleGrid } from './WeeklyScheduleGrid';
import { createEmptyGrid, gridToCron, cronToGrid, formatGridSummary, formatHour } from '../lib/cronUtils';
import api from '../lib/api';
import { Schedule } from '../lib/api';

interface ScheduleModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  schedule?: Schedule | null;
}

const TIMEZONES = [
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'UTC',
  'Europe/London',
  'Asia/Tokyo',
  'Australia/Sydney',
];

export function ScheduleModal({
  isOpen,
  onClose,
  onSuccess,
  schedule,
}: ScheduleModalProps) {
  // Form state
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [timezone, setTimezone] = useState('America/New_York');
  const [grid, setGrid] = useState<boolean[][]>(createEmptyGrid());
  const [showCronMode, setShowCronMode] = useState(false);
  const [sleepCron, setSleepCron] = useState('');
  const [wakeCron, setWakeCron] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [nameError, setNameError] = useState('');

  // Initialize form when modal opens
  useEffect(() => {
    if (isOpen) {
      if (schedule) {
        // Edit mode: populate from existing schedule
        setName(schedule.name);
        setDescription(schedule.description || '');
        setTimezone(schedule.timezone);
        const parsedGrid = cronToGrid(schedule.sleep_cron, schedule.wake_cron);
        setGrid(parsedGrid);
        setSleepCron(schedule.sleep_cron);
        setWakeCron(schedule.wake_cron);
      } else {
        // Create mode: reset form
        setName('');
        setDescription('');
        setTimezone('America/New_York');
        setGrid(createEmptyGrid());
        setSleepCron('');
        setWakeCron('');
        setNameError('');
      }
      setError(null);
    }
  }, [isOpen, schedule]);

  // Sync CRON when grid changes (in grid mode)
  useEffect(() => {
    if (!showCronMode && isOpen) {
      const result = gridToCron(grid);
      if (result) {
        setSleepCron(result.sleepCron);
        setWakeCron(result.wakeCron);
      }
    }
  }, [grid, showCronMode, isOpen]);

  // Handle mode switch - convert grid to CRON
  const handleSwitchToCron = () => {
    if (!showCronMode) {
      const result = gridToCron(grid);
      if (result) {
        setSleepCron(result.sleepCron);
        setWakeCron(result.wakeCron);
      }
      setShowCronMode(true);
    } else {
      setShowCronMode(false);
    }
  };

  // Handle CRON mode to grid conversion
  const handleSwitchToGrid = () => {
    try {
      const parsedGrid = cronToGrid(sleepCron, wakeCron);
      setGrid(parsedGrid);
      setShowCronMode(false);
      setError(null);
    } catch (err) {
      setError('Invalid CRON expressions. Please check the format.');
    }
  };

  // Get summary text for grid
  const getSummaryText = () => {
    if (showCronMode) {
      const sleepSummary = sleepCron || 'Not set';
      const wakeSummary = wakeCron || 'Not set';
      return `Sleep: ${sleepSummary} | Wake: ${wakeSummary}`;
    }

    const summary = formatGridSummary(grid);
    return `Sleep: ${summary.sleepHours} on ${summary.activeDays}`;
  };

  // Validate form
  const validateForm = (): boolean => {
    if (!name.trim()) {
      setNameError('Schedule name is required');
      return false;
    }
    setNameError('');

    if (showCronMode) {
      if (!sleepCron.trim() || !wakeCron.trim()) {
        setError('Both sleep and wake CRON expressions are required');
        return false;
      }
    } else {
      // Check if any sleep hours are selected
      const hasSleep = grid.some(day => day.some(isSleep => isSleep));
      if (!hasSleep) {
        setError('Please select at least one sleep hour');
        return false;
      }
    }

    setError(null);
    return true;
  };

  // Submit handler
  const handleSubmit = async () => {
    if (!validateForm()) return;

    setLoading(true);

    try {
      const scheduleData: Omit<Schedule, 'id' | 'created_at' | 'updated_at'> = {
        name,
        description,
        timezone,
        selectors: [],
        sleep_cron: showCronMode ? sleepCron : (gridToCron(grid)?.sleepCron || '0 22 * * 1-5'),
        wake_cron: showCronMode ? wakeCron : (gridToCron(grid)?.wakeCron || '0 7 * * 1-5'),
        enabled: true,
      };

      if (schedule) {
        // Update existing schedule
        await api.updateSchedule(schedule.id, scheduleData);
      } else {
        // Create new schedule
        await api.createSchedule(scheduleData);
      }

      onSuccess();
      onClose();
    } catch (err) {
      setError('Failed to save schedule. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Get title based on mode
  const getTitle = () => {
    return schedule ? 'Edit Schedule' : 'Create Schedule';
  };

  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <DialogBackdrop
        transition
        className="fixed inset-0 bg-black/50 backdrop-blur-sm duration-200 ease-out data-[closed]:opacity-0"
      />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel
          transition
          className="max-w-4xl w-full bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-2xl duration-200 ease-out data-[closed]:scale-95 data-[closed]:opacity-0 flex flex-col max-h-[90vh]"
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <DialogTitle className="text-xl font-bold text-white">
              {getTitle()}
            </DialogTitle>
            <button
              onClick={onClose}
              className="text-slate-400 hover:text-white transition-colors"
              aria-label="Close modal"
            >
              <XMarkIcon className="w-6 h-6" />
            </button>
          </div>

          {/* Name input */}
          <div className="mb-4">
            <label htmlFor="scheduleName" className="block text-sm font-medium text-slate-300 mb-1">
              Schedule Name <span className="text-red-500">*</span>
            </label>
            <input
              id="scheduleName"
              type="text"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (nameError) setNameError('');
              }}
              placeholder="e.g., Nightly Sleep"
              className={clsx(
                'w-full px-4 py-2 bg-slate-900 border rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500',
                nameError ? 'border-red-500' : 'border-slate-700'
              )}
            />
            {nameError && <p className="mt-1 text-sm text-red-500">{nameError}</p>}
          </div>

          {/* Description input */}
          <div className="mb-4">
            <label htmlFor="scheduleDescription" className="block text-sm font-medium text-slate-300 mb-1">
              Description (Optional)
            </label>
            <textarea
              id="scheduleDescription"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What this schedule does..."
              rows={2}
              className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          {/* Timezone select */}
          <div className="mb-6">
            <label htmlFor="timezone" className="block text-sm font-medium text-slate-300 mb-1">
              Timezone
            </label>
            <select
              id="timezone"
              value={timezone}
              onChange={(e) => setTimezone(e.target.value)}
              className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              {TIMEZONES.map((tz) => (
                <option key={tz} value={tz}>
                  {tz.replace('_', ' ')}
                </option>
              ))}
            </select>
          </div>

          {/* Time selection section */}
          <div className="mb-6 flex-1 overflow-y-auto pr-2">
            <div className="mb-4">
              <h3 className="text-sm font-medium text-slate-300 mb-1">Sleep Hours</h3>
              <p className="text-xs text-slate-400">Paint the hours when databases should sleep (dark = sleep)</p>
            </div>

            {/* Mode toggle */}
            <div className="mb-4">
              <button
                onClick={showCronMode ? handleSwitchToGrid : handleSwitchToCron}
                className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors"
              >
                {showCronMode ? 'Switch to Grid' : 'Switch to CRON'}
              </button>
            </div>

            {/* Grid or CRON mode */}
            {showCronMode ? (
              <div className="space-y-4">
                <div>
                  <label htmlFor="sleepCron" className="block text-sm font-medium text-slate-300 mb-1">
                    Sleep CRON <span className="text-red-500">*</span>
                  </label>
                  <input
                    id="sleepCron"
                    type="text"
                    value={sleepCron}
                    onChange={(e) => setSleepCron(e.target.value)}
                    placeholder="0 22 * * 1-5"
                    className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono"
                  />
                  <p className="mt-1 text-xs text-slate-500">
                    Standard 5-field CRON format. See{' '}
                    <a href="https://crontab.guru" target="_blank" rel="noopener noreferrer" className="text-indigo-400 hover:underline">
                      crontab.guru
                    </a>
                  </p>
                </div>

                <div>
                  <label htmlFor="wakeCron" className="block text-sm font-medium text-slate-300 mb-1">
                    Wake CRON <span className="text-red-500">*</span>
                  </label>
                  <input
                    id="wakeCron"
                    type="text"
                    value={wakeCron}
                    onChange={(e) => setWakeCron(e.target.value)}
                    placeholder="0 7 * * 1-5"
                    className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono"
                  />
                </div>
              </div>
            ) : (
              <WeeklyScheduleGrid grid={grid} onChange={setGrid} />
            )}

            {/* Summary display */}
            <div className="mt-4 p-3 bg-slate-900 rounded-lg border border-slate-700">
              <p className="text-sm text-slate-300">
                <span className="font-medium">Current schedule:</span>{' '}
                <span className="text-indigo-400">{getSummaryText()}</span>
              </p>
            </div>
          </div>

          {/* Error message */}
          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/50 rounded-lg">
              <p className="text-sm text-red-400">{error}</p>
            </div>
          )}

          {/* Footer */}
          <div className="flex justify-end gap-3 mt-6 pt-4 border-t border-slate-700">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={loading}
              className={clsx(
                'px-4 py-2 text-white rounded-lg transition-colors disabled:opacity-50',
                schedule
                  ? 'bg-indigo-600 hover:bg-indigo-700'
                  : 'bg-green-600 hover:bg-green-700'
              )}
            >
              {loading ? 'Saving...' : schedule ? 'Save Changes' : 'Create Schedule'}
            </button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  );
}

export default ScheduleModal;
