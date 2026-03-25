import { ThemeProvider } from '@/contexts/Theme'
import { AppSidebar } from '@components/Sidebar'
import { SidebarProvider, SidebarTrigger } from '@components/ui/sidebar'
import { Outlet } from '@tanstack/react-router'
import { ThemeToggle } from './components/ThemeToggle'

function App() {

  return (
    <ThemeProvider defaultTheme='dark' storageKey='theme'>
      <SidebarProvider>
        <AppSidebar />
        <SidebarTrigger />
        <Outlet />
      </SidebarProvider>
      <ThemeToggle />
    </ThemeProvider>
  )
}

export default App
