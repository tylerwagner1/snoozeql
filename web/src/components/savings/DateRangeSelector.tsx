import clsx from 'clsx'

export type DateRange = '7d' | '30d' | '90d'

interface DateRangeSelectorProps {
  value: DateRange
  onChange: (value: DateRange) => void
}

const options: Array<{ value: DateRange; label: string }> = [
  { value: '7d', label: '7 days' },
  { value: '30d', label: '30 days' },
  { value: '90d', label: '90 days' },
]

export function DateRangeSelector({ value, onChange }: DateRangeSelectorProps) {
  return (
    <div className="flex items-center gap-1 bg-slate-800/50 rounded-lg p-1 border border-slate-700">
      {options.map((opt) => (
        <button
          key={opt.value}
          onClick={() => onChange(opt.value)}
          className={clsx(
            'px-4 py-2 text-sm font-medium rounded-md transition-all',
            value === opt.value
              ? 'bg-blue-600 text-white shadow-lg shadow-blue-500/20'
              : 'text-slate-400 hover:text-white hover:bg-slate-700'
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
