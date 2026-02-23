import { useState, useEffect, useMemo } from 'react';
import { Plus, Filter } from 'lucide-react';
import type { Instance, Selector } from '../lib/api';
import api from '../lib/api';
import { FilterRule } from './FilterRule';
import { FilterPreview } from './FilterPreview';
import { matchInstance, createEmptySelector } from '../lib/filterUtils';

interface FilterBuilderProps {
  selectors: Selector[];
  onChange: (selectors: Selector[]) => void;
  /** Optional: pre-fetched instances for client-side preview */
  instances?: Instance[];
}

export function FilterBuilder({ selectors, onChange, instances: propInstances }: FilterBuilderProps) {
  const [operator, setOperator] = useState<'and' | 'or'>('and');
  const [instances, setInstances] = useState<Instance[]>(propInstances || []);
  const [loading, setLoading] = useState(!propInstances);

  // Fetch instances if not provided
  useEffect(() => {
    if (!propInstances) {
      setLoading(true);
      api.getInstances()
        .then(setInstances)
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [propInstances]);

  // Update instances if props change
  useEffect(() => {
    if (propInstances) {
      setInstances(propInstances);
    }
  }, [propInstances]);

  // Compute matched instances client-side for instant preview
  const matchedInstances = useMemo(() => {
    if (selectors.length === 0) return [];
    return instances.filter((instance) => matchInstance(instance, selectors, operator));
  }, [instances, selectors, operator]);

  const addRule = () => {
    onChange([...selectors, createEmptySelector()]);
  };

  const updateRule = (index: number, selector: Selector) => {
    const updated = [...selectors];
    updated[index] = selector;
    onChange(updated);
  };

  const removeRule = (index: number) => {
    onChange(selectors.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Filter className="h-4 w-4 text-slate-400" />
          <h3 className="text-sm font-medium text-white">Instance Filters</h3>
        </div>

        {/* AND/OR toggle */}
        {selectors.length > 1 && (
          <div className="flex items-center gap-2">
            <span className="text-xs text-slate-400">Combine with:</span>
            <div className="flex rounded-lg overflow-hidden border border-slate-600">
              <button
                onClick={() => setOperator('and')}
                className={`px-3 py-1 text-xs font-medium transition-colors ${
                  operator === 'and'
                    ? 'bg-indigo-600 text-white'
                  : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                }`}
              >
                AND
              </button>
              <button
                onClick={() => setOperator('or')}
                className={`px-3 py-1 text-xs font-medium transition-colors ${
                  operator === 'or'
                    ? 'bg-indigo-600 text-white'
                  : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                }`}
              >
                OR
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Filter rules */}
      {selectors.length > 0 ? (
        <div className="space-y-2">
          {selectors.map((selector, index) => (
            <div key={index}>
              {index > 0 && (
                <div className="flex items-center justify-center py-1">
                  <span className="text-xs font-medium text-indigo-400 bg-slate-800 px-2 py-0.5 rounded">
                    {operator.toUpperCase()}
                  </span>
                </div>
              )}
              <FilterRule
                selector={selector}
                onChange={(sel) => updateRule(index, sel)}
                onRemove={() => removeRule(index)}
                index={index}
              />
            </div>
          ))}
        </div>
      ) : (
        <div className="text-center py-4 bg-slate-800/30 rounded-lg border border-dashed border-slate-600">
          <p className="text-sm text-slate-400 mb-2">No filters added yet</p>
          <p className="text-xs text-slate-500">
            Add filters to select which instances this schedule applies to
          </p>
        </div>
      )}

      {/* Add rule button */}
      <button
        onClick={addRule}
        className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-indigo-400 bg-slate-800 hover:bg-slate-700 rounded-lg border border-slate-600 transition-colors"
      >
        <Plus className="h-4 w-4" />
        Add Filter Rule
      </button>

      {/* Preview panel */}
      <FilterPreview
        matchedInstances={matchedInstances}
        totalInstances={instances.length}
        loading={loading}
      />
    </div>
  );
}

export default FilterBuilder;
