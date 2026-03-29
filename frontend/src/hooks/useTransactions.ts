import { useQuery } from '@tanstack/react-query'
import { TransactionService } from '@/api/TransactionService.ts';
import type { Transaction } from '@/lib/types/transaction.ts';
import { Route } from '@/pages/_authenticated/dashboard.tsx';

const minutes = 1000 * 60


export function useTransactions() {
  const { from, to } = Route.useSearch()

  const { data: transactions, error, isError } = useQuery<Transaction[]>({
    queryKey: ['transactions', from, to],
    queryFn: () => TransactionService.getTransactionsByDateRange(from, to),
    staleTime: 2 * minutes,
  })

  return {
    transactions,
    error,
    isError
  }
}
