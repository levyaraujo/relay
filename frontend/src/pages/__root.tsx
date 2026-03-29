import type { AuthState } from '@/contexts/Auth'
import { dashboardInterval } from '@/lib/const'
import type { QueryClient } from '@tanstack/react-query'
import { createRootRouteWithContext, Link, Outlet } from '@tanstack/react-router'

interface RouterContext {
  auth: AuthState
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => <Outlet />,
  notFoundComponent: NotFound,
})

function NotFound() {
  return (
    <div className='flex min-h-svh flex-col items-center justify-center gap-4 text-center'>
      <h1 className='text-6xl font-bold text-foreground'>404</h1>
      <p className='text-lg text-muted-foreground'>
        This page doesn&apos;t exist.
      </p>
      <Link
        search={ dashboardInterval }
        to='/dashboard'
        className='mt-2 inline-flex items-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90'
      >
        Go to Dashboard
      </Link>
    </div>
  )
}
