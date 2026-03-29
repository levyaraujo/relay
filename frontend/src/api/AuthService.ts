import api from '@/api/client.ts';
import type { Login, LoginResponse } from '@/lib/types/login.ts';
import type { User } from '@/lib/types/user';
import { AxiosError } from 'axios';

export class AuthService {
  static async login(login: Login): Promise<LoginResponse> {
    try {
      const { data } = await api.post<LoginResponse>('/auth/login', login)
      return data
    } catch (err) {
      if (err instanceof AxiosError) {
        if (err.status === 401) {
          throw new Error('Invalid credentials');
        }
      }

      if (err instanceof AxiosError && !err.response) {
        throw new Error('Unable to connect. Please try again.')
      }
      throw err
    }
  }

  static async getUserData() {
    try {
      const { data } = await api.get<User>('/api/users/me')

      return data

    } catch {
      throw new Error('Unable to get user')
    }
  }
}
