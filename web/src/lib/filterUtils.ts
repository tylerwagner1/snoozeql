import type { Instance, Selector } from './api';

// Match types that mirror backend models.MatchType
export type MatchType = 'exact' | 'contains' | 'prefix' | 'suffix' | 'regex';

export interface Matcher {
  pattern: string;
  type: MatchType;
}

/**
 * Check if an instance matches a set of selectors
 * @param instance The instance to check
 * @param selectors Array of selectors to match against
 * @param operator 'and' (all must match) or 'or' (any must match)
 * @returns true if instance matches
 */
export function matchInstance(
  instance: Instance,
  selectors: Selector[],
  operator: 'and' | 'or' = 'and'
): boolean {
  if (selectors.length === 0) {
    return false; // Require at least one selector
  }

  for (const selector of selectors) {
    const matches = matchSelector(instance, selector);
    if (operator === 'or' && matches) {
      return true;
    }
    if (operator === 'and' && !matches) {
      return false;
    }
  }

  return operator === 'and';
}

/**
 * Check if an instance matches a single selector
 * All non-null fields in the selector must match (AND within selector)
 */
export function matchSelector(instance: Instance, selector: Selector): boolean {
  // Check name matcher
  if (selector.name?.pattern) {
    if (!matchField(instance.name, selector.name)) {
      return false;
    }
  }

  // Check provider (exact match: "aws" or "gcp")
  if (selector.provider) {
    const instanceProvider = instance.provider.startsWith('aws') ? 'aws' : 'gcp';
    if (selector.provider !== instanceProvider) {
      return false;
    }
  }

  // Check region matcher
  if (selector.region?.pattern) {
    if (!matchField(instance.region, selector.region)) {
      return false;
    }
  }

  // Check engine matcher
  if (selector.engine?.pattern) {
    if (!matchField(instance.engine, selector.engine)) {
      return false;
    }
  }

  // Check tags (all specified tags must match)
  if (selector.tags) {
    for (const [tagKey, matcher] of Object.entries(selector.tags)) {
      if (!matcher?.pattern) continue;
      const tagValue = instance.tags?.[tagKey];
      if (!tagValue || !matchField(tagValue, matcher)) {
        return false;
      }
    }
  }

  return true;
}

/**
 * Apply a matcher pattern to a string value
 */
export function matchField(
  value: string,
  matcher: { pattern: string; type: string }
): boolean {
  if (!matcher || !matcher.pattern) {
    return true;
  }

  switch (matcher.type) {
    case 'exact':
      return value === matcher.pattern;
    case 'contains':
      return value.toLowerCase().includes(matcher.pattern.toLowerCase());
    case 'prefix':
      return value.toLowerCase().startsWith(matcher.pattern.toLowerCase());
    case 'suffix':
      return value.toLowerCase().endsWith(matcher.pattern.toLowerCase());
    case 'regex':
      try {
        const re = new RegExp(matcher.pattern, 'i');
        return re.test(value);
      } catch {
        return false;
      }
    default:
      return false;
  }
}

/**
 * Validate a regex pattern
 * @returns Error message if invalid, empty string if valid
 */
export function validateRegex(pattern: string): string {
  if (!pattern) return '';
  try {
    new RegExp(pattern);
    return '';
  } catch (e) {
    return e instanceof Error ? e.message : 'Invalid regex';
  }
}

/**
 * Create an empty selector with default values
 */
export function createEmptySelector(): Selector {
  return {
    name: { pattern: '', type: 'contains' },
  };
}

/**
 * Get human-readable description of a selector
 */
export function describeSelectorRule(selector: Selector): string {
  const parts: string[] = [];

  if (selector.name?.pattern) {
    parts.push(`name ${describeMatchType(selector.name.type)} "${selector.name.pattern}"`);
  }
  if (selector.provider) {
    parts.push(`provider is ${selector.provider.toUpperCase()}`);
  }
  if (selector.region?.pattern) {
    parts.push(`region ${describeMatchType(selector.region.type)} "${selector.region.pattern}"`);
  }
  if (selector.engine?.pattern) {
    parts.push(`engine ${describeMatchType(selector.engine.type)} "${selector.engine.pattern}"`);
  }
  if (selector.tags) {
    for (const [key, matcher] of Object.entries(selector.tags)) {
      if (matcher?.pattern) {
        parts.push(`tag "${key}" ${describeMatchType(matcher.type)} "${matcher.pattern}"`);
      }
    }
  }

  return parts.length > 0 ? parts.join(' AND ') : 'No conditions';
}

function describeMatchType(type: string): string {
  switch (type) {
    case 'exact':
      return 'equals';
    case 'contains':
      return 'contains';
    case 'prefix':
      return 'starts with';
    case 'suffix':
      return 'ends with';
    case 'regex':
      return 'matches';
    default:
      return type;
  }
}
