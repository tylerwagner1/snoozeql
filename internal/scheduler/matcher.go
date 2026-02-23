package scheduler

import (
	"regexp"
	"strings"

	"snoozeql/internal/models"
)

// MatchInstance checks if an instance matches a set of selectors.
// operator is "and" (all selectors must match) or "or" (any selector must match).
// Empty selectors returns false (require explicit selection).
func MatchInstance(instance *models.Instance, selectors []models.Selector, operator string) bool {
	if len(selectors) == 0 {
		return false // Require at least one selector
	}

	for _, sel := range selectors {
		matches := MatchSelector(instance, &sel)
		if operator == "or" && matches {
			return true
		}
		if operator == "and" && !matches {
			return false
		}
	}

	return operator == "and"
}

// MatchSelector checks if an instance matches a single selector.
// All non-nil fields in the selector must match (AND within selector).
func MatchSelector(instance *models.Instance, sel *models.Selector) bool {
	// Check name matcher
	if sel.Name != nil {
		if !matchMatcher(instance.Name, sel.Name) {
			return false
		}
	}

	// Check provider (exact match: "aws" or "gcp")
	if sel.Provider != nil {
		instanceProvider := "gcp"
		if strings.HasPrefix(instance.Provider, "aws") {
			instanceProvider = "aws"
		}
		if *sel.Provider != instanceProvider {
			return false
		}
	}

	// Check region matcher
	if sel.Region != nil {
		if !matchMatcher(instance.Region, sel.Region) {
			return false
		}
	}

	// Check engine matcher
	if sel.Engine != nil {
		if !matchMatcher(instance.Engine, sel.Engine) {
			return false
		}
	}

	// Check tags (all specified tags must match)
	if sel.Tags != nil {
		for tagKey, matcher := range sel.Tags {
			tagValue, ok := instance.Tags[tagKey]
			if !ok {
				return false // Tag key doesn't exist
			}
			if !matchMatcher(tagValue, matcher) {
				return false
			}
		}
	}

	return true
}

// matchMatcher applies a Matcher to a string value
func matchMatcher(value string, matcher *models.Matcher) bool {
	if matcher == nil {
		return true
	}

	switch matcher.Type {
	case models.MatchExact:
		return value == matcher.Pattern
	case models.MatchContains:
		return strings.Contains(value, matcher.Pattern)
	case models.MatchPrefix:
		return strings.HasPrefix(value, matcher.Pattern)
	case models.MatchSuffix:
		return strings.HasSuffix(value, matcher.Pattern)
	case models.MatchRegex:
		re, err := regexp.Compile(matcher.Pattern)
		if err != nil {
			return false // Invalid regex doesn't match
		}
		return re.MatchString(value)
	default:
		return false
	}
}

// ValidateSelectors checks if all selectors have valid patterns.
// Returns error message if invalid, empty string if valid.
func ValidateSelectors(selectors []models.Selector) string {
	for i, sel := range selectors {
		if sel.Name != nil && sel.Name.Type == models.MatchRegex {
			if _, err := regexp.Compile(sel.Name.Pattern); err != nil {
				return "Invalid regex in selector " + string(rune(i+1)) + " name: " + err.Error()
			}
		}
		if sel.Region != nil && sel.Region.Type == models.MatchRegex {
			if _, err := regexp.Compile(sel.Region.Pattern); err != nil {
				return "Invalid regex in selector " + string(rune(i+1)) + " region: " + err.Error()
			}
		}
		if sel.Engine != nil && sel.Engine.Type == models.MatchRegex {
			if _, err := regexp.Compile(sel.Engine.Pattern); err != nil {
				return "Invalid regex in selector " + string(rune(i+1)) + " engine: " + err.Error()
			}
		}
		for tagKey, matcher := range sel.Tags {
			if matcher != nil && matcher.Type == models.MatchRegex {
				if _, err := regexp.Compile(matcher.Pattern); err != nil {
					return "Invalid regex in selector " + string(rune(i+1)) + " tag '" + tagKey + "': " + err.Error()
				}
			}
		}
	}
	return ""
}
