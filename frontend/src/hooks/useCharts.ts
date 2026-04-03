import { useQuery } from '@tanstack/react-query'

import { ChartService } from '@/api/ChartService'
import type { DoubleChart } from '@/lib/types/charts'
import { subDays } from 'date-fns'
import { dateEndISO, dateStartISO, minutes } from '@/lib/dates'

export type CashFlowRange = {
  days: number
  group: 'day' | 'week' | 'month'
}

export function useCashFlow(range: CashFlowRange) {
  const { days, group } = range

  const today = new Date()
  const toDate = today.toISOString().split('T')[0]
  const fromDate = subDays(today, days).toISOString().split('T')[0]

  const to = dateEndISO(toDate)
  const from = dateStartISO(fromDate)

  const { data, isError, error, isLoading } = useQuery<DoubleChart[]>({
    queryKey: ['cash-flow', days, group],
    queryFn: () => ChartService.getCashFlowChartData(from, to, group),
    staleTime: 2 * minutes
  })

  return {
    cashFlow: data,
    isError,
    error,
    isLoading
  }
}
