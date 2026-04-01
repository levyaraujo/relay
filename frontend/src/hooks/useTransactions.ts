import { useQuery } from '@tanstack/react-query'
import { TransactionService } from '@/api/TransactionService.ts';
import type { DashboardSummary } from '@/lib/types/transaction.ts';
import { Route } from '@/pages/_authenticated/dashboard.tsx';

const minutes = 1000 * 60


export function useDashboardSummary() {
  const { from, to } = Route.useSearch()

  const { data: dashboardSummary, error, isError } = useQuery<DashboardSummary>({
    queryKey: ['dashboard-summary', from, to],
    queryFn: () => TransactionService.getDashboardSummary(from, to),
    staleTime: 2 * minutes,
  })

  return {
    dashboardSummary,
    error,
    isError
  }
}
