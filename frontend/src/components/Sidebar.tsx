import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import { Link } from '@tanstack/react-router'
import { BanknoteArrowDown, BanknoteArrowUp, Compass, LayoutDashboard } from 'lucide-react'

export function AppSidebar() {

  return (
    <Sidebar collapsible="icon" className='flex flex-row justify-center'>
      <Header />
      <SidebarContent>
        <Dashboard />
        <IncomeOutcome />
        <SidebarGroup />
      </SidebarContent>
      <SidebarFooter />
    </Sidebar>
  )
}


function Dashboard() {
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton asChild>
          <Link to='/dashboard'>
            <LayoutDashboard />
            Dashboard
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}

function IncomeOutcome() {
  const data = [
    {
      name: 'A pagar',
      icon: BanknoteArrowUp,
      url: '/payables'
    },
    {
      name: 'A receber',
      icon: BanknoteArrowDown,
      url: '/receivables'
    }
  ]

  return (
    <SidebarGroup>
      <SidebarGroupLabel>Entradas e Saídas</SidebarGroupLabel>
      <SidebarMenu>
        {data.map(item => (
          <SidebarMenuItem key={item.name}>
            <SidebarMenuButton asChild>
              <Link to={item.url}>
                <item.icon />
                <span>{item.name}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  )
}

function Header() {
  return (
    <SidebarHeader className='pl-3'>
      <Compass />
    </SidebarHeader>
  )
}
