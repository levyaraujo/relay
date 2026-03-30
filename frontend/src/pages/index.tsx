import { dashboardInterval } from '@/lib/const'
import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/')({
  beforeLoad: () => {
    throw redirect({ to: '/dashboard', search: dashboardInterval })
  },
})