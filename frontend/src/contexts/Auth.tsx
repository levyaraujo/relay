import { AuthService } from '@/api/AuthService.ts';
import { dashboardInterval } from '@/lib/const';
import type { Login, LoginResponse } from '@/lib/types/login.ts';
import { router } from '@/router';
import { type UseMutateFunction, useMutation } from '@tanstack/react-query';
import { createContext, useState } from 'react';

export interface AuthState {
  isAuthenticated: boolean;
  login: UseMutateFunction<LoginResponse, Error, Login>;
  logout: () => void;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
}

export const AuthContext = createContext({} as AuthState);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() => {
    return !!localStorage.getItem('accessToken');
  });

  const { mutate, isPending, isError, error } = useMutation({
    mutationFn: AuthService.login,
    onSuccess: (data) => {
      setIsAuthenticated(true)
      localStorage.setItem('accessToken', data.AccessToken)
      router.invalidate()
      router.navigate({ to: '/dashboard', search: dashboardInterval })
    }
  })

  const logout = () => {
    localStorage.removeItem('accessToken');
    setIsAuthenticated(false);
    router.invalidate()
    router.navigate({ to: '/login' })
  }

  const value: AuthState = {
    isAuthenticated,
    login: mutate,
    logout,
    isLoading: isPending,
    isError,
    error,
  }

  return (
    <AuthContext.Provider value={ value }>
      { children }
    </AuthContext.Provider>
  )
}
