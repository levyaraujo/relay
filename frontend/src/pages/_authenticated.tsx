
import { AuthService } from '@/api/AuthService';
import { TransactionService } from '@/api/TransactionService';
import { dashboardInterval } from '@/lib/const';
import { AppSidebar } from '@components/Sidebar.tsx';
import { ThemeToggle } from '@components/ThemeToggle.tsx';
import { SidebarProvider, SidebarTrigger } from '@components/ui/sidebar.tsx';
import { TooltipProvider } from '@components/ui/tooltip.tsx';
import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';
import axios from 'axios';

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: ({ context, location }) => {
    if (!context.auth?.isAuthenticated) {
      throw redirect({
        to: '/login',
        search: {
          redirect: location.href
        }
      })
    }
  },
  loader: async ({ context: { queryClient, auth } }) => {
    try {
      await queryClient.ensureQueryData({
        queryKey: ['dashboard-summary', dashboardInterval.from, dashboardInterval.to],
        queryFn: () => TransactionService.getDashboardSummary(dashboardInterval.from, dashboardInterval.to),
      })

      return await queryClient.ensureQueryData({
        queryKey: ['user'],
        queryFn: AuthService.getUserData,
      })
    } catch (err) {
      if (axios.isAxiosError(err) && (err.response?.status === 401 || err.response?.status === 403)) {
        auth.logout()
        throw redirect({ to: '/login' })
      }

      throw err
    }
  },
  component: () => <DashboardLayout />,
})


function DashboardLayout() {
  return (
    <SidebarProvider>
      <TooltipProvider>
        <AppSidebar />
        <main className='flex flex-1 flex-col overflow-auto'>
          <header className='flex items-center justify-between p-2'>
            <SidebarTrigger />
            <ThemeToggle />
          </header>
          <div className='flex-1'>
            <Outlet />
          </div>
        </main>
      </TooltipProvider>
    </SidebarProvider>
  )
}
