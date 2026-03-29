import { queryClient } from '@/api/client';
import { routeTree } from '@/routeTree.gen.ts';
import { createRouter } from '@tanstack/react-router';


export const router = createRouter({
  routeTree,
  context: {
    queryClient,
    auth: undefined!,
  },
  notFoundMode: 'root'
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
