import type { DoubleChart } from '@/lib/types/charts'
import api from './client'

export class ChartService {
  static async getCashFlowChartData(from: string, to: string, group: 'day' | 'week' | 'month') {
    const { data } = await api.get<DoubleChart[]>(`/api/charts/cash-flow?from=${from}&to=${to}&group=${group}`)

    return data
  }
}
