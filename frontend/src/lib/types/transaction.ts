export interface Transaction {
  id: string
  company: string
  user: string
  account: string
  vendor: string
  type: 'credit' | 'debit'
  amount: number
  description: string
  created_at: string
  due_date: string
  paid_date: string
  origin: string
}

export interface DashboardSummary {
  totalIncome: number
  totalExpenses: number
  balance: number
  totalTransactions: number
}
