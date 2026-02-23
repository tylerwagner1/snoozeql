import { useState } from 'react';
import { X } from 'lucide-react';
import clsx from 'clsx';
import type { Selector } from '../lib/api';
import { validateRegex } from '../lib/filterUtils';

interface FilterRuleProps {
  selector: Selector;
  onChange: (selector: Selector) => void;
  onRemove: () => void;
  index: number;
}

type FieldType = 'name' | 'provider' | 'region' | 'engine' | 'tag';
type MatchType = 'exact' | 'contains' | 'prefix' | 'suffix' | 'regex';

const MATCH_TYPES: { value: MatchType; label: string }[] = [
  { value: 'contains', label: 'contains' },
  { value: 'exact', label: 'equals' },
  { value: 'prefix', label: 'starts with' },
  { value: 'suffix', label: 'ends with' },
  { value: 'regex', label: 'matches regex' },
];

export function FilterRule({ selector, onChange, onRemove, index }: FilterRuleProps) {
  // Determine current field type from selector
  const getFieldType = (): FieldType => {
    if (selector.provider) return 'provider';
    if (selector.region?.pattern) return 'region';
    if (selector.engine?.pattern) return 'engine';
    if (selector.tags && Object.keys(selector.tags).length > 0) return 'tag';
    return 'name';
  };

  const [fieldType, setFieldType] = useState<FieldType>(getFieldType());
  const [tagKey, setTagKey] = useState<string>(
    selector.tags ? Object.keys(selector.tags)[0] || '' : ''
  );
  const [regexError, setRegexError] = useState<string>('');

  // Get current matcher based on field type
  const getMatcher = () => {
    switch (fieldType) {
      case 'name':
        return selector.name || { pattern: '', type: 'contains' };
      case 'region':
        return selector.region || { pattern: '', type: 'contains' };
      case 'engine':
        return selector.engine || { pattern: '', type: 'contains' };
      case 'tag':
        return selector.tags?.[tagKey] || { pattern: '', type: 'contains' };
      default:
        return { pattern: '', type: 'contains' };
    }
  };

  const matcher = getMatcher();

  // Update selector when field type changes
  const handleFieldTypeChange = (newType: FieldType) => {
    setFieldType(newType);
    setRegexError('');

    // Create new selector based on new field type
    if (newType === 'provider') {
      onChange({ name: selector.name || { pattern: '', type: 'contains' }, provider: 'aws' });
    } else if (newType === 'name') {
      onChange({ name: { pattern: selector.name?.pattern || '', type: 'contains' } });
    } else if (newType === 'region') {
      onChange({ name: selector.name || { pattern: '', type: 'contains' }, region: { pattern: '', type: 'contains' } });
    } else if (newType === 'engine') {
      onChange({ name: selector.name || { pattern: '', type: 'contains' }, engine: { pattern: '', type: 'contains' } });
    } else if (newType === 'tag') {
      onChange({ name: selector.name || { pattern: '', type: 'contains' }, tags: { [tagKey || 'env']: { pattern: '', type: 'contains' } } });
      if (!tagKey) setTagKey('env');
    }
  };

  // Update match type
  const handleMatchTypeChange = (type: MatchType) => {
    const newMatcher = { pattern: matcher.pattern, type };

    // Validate regex if switching to regex type
    if (type === 'regex' && matcher.pattern) {
      const error = validateRegex(matcher.pattern);
      setRegexError(error);
    } else {
      setRegexError('');
    }

    if (fieldType === 'name') {
      onChange({ ...selector, name: newMatcher });
    } else if (fieldType === 'region') {
      onChange({ ...selector, region: newMatcher });
    } else if (fieldType === 'engine') {
      onChange({ ...selector, engine: newMatcher });
    } else if (fieldType === 'tag') {
      onChange({ ...selector, tags: { [tagKey]: newMatcher } });
    }
  };

  // Update pattern value
  const handlePatternChange = (pattern: string) => {
    // Validate regex if in regex mode
    if (matcher.type === 'regex') {
      const error = validateRegex(pattern);
      setRegexError(error);
    }

    const newMatcher = { pattern, type: matcher.type };

    if (fieldType === 'name') {
      onChange({ ...selector, name: newMatcher });
    } else if (fieldType === 'region') {
      onChange({ ...selector, region: newMatcher });
    } else if (fieldType === 'engine') {
      onChange({ ...selector, engine: newMatcher });
    } else if (fieldType === 'tag') {
      onChange({ ...selector, tags: { [tagKey]: newMatcher } });
    }
  };

  // Update provider selection
  const handleProviderChange = (provider: string) => {
    onChange({ ...selector, provider });
  };

  // Update tag key
  const handleTagKeyChange = (key: string) => {
    setTagKey(key);
    const currentMatcher = selector.tags?.[tagKey] || { pattern: '', type: 'contains' };
    onChange({ ...selector, tags: { [key]: currentMatcher } });
  };

  return (
    <div className="flex flex-col gap-2 bg-slate-700/50 rounded-lg p-3 border border-slate-600">
      <div className="flex items-center gap-2">
        <span className="text-xs text-slate-400 font-medium w-6">#{index + 1}</span>

        {/* Field type dropdown */}
        <select
          value={fieldType}
          onChange={(e) => handleFieldTypeChange(e.target.value as FieldType)}
          className="bg-slate-800 text-sm text-white rounded-lg px-3 py-1.5 border border-slate-600 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
        >
          <option value="name">Instance Name</option>
          <option value="provider">Cloud Provider</option>
          <option value="region">Region</option>
          <option value="engine">Engine</option>
          <option value="tag">Tag</option>
        </select>

        {/* Provider-specific: dropdown for AWS/GCP */}
        {fieldType === 'provider' && (
          <select
            value={selector.provider || 'aws'}
            onChange={(e) => handleProviderChange(e.target.value)}
            className="bg-slate-800 text-sm text-white rounded-lg px-3 py-1.5 border border-slate-600 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
          >
            <option value="aws">AWS</option>
            <option value="gcp">GCP</option>
          </select>
        )}

        {/* Tag-specific: tag key input */}
        {fieldType === 'tag' && (
          <input
            type="text"
            value={tagKey}
            onChange={(e) => handleTagKeyChange(e.target.value)}
            placeholder="tag key"
            className="w-24 bg-slate-900 text-sm text-white rounded-lg px-3 py-1.5 border border-slate-600 focus:ring-2 focus:ring-indigo-500 focus:outline-none placeholder-slate-500"
          />
        )}

        {/* Match type dropdown (not for provider) */}
        {fieldType !== 'provider' && (
          <select
            value={matcher.type}
            onChange={(e) => handleMatchTypeChange(e.target.value as MatchType)}
            className="bg-slate-800 text-sm text-white rounded-lg px-3 py-1.5 border border-slate-600 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
          >
            {MATCH_TYPES.map((mt) => (
              <option key={mt.value} value={mt.value}>
                {mt.label}
              </option>
            ))}
          </select>
        )}

        {/* Pattern input (not for provider) */}
        {fieldType !== 'provider' && (
          <input
            type="text"
            value={matcher.pattern}
            onChange={(e) => handlePatternChange(e.target.value)}
            placeholder={matcher.type === 'regex' ? 'regex pattern...' : 'value...'}
            className={clsx(
              'flex-1 bg-slate-900 text-sm text-white rounded-lg px-3 py-1.5 border focus:ring-2 focus:ring-indigo-500 focus:outline-none placeholder-slate-500',
                  regexError ? 'border-red-500' : 'border-slate-600'
            )}
          />
        )}

        {/* Remove button */}
        <button
          onClick={onRemove}
          className="p-1.5 text-slate-400 hover:text-red-400 hover:bg-slate-600 rounded-lg transition-colors"
          title="Remove rule"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      {/* Regex error message */}
      {regexError && (
        <p className="text-xs text-red-400 pl-8">{regexError}</p>
      )}
    </div>
  );
}

export default FilterRule;
