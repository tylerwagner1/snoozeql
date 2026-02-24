import { TrendingDown, Clock, DollarSign } from 'lucide-react'
import { formatCurrency } from '../../lib/formatters'
import type { SavingsSummary } from '../../lib/api'

interface SavingsSummaryCardsProps {
  data: SavingsSummary | null
  loading: boolean
}

export function SavingsSummaryCards({ data, loading }: SavingsSummaryCardsProps) {
  if (loading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-slate-800 rounded-xl p-5 border border-slate-700 animate-pulse">
            <div className="h-4 bg-slate-700 rounded w-24 mb-3" />
            <div className="h-8 bg-slate-700 rounded w-32 mb-2" />
            <div className="h-3 bg-slate-700 rounded w-20" />
          </div>
        ))}
      </div>
    )
  }

  if (!data) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-slate-800 rounded-xl p-5 border border-slate-700">
            <p className="text-slate-400 text-center p-4">No data available</p>
          </div>
        ))}
      </div>
    )
  }

  const cards = [
    {
      title: 'Total Savings',
      value: formatCurrency(data.total_savings_cents),
      subtitle: `${data.period.start} - ${data.period.end}`,
      icon: TrendingDown,
      gradient: 'from-green-500 to-emerald-600',
      shadow: 'shadow-green-500/20',
      hoverBorder: 'hover:border-green-500/50',
      subtitleColor: 'text-green-400',
    },
    {
      title: 'Ongoing Savings',
      value: formatCurrency(data.ongoing_savings_cents),
      subtitle: 'Currently accumulating',
      icon: Clock,
      gradient: 'from-blue-500 to-indigo-600',
      shadow: 'shadow-blue-500/20',
      hoverBorder: 'hover:border-blue-500/50',
      subtitleColor: 'text-blue-400',
    },
    {
      title: 'Top Savers',
      value: data.top_savers && data.top_savers.length > 0 ? data.top_savers.length.toString() : '0',
      subtitle: data.top_savers && data.top_savers.length > 0 
        ? `Best: ${formatCurrency(data.top_savers[0]?.savings_cents || 0)}`
        : 'No savings yet',
      icon: DollarSign,
      gradient: 'from-purple-500 to-pink-600',
      shadow: 'shadow-purple-500/20',
      hoverBorder: 'hover:border-purple-500/50',
      subtitleColor: 'text-purple-400',
    },
  ]

  return (
    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
      {cards.map((card) => (
        <div
          key={card.title}
          className={`bg-slate-800 rounded-xl p-5 shadow-lg border border-slate-700 ${card.hoverBorder} transition-all group`}
        >
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm text-slate-400 font-medium">{card.title}</p>
            <div className={`p-2 bg-gradient-to-br ${card.gradient} rounded-lg group-hover:scale-105 transition-transform shadow-lg ${card.shadow}`}>
              <card.icon className="h-5 w-5 text-white" />
            </div>
          </div>
          <p className="text-3xl font-bold text-white mb-1">{card.value}</p>
          <p className={`text-sm ${card.subtitleColor}`}>{card.subtitle}</p>
        </div>
      ))}
    </div>
  )
}
