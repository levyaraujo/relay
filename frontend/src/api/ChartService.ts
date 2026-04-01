import type { DoubleChart } from '@/lib/types/charts'
import api from './client'

export class ChartService {
  static async getCashFlowChartData(from: string, to: string) {
    const fromRFC = new Date(from + 'T00:00:00Z').toISOString()
    const toRFC = new Date(to + 'T23:59:59Z').toISOString()
    const { data } = await api.get<{ cashFlow: DoubleChart[] }>(`/api/charts/cash-flow?from=${fromRFC}&to=${toRFC}`)

    return data
  }
}
