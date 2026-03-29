import { AuthService } from '@/api/AuthService'
import type { User } from '@/lib/types/user'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { toast } from 'sonner'

const hours = 1000 * 60 * 60


export function useUserInfo() {
  const { data: user, error, isError } = useQuery<User>({
    queryKey: ['user'],
    queryFn: AuthService.getUserData,
    staleTime: 24 * hours,
    throwOnError: false
  })


  useEffect(() => {
    if (isError && error) toast.error(error.message)
  }, [error])


  return {
    user,
    error
  }
}
