'use client'

import { TrendingUp } from 'lucide-react'
import { Bar, BarChart, CartesianGrid, XAxis } from 'recharts'

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import { Button } from '../ui/button'

interface ChartSeries<T> {
  key: keyof T & string
  label?: string
  color?: string
}

interface ChartBarMultipleProps<T> {
  data: T[]
  title: string
  description: string
  labelKey: keyof T & string
  left: ChartSeries<T>
  right: ChartSeries<T>
  valueFormatter?: (value: number) => string
  ranges?: string[]
  activeRange?: string
  onRangeChange?: (range: string) => void
}

export function ChartBarMultiple<T>({ data, title, description, labelKey, left, right, valueFormatter, ranges, activeRange, onRangeChange }: ChartBarMultipleProps<T>) {

  const chartConfig = {
    [left.key]: { label: left.label ?? left.key, color: left.color ?? 'var(--chart-1)' },
    [right.key]: { label: right.label ?? right.key, color: right.color ?? 'var(--chart-2)' },
  } satisfies ChartConfig

  return (
    <Card>
      <CardHeader className='flex flex-row justify-between'>
        <div>
          <CardTitle>{ title }</CardTitle>
          <CardDescription>{ description }</CardDescription>
        </div>
        { ranges && (
          <div className='flex gap-1'>
            { ranges.map((range) => (
              <Button
                key={ range }
                variant={ range === activeRange ? 'default' : 'outline' }
                size='sm'
                onClick={ () => onRangeChange?.(range) }
              >
                { range }
              </Button>
            )) }
          </div>
        ) }
      </CardHeader>
      <CardContent className='px-0'>
        <ChartContainer config={ chartConfig } className='h-80 w-full'>
          <BarChart accessibilityLayer data={ data }>
            <CartesianGrid vertical={ false } />
            <XAxis
              dataKey={ labelKey as string }
              tickLine={ false }
              tickMargin={ 10 }
              axisLine={ false }
              tickFormatter={ (value) => value.slice(0, 3) }
            />
            <ChartTooltip
              cursor={ false }
              content={ <ChartTooltipContent indicator='dashed' formatter={ valueFormatter ? (value, name) => (
                <span className='flex flex-1 items-center justify-between gap-2'>
                  <span className='text-muted-foreground'>{ chartConfig[name as string]?.label ?? name }</span>
                  <span className='font-medium tabular-nums text-right'>{ valueFormatter(value as number) }</span>
                </span>
              ) : undefined } /> }
            />
            <Bar dataKey={ left.key as string } fill={ `var(--color-${left.key})` } radius={ 4 } />
            <Bar dataKey={ right.key as string } fill={ `var(--color-${right.key})` } radius={ 4 } />
          </BarChart>
        </ChartContainer>
      </CardContent>
      <CardFooter className='flex-col items-start gap-2 text-sm'>
        <div className='flex gap-2 leading-none font-medium'>
          Trending up by 5.2% this month <TrendingUp className='h-4 w-4' />
        </div>
        <div className='leading-none text-muted-foreground'>
          Showing total visitors for the last 6 months
        </div>
      </CardFooter>
    </Card>
  )
}
