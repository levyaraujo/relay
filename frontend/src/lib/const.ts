import { endOfMonth, format, startOfMonth } from 'date-fns';

const today = new Date()

export const dashboardInterval = {
  from: format(startOfMonth(today), 'yyyy-MM-dd'),
  to: format(endOfMonth(today), 'yyyy-MM-dd')
}
