import { useState, useEffect } from 'react';
import { Dialog, DialogPanel, DialogTitle } from '@headlessui/react';
import { X } from 'lucide-react';
import clsx from 'clsx';
import { describeCron } from '../lib/cronUtils';
import api from '../lib/api';
import { Schedule, Instance, Selector } from '../lib/api';
import { FilterBuilder } from './FilterBuilder';

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
  const [sleepCron, setSleepCron] = useState('');
  const [wakeCron, setWakeCron] = useState('');
  const [selectors, setSelectors] = useState<Selector[]>([]);
  const [instances, setInstances] = useState<Instance[]>([]);
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
        setSleepCron(schedule.sleep_cron);
        setWakeCron(schedule.wake_cron);
        setSelectors(schedule.selectors || []);
      } else {
        // Create mode: reset form with sensible defaults
        setName('');
        setDescription('');
        setTimezone('America/New_York');
        setSleepCron('0 22 * * 1-5'); // Default: 10pm weekdays
        setWakeCron('0 7 * * 1-5');   // Default: 7am weekdays
        setSelectors([]);
        setNameError('');
      }
      setError(null);
    }
  }, [isOpen, schedule]);

  // Fetch instances for preview
  useEffect(() => {
    if (isOpen) {
      api.getInstances()
        .then(setInstances)
        .catch(console.error);
    }
  }, [isOpen]);

  // Validate form
  const validateForm = (): boolean => {
    if (!name.trim()) {
      setNameError('Schedule name is required');
      return false;
    }
    setNameError('');

    if (!sleepCron.trim() || !wakeCron.trim()) {
      setError('Both sleep and wake CRON expressions are required');
      return false;
    }

    // Validate CRON syntax
    try {
      describeCron(sleepCron);
      describeCron(wakeCron);
    } catch (err) {
      setError('Both sleep and wake CRON expressions must be valid');
      return false;
    }

    return true;
  };

  const handleSubmit = async () => {
    if (!validateForm()) return;

    setLoading(true);
    setError(null);

    try {
      const scheduleData = {
        name,
        description,
        timezone,
        sleep_cron: sleepCron,
        wake_cron: wakeCron,
        selectors,
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
    } catch (err: any) {
      const errorMessage = err?.response?.data?.message || err?.response?.data?.error || err?.message || 'Failed to save schedule';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  // Get summary text
  const getSummaryText = () => {
    if (!sleepCron && !wakeCron) {
      return 'None set';
    }
    const sleepDesc = sleepCron ? describeCron(sleepCron) : 'Not set';
    const wakeDesc = wakeCron ? describeCron(wakeCron) : 'Not set';
    return `Wake: ${wakeDesc} | Sleep: ${sleepDesc}`;
  };

  if (!isOpen) return null;

  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <div className="fixed inset-0 bg-black/70" aria-hidden="true" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-xl bg-slate-900 border border-slate-700 shadow-2xl">
          <div className="flex items-center justify-between p-6 border-b border-slate-800">
            <DialogTitle className="text-xl font-semibold text-white">
              {schedule ? 'Edit Schedule' : 'Create New Schedule'}
            </DialogTitle>
            <button
              onClick={onClose}
              className="text-slate-400 hover:text-white transition-colors"
            >
              <X className="h-6 w-6" />
            </button>
          </div>

          <div className="p-6 space-y-6">
            {/* Name field */}
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-slate-300 mb-1">
                Schedule Name <span className="text-red-500">*</span>
              </label>
              <input
                id="name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Office Hours Schedule"
                className={`w-full px-4 py-2 bg-slate-900 border rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 ${nameError ? 'border-red-500' : 'border-slate-700'}`}
              />
              {nameError && <p className="mt-1 text-xs text-red-400">{nameError}</p>}
            </div>

            {/* Description field */}
            <div>
              <label htmlFor="description" className="block text-sm font-medium text-slate-300 mb-1">
                Description
              </label>
              <textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="What this schedule does..."
                rows={2}
                className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>

            {/* Timezone select */}
            <div>
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

            {/* Wake CRON */}
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
              {wakeCron && (
                <p className="mt-1 text-xs text-slate-400 bg-slate-800/50 p-2 rounded">
                  <span className="font-medium text-slate-300">What this does:</span> {describeCron(wakeCron)}
                  <br />
                  <span className="text-slate-500 mt-1 block">
                    {wakeCron} → Databases will wake at this time
                  </span>
                </p>
              )}
            </div>

            {/* Sleep CRON */}
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
              {sleepCron && (
                <p className="mt-1 text-xs text-slate-400 bg-slate-800/50 p-2 rounded">
                  <span className="font-medium text-slate-300">What this does:</span> {describeCron(sleepCron)}
                  <br />
                  <span className="text-slate-500 mt-1 block">
                    {sleepCron} → Databases will sleep at this time
                  </span>
                </p>
              )}
              <p className="mt-1 text-xs text-slate-500">
                Standard 5-field CRON format. See{' '}
                <a href="https://crontab.guru" target="_blank" rel="noopener noreferrer" className="text-indigo-400 hover:underline">
                  crontab.guru
                </a>
              </p>
            </div>

            {/* Summary display */}
            <div className="p-4 bg-slate-800/50 rounded-lg border border-slate-700">
              <p className="text-sm text-slate-300">
                <span className="font-medium">Schedule summary:</span>{' '}
                <span className="text-indigo-400">{getSummaryText()}</span>
              </p>
            </div>

            {/* Filter Builder Section */}
            <div className="pt-6 border-t border-slate-700">
              <FilterBuilder
                selectors={selectors}
                onChange={setSelectors}
                instances={instances}
              />
            </div>

            {/* Error message */}
            {error && (
              <div className="p-4 bg-red-500/10 border border-red-500/50 rounded-lg">
                <p className="text-sm text-red-400">{error}</p>
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="flex justify-end gap-3 p-6 border-t border-slate-800">
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
