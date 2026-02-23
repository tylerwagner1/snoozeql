/**
 * Format cents as a currency string (e.g., 152345 -> "$1,523.45")
 * Uses Intl.NumberFormat for proper locale handling
 */
export function formatCurrency(cents: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(cents / 100)
}

/**
 * Format a number of hours as a readable string (e.g., 24.5 -> "24.5h")
 */
export function formatHours(hours: number): string {
  return `${hours.toFixed(1)}h`
}
