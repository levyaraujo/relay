import { DatePicker } from '@/components/DatePicker'
import { dashboardInterval } from '@/lib/const'
import { Separator } from '@components/ui/separator'
import { createFileRoute } from '@tanstack/react-router'
import {
  ArrowDownRight,
  ArrowUpRight,
  BanknoteArrowDown,
  BanknoteArrowUp,
  DollarSign,
  TrendingUp,
} from 'lucide-react'
import { useTransactions } from '@/hooks/useTransactions.ts';
import type { Transaction } from '@/lib/types/transaction.ts';



export const Route = createFileRoute('/_authenticated/dashboard')({
  validateSearch: (search: Record<string, unknown>) => ({
    from: (search.from) ?? dashboardInterval.from,
    to: (search.to as string) ?? dashboardInterval.to,
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
  const bars = [65, 45, 80, 35, 70, 50, 90, 40, 60, 55, 85, 48]
  const months = ['Jan', 'Fev', 'Mar', 'Abr', 'Mai', 'Jun', 'Jul', 'Ago', 'Set', 'Out', 'Nov', 'Dez']

  return (
    <div className='flex flex-col gap-4 rounded-xl border border-border bg-card p-5'>
      <div className='flex items-center justify-between'>
        <h2 className='text-base font-semibold'>Fluxo de Caixa</h2>
        <div className='flex gap-1'>
          { ['7D', '30D', '90D', '12M'].map(period => (
            <button
              key={ period }
              className={ `rounded-md px-2.5 py-1 text-xs font-medium ${period === '30D' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:bg-muted/50'}` }
            >
              { period }
            </button>
          )) }
        </div>
      </div>
      <div className='flex items-end gap-2 rounded-lg bg-background p-4' style={ { height: 200 } }>
        { bars.map((h, i) => (
          <div key={ months[i] } className='flex flex-1 flex-col items-center gap-1'>
            <div className='flex w-full items-end gap-0.5' style={ { height: 140 } }>
              <div
                className='flex-1 rounded-sm bg-green-500/70'
                style={ { height: `${h}%` } }
              />
              <div
                className='flex-1 rounded-sm bg-red-500/70'
                style={ { height: `${h * 0.7}%` } }
              />
            </div>
            <span className='text-[10px] text-muted-foreground'>{ months[i] }</span>
          </div>
        )) }
      </div>
    </div>
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
  const currency = Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' })
  const { transactions } = useTransactions() as { transactions: Transaction[] }

  const totalIncome = transactions
    ?.filter(transaction => transaction.type === 'credit')
    .reduce((acc, cur) => acc + cur.amount, 0) as number

  const totalExpenses = transactions
    ?.filter(transaction => transaction.type === 'debit')
    .reduce((acc, cur) => acc + cur.amount, 0) as number

  const balance = totalIncome- totalExpenses
  const totalTransactions = transactions.length


  return (
    <div className='flex flex-1 flex-col gap-6 p-2'>
      <div>
        <h1 className='text-2xl font-bold text-primary'>Dashboard</h1>
        <p className='text-sm text-muted-foreground'>Visão geral do seu negócio</p>
        <DatePicker />
      </div>

      <div className='grid grid-cols-4 gap-4'>
        <StatCard title='Receita (mês)' value={ `${currency.format(totalIncome)}` } trend='+12%' trendUp icon={ BanknoteArrowDown } />
        <StatCard title='Despesa (mês)' value={ `${currency.format(totalExpenses)}` } trend='-3%' trendUp={ false } icon={ BanknoteArrowUp } />
        <StatCard title='Saldo projetado' value={ currency.format(balance) } trend='Fim do mês' trendUp icon={ DollarSign } />
        <StatCard title='Transações' value={ totalTransactions.toString() } trend='+8%' trendUp icon={ TrendingUp } />
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
