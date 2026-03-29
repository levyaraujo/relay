import { AuthContext } from '@/contexts/Auth'
import { useContext } from 'react'


export function useAuth() {
  return useContext(AuthContext)
}
