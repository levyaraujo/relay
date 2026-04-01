import type { DashboardSummary, Transaction } from '@/lib/types/transaction.ts';
import api from '@/api/client.ts';

export class TransactionService {
  static async getTransactionsByDateRange(from: string, to: string): Promise<Transaction[]> {
    const fromRFC = new Date(from + 'T00:00:00Z').toISOString()
    const toRFC = new Date(to + 'T23:59:59Z').toISOString()
    const { data } = await api.get<{ transactions: Transaction[] }>(`/api/transactions?from=${fromRFC}&to=${toRFC}`)

    return data.transactions
  }

  static async getDashboardSummary(from: string, to: string): Promise<DashboardSummary> {
    const fromRFC = new Date(from + 'T00:00:00Z').toISOString()
    const toRFC = new Date(to + 'T23:59:59Z').toISOString()
    const { data } = await api.get<DashboardSummary>(`/api/dashboard/summary?from=${fromRFC}&to=${toRFC}`)

    return data
  }
}
