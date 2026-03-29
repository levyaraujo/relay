import { ThemeProvider } from '@/contexts/Theme';
import { useAuth } from '@/hooks/useAuth.ts';
import { router } from '@/router.tsx';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { RouterProvider } from '@tanstack/react-router';
import { Toaster } from 'sonner';


function App() {
  const auth = useAuth()

  return (
    <ThemeProvider defaultTheme='light' storageKey='theme'>
      <RouterProvider router={ router } context={ { auth } } />
      <ReactQueryDevtools initialIsOpen={ false } />
      <Toaster />
    </ThemeProvider>
  )
}

export default App
