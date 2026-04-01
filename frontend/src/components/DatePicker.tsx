'use client'

import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Field } from '@/components/ui/field'
import { Popover, PopoverContent, PopoverTrigger, } from '@/components/ui/popover'
import { Route } from '@/pages/_authenticated/dashboard'
import { format, parseISO } from 'date-fns'
import { CalendarIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { type DateRange } from 'react-day-picker'
import { ptBR } from 'react-day-picker/locale'

export function DatePicker() {
  const { from, to } = Route.useSearch()
  const navigate = Route.useNavigate()

  const [date, setDate] = useState<DateRange | undefined>({
    from: parseISO(from as string),
    to: parseISO(to),
  })

  useEffect(() => {
    setDate({ from: parseISO(from as string), to: parseISO(to) })
  }, [from, to])

  useEffect(() => {
    if (date?.from && date?.to) {
      const [start, end] = date.from <= date.to
        ? [date.from, date.to]
        : [date.to, date.from]
      const newFrom = format(start, 'yyyy-MM-dd')
      const newTo = format(end, 'yyyy-MM-dd')
      if (newFrom !== from || newTo !== to) {
        navigate({ search: { from: newFrom, to: newTo } })
      }
    }
  }, [date, from, to, navigate])

  return (
    <Field className='mx-auto w-60'>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            variant='outline'
            id='date-picker-range'
            className='justify-start px-2.5 font-normal'
          >
            <CalendarIcon />
            { date?.from ? (
              date.to ? (
                <>
                  { format(date.from, 'LLL dd, y') } -{ ' ' }
                  { format(date.to, 'LLL dd, y') }
                </>
              ) : (
                format(date.from, 'LLL dd, y')
              )
            ) : (
              <span>Pick a date</span>
            ) }
          </Button>
        </PopoverTrigger>
        <PopoverContent className='w-auto p-0' align='start'>
          <Calendar
            mode='range'
            defaultMonth={ date?.from }
            selected={ date }
            onSelect={ setDate }
            numberOfMonths={ 2 }
            showOutsideDays={ false }
            locale={ ptBR }
          />
        </PopoverContent>
      </Popover>
    </Field>
  )
}
