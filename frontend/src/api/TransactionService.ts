import type { DashboardSummary, Transaction } from '@/lib/types/transaction.ts'
import { dateEndISO, dateStartISO } from '@/lib/dates'
import api from '@/api/client.ts'

export class TransactionService {
  static async getTransactionsByDateRange(from: string, to: string): Promise<Transaction[]> {
    const fromRFC = dateStartISO(from)
    const toRFC = dateEndISO(to)
    const { data } = await api.get<{ transactions: Transaction[] }>(`/api/transactions?from=${fromRFC}&to=${toRFC}`)

    return data.transactions
  }

  static async getDashboardSummary(from: string, to: string): Promise<DashboardSummary> {
    const fromRFC = dateStartISO(from)
    const toRFC = dateEndISO(to)
    const { data } = await api.get<DashboardSummary>(`/api/dashboard/summary?from=${fromRFC}&to=${toRFC}`)

    return data
  }
}
