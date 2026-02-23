import { useState } from 'react';
import { ChevronDown, ChevronUp, Database } from 'lucide-react';
import clsx from 'clsx';
import type { Instance } from '../lib/api';

interface FilterPreviewProps {
  matchedInstances: Instance[];
  totalInstances: number;
  loading?: boolean;
}

export function FilterPreview({ matchedInstances, totalInstances, loading }: FilterPreviewProps) {
  const [expanded, setExpanded] = useState(false);
  const displayLimit = 5;
  const hasMore = matchedInstances.length > displayLimit;

  if (loading) {
    return (
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
        <div className="flex items-center gap-2 text-slate-400">
          <div className="animate-spin h-4 w-4 border-2 border-slate-400 border-t-transparent rounded-full" />
          <span className="text-sm">Loading preview...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-slate-800/50 rounded-lg border border-slate-700 overflow-hidden">
      {/* Header with count */}
      <div className="flex items-center justify-between px-4 py-3 bg-slate-900/50 border-b border-slate-700">
        <div className="flex items-center gap-2">
          <Database className="h-4 w-4 text-slate-400" />
          <span className="text-sm font-medium text-white">
            Preview: {matchedInstances.length} of {totalInstances} instances
          </span>
        </div>
        <span
          className={clsx(
            'px-2.5 py-1 text-xs rounded-full font-medium',
            matchedInstances.length > 0
              ? 'bg-green-500/10 text-green-400 border border-green-500/30'
            : 'bg-slate-500/10 text-slate-400 border border-slate-500/30'
          )}
        >
          {matchedInstances.length > 0 ? `${matchedInstances.length} match` : 'No matches'}
        </span>
      </div>

      {/* Instance list */}
      {matchedInstances.length > 0 ? (
        <div className="divide-y divide-slate-700">
          {(expanded ? matchedInstances : matchedInstances.slice(0, displayLimit)).map(
            (instance) => (
              <div
                key={instance.id}
                className="flex items-center gap-3 px-4 py-2 hover:bg-slate-700/30"
              >
                <div className="flex-shrink-0 h-8 w-8 bg-gradient-to-br from-blue-600 to-cyan-700 rounded-lg flex items-center justify-center">
                  <span className="text-white text-xs font-bold">
                    {instance.name.substring(0, 2).toUpperCase()}
                  </span>
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white truncate">
                    {instance.name}
                  </p>
                  <p className="text-xs text-slate-400 truncate">
                    {instance.provider.startsWith('aws') ? 'AWS' : 'GCP'} · {instance.region} · {instance.engine}
                  </p>
                </div>
                <span
                  className={clsx(
                    'px-2 py-0.5 text-xs rounded-full font-medium',
                    instance.status === 'running' || instance.status === 'available'
                      ? 'bg-green-500/10 text-green-400'
                    : 'bg-slate-500/10 text-slate-400'
                  )}
                >
                  {instance.status}
                </span>
              </div>
            )
          )}

          {/* Show more/less button */}
          {hasMore && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="w-full flex items-center justify-center gap-1 px-4 py-2 text-sm text-indigo-400 hover:text-indigo-300 hover:bg-slate-700/30 transition-colors"
            >
              {expanded ? (
                <>
                  <ChevronUp className="h-4 w-4" />
                  Show less
                </>
              ) : (
                <>
                  <ChevronDown className="h-4 w-4" />
                  Show {matchedInstances.length - displayLimit} more
                </>
              )}
            </button>
          )}
        </div>
      ) : (
        <div className="p-6 text-center">
          <Database className="h-8 w-8 text-slate-500 mx-auto mb-2" />
          <p className="text-sm text-slate-400">No instances match the current filters</p>
          <p className="text-xs text-slate-500 mt-1">
            Add or modify filter rules to select instances
          </p>
        </div>
      )}
    </div>
  );
}

export default FilterPreview;
