import { ChartBarMultiple } from '@/components/charts/DoubleChart'
import { DatePicker } from '@/components/DatePicker'
import { useCashFlow } from '@/hooks/useCharts'
import { useDashboardSummary } from '@/hooks/useTransactions.ts'
import { dashboardInterval } from '@/lib/const'
import { formatReal } from '@/lib/formatters'
import type { DashboardSummary, Transaction } from '@/lib/types/transaction.ts'
import { Separator } from '@components/ui/separator'
import { createFileRoute } from '@tanstack/react-router'
import { ArrowDownRight, ArrowUpRight, BanknoteArrowDown, BanknoteArrowUp, DollarSign, TrendingUp, } from 'lucide-react'
import { useState } from 'react'


export const Route = createFileRoute('/_authenticated/dashboard')({
  validateSearch: (search: Record<string, unknown>): { from: string; to: string } => ({
    from: typeof search.from === 'string' ? search.from : dashboardInterval.from,
    to: typeof search.to === 'string' ? search.to : dashboardInterval.to,
  }),
  component: Dashboard,
})

interface StatCardProps {
  title: string
  value: string
  trend: string
  trendUp: boolean
  icon: React.ElementType
}

function StatCard({ title, value, trend, trendUp, icon: Icon }: StatCardProps) {
  return (
    <div className='flex flex-col gap-2 rounded-xl border border-border bg-card p-5'>
      <div className='flex items-center justify-between'>
        <span className='text-xs font-medium text-muted-foreground'>{ title }</span>
        <Icon className='size-4 text-muted-foreground' />
      </div>
      <span className='text-2xl font-bold'>{ value }</span>
      <div className='flex items-center gap-1 text-xs'>
        { trendUp
          ? <ArrowUpRight className='size-3 text-green-500' />
          : <ArrowDownRight className='size-3 text-red-500' /> }
        <span className={ trendUp ? 'text-green-500' : 'text-red-500' }>{ trend }</span>
        <span className='text-muted-foreground'>vs mês anterior</span>
      </div>
    </div>
  )
}
const transactions: Transaction[] = [
  { description: 'Pagamento Fornecedor A', category: 'Fornecedores', type: 'debit', amount: 'R$ 4.200', dueDate: '25/03' },
  { description: 'Recebimento Cliente X', category: 'Vendas', type: 'credit', amount: 'R$ 8.500', dueDate: '26/03' },
  { description: 'Aluguel Escritório', category: 'Infraestrutura', type: 'debit', amount: 'R$ 3.800', dueDate: '28/03' },
  { description: 'Consultoria Projeto Y', category: 'Serviços', type: 'credit', amount: 'R$ 12.000', dueDate: '30/03' },
  { description: 'Folha de pagamento', category: 'Pessoal', type: 'debit', amount: 'R$ 15.400', dueDate: '05/04' },
]

function TransactionsTable() {
  return (
    <div className='flex flex-col gap-0 rounded-xl border border-border bg-card'>
      <div className='flex items-center justify-between px-5 pt-5 pb-3'>
        <h2 className='text-base font-semibold'>Transações Recentes</h2>
        <button className='rounded-lg border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-muted'>
          Ver tudo
        </button>
      </div>
      <div className='grid grid-cols-[1fr_100px_100px_90px] gap-x-4 px-5 pb-2 text-xs font-medium text-muted-foreground'>
        <span>Descrição</span>
        <span>Tipo</span>
        <span>Valor</span>
        <span>Vencimento</span>
      </div>
      { transactions.map(tx => (
        <div key={ tx.description }>
          <Separator />
          <div className='grid grid-cols-[1fr_100px_100px_90px] gap-x-4 px-5 py-3 text-sm'>
            <div className='flex flex-col'>
              <span>{ tx.description }</span>
              <span className='text-xs text-muted-foreground'>{ tx.category }</span>
            </div>
            <span className={ tx.type === 'debit' ? 'text-red-500' : 'text-green-500' }>
              { tx.type === 'debit' ? 'Débito' : 'Crédito' }
            </span>
            <span className={ `font-semibold ${tx.type === 'debit' ? 'text-red-500' : 'text-green-500'}` }>
              { tx.amount }
            </span>
            <span className='text-muted-foreground'>{ tx.dueDate }</span>
          </div>
        </div>
      )) }
    </div>
  )
}

function CashFlowChart() {
  const rangeConfig = {
    '7d':  { days: 7, group: 'day' },
    '30d': { days: 30, group: 'day' },
    '90d': { days: 90, group: 'week' },
    '12m': { days: 365, group: 'month' },
  } as const

  const [rangeKey, setRangeKey] = useState<keyof typeof rangeConfig>('7d')
  const { cashFlow } = useCashFlow(rangeConfig[rangeKey])

  return (
    <ChartBarMultiple
      title='Fluxo de Caixa'
      description='Por intervalo'
      data={ cashFlow ?? [] }
      labelKey='label'
      left={ { key: 'income', label: 'Receita', color: '#22c55e' } }
      right={ { key: 'expense', label: 'Despesa', color: '#ef4444' } }
      valueFormatter={ formatReal }
      ranges={ Object.keys(rangeConfig) }
      activeRange={ rangeKey }
      onRangeChange={ (r) => setRangeKey(r as keyof typeof rangeConfig) }
    />
  )
}

function UpcomingPayments() {
  const items = [
    { name: 'Fornecedor A', date: '25/03', amount: 'R$ 4.200' },
    { name: 'Aluguel', date: '28/03', amount: 'R$ 3.800' },
    { name: 'Internet', date: '30/03', amount: 'R$ 280' },
    { name: 'Energia', date: '02/04', amount: 'R$ 650' },
  ]

  return (
    <div className='flex flex-col gap-0 rounded-xl border border-border bg-card p-5'>
      <h2 className='text-base font-semibold'>Próximos Vencimentos</h2>
      { items.map(item => (
        <div key={ item.name }>
          <Separator className='my-2' />
          <div className='flex items-center justify-between py-1'>
            <div className='flex flex-col'>
              <span className='text-sm'>{ item.name }</span>
              <span className='text-xs text-muted-foreground'>{ item.date }</span>
            </div>
            <span className='text-sm font-semibold text-red-500'>{ item.amount }</span>
          </div>
        </div>
      )) }
    </div>
  )
}

function AlertsPanel() {
  const alerts = [
    { title: 'Fatura vencida', desc: 'Fornecedor A — R$ 2.100', date: 'Venceu 20/03', severity: 'red' as const },
    { title: 'Fatura vencida', desc: 'Aluguel — R$ 1.900', date: 'Venceu 22/03', severity: 'red' as const },
    { title: 'Vence amanhã', desc: 'Internet — R$ 850', date: 'Vence 25/03', severity: 'amber' as const },
  ]

  return (
    <div className='flex flex-col gap-0 rounded-xl border border-border bg-card p-5'>
      <div className='flex items-center justify-between'>
        <h2 className='text-base font-semibold'>Alertas</h2>
        <span className='rounded-md bg-amber-500/15 px-2 py-0.5 text-xs font-medium text-amber-500'>
          { alerts.length }
        </span>
      </div>
      { alerts.map(alert => (
        <div key={ alert.desc }>
          <Separator className='my-2' />
          <div className='flex gap-3 py-1'>
            <div className={ `mt-1.5 size-2 shrink-0 rounded-full ${alert.severity === 'red' ? 'bg-red-500' : 'bg-amber-500'}` } />
            <div className='flex flex-col'>
              <span className={ `text-sm font-medium ${alert.severity === 'red' ? 'text-red-500' : 'text-amber-500'}` }>
                { alert.title }
              </span>
              <span className='text-sm'>{ alert.desc }</span>
              <span className='text-xs text-muted-foreground'>{ alert.date }</span>
            </div>
          </div>
        </div>
      )) }
    </div>
  )
}

function Dashboard() {
  const { dashboardSummary } = useDashboardSummary() as { dashboardSummary: DashboardSummary }

  const { totalIncome, totalExpenses, balance, totalTransactions } = dashboardSummary


  return (
    <div className='flex flex-1 flex-col gap-6 p-2'>
      <div>
        <h1 className='text-2xl font-bold text-primary'>Dashboard</h1>
        <p className='text-sm text-muted-foreground'>Visão geral do seu negócio</p>
        <DatePicker />
      </div>

      <div className='grid grid-cols-4 gap-4'>
        <StatCard title='Receita (mês)' value={ formatReal(totalIncome) } trend='+12%' trendUp icon={ BanknoteArrowDown } />
        <StatCard title='Despesa (mês)' value={ formatReal(totalExpenses) } trend='-3%' trendUp={ false } icon={ BanknoteArrowUp } />
        <StatCard title='Saldo projetado' value={ formatReal(balance) } trend='Fim do mês' trendUp icon={ DollarSign } />
        <StatCard title='Lançamentos' value={ totalTransactions.toString() } trend='+8%' trendUp icon={ TrendingUp } />
      </div>

      <div className='grid grid-cols-[1fr_320px] gap-4'>
        <CashFlowChart />
        <AlertsPanel />
      </div>

      <div className='grid grid-cols-[1fr_320px] gap-4'>
        <TransactionsTable />
        <UpcomingPayments />
      </div>
    </div>
  )
}
