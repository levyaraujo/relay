'use client'

import * as React from 'react'

import { useUserInfo } from '@/hooks/useUser'
import { NavMain } from '@components/ui/nav-main'
import { NavProjects } from '@components/ui/nav-projects'
import { NavUser } from '@components/ui/nav-user'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@components/ui/sidebar'
import { TeamSwitcher } from '@components/ui/team-switcher'
import { AudioLinesIcon, FileSpreadsheet, FrameIcon, GalleryVerticalEndIcon, MapIcon, PieChartIcon, TerminalIcon, Wallet } from 'lucide-react'


const data = {
  teams: [
    {
      name: 'Acme Inc',
      logo: <GalleryVerticalEndIcon />,
      plan: 'Enterprise',
    },
    {
      name: 'Acme Corp.',
      logo: <AudioLinesIcon />,
      plan: 'Startup',
    },
    {
      name: 'Evil Corp.',
      logo: (
        <TerminalIcon
        />
      ),
      plan: 'Free',
    },
  ],
  navMain: [
    {
      title: 'Financeiro',
      url: '#',
      icon: <Wallet />,
      isActive: true,
      items: [
        {
          title: 'A receber',
          url: '/receivables',
        },
        {
          title: 'A pagar',
          url: '/payables',
        },
      ],
    },
    {
      title: 'Relatórios',
      url: '#',
      icon: <FileSpreadsheet />,
      isActive: true,
      items: [
        {
          title: 'DRE',
          url: '/reports/dre',
        },
        {
          title: 'Fluxo de Caixa',
          url: '/reports/cash-flow',
        },
        {
          title: 'Vencimentos',
          url: '/reports/aging',
        },
        {
          title: 'Balanço Patrimonial',
          url: '/reports/balance-sheet',
        },
      ],
    },
  ],
  projects: [
    {
      name: 'Design Engineering',
      url: '#',
      icon: (
        <FrameIcon
        />
      ),
    },
    {
      name: 'Sales & Marketing',
      url: '#',
      icon: (
        <PieChartIcon
        />
      ),
    },
    {
      name: 'Travel',
      url: '#',
      icon: (
        <MapIcon
        />
      ),
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { user } = useUserInfo()


  return (
    <Sidebar collapsible='icon' { ...props }>
      <SidebarHeader>
        <TeamSwitcher teams={ data.teams } />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={ data.navMain } />
        <NavProjects projects={ data.projects } />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={ user } />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
