/**
 * Calendar date strings (YYYY-MM-DD) → UTC instants for API query bounds.
 */

/** Start of the given calendar day in UTC, as ISO 8601. */
export function dateStartISO(dateOnly: string): string {
  return new Date(`${dateOnly}T00:00:00Z`).toISOString()
}

/** End of the given calendar day in UTC (23:59:59), as ISO 8601. */
export function dateEndISO(dateOnly: string): string {
  return new Date(`${dateOnly}T23:59:59Z`).toISOString()
}


export const minutes = 1000 * 60
