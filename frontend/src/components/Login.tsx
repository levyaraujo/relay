import { useAuth } from '@/hooks/useAuth.ts'
import { cn } from '@components/lib/utils'
import { Button } from '@components/ui/button'
import { Card, CardContent } from '@components/ui/card'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel, FieldSeparator,
} from '@components/ui/field'
import { Input } from '@components/ui/input'
import { Toaster } from '@components/ui/sonner'
import { zodResolver } from '@hookform/resolvers/zod'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

const schema =  z.object({
  email: z.email({
    error: 'Email is required',
  }),
  password: z.string().min(1, 'Password is required'),
})

type LoginData = z.infer<typeof schema>

export function Login({
  className,
  ...props
}: React.ComponentProps<'div'>) {
  const { login, isError, isLoading, error } = useAuth()
  const { handleSubmit, register, formState } = useForm<LoginData>({
    resolver: zodResolver(schema)
  })

  const handleUserLogin = handleSubmit(async(data) => {
    login(data)
  })

  useEffect(() => {
    if (isError && error) toast.error(error.message)
  }, [isError, error])

  return (
    <div className={ cn('flex min-h-svh flex-col justify-center gap-6 m-auto w-5/10 lg:w-6/10 md:w-8/10 sm:w-9/10', className) } { ...props }>
      <Toaster />
      <Card className='overflow-hidden p-0'>
        <CardContent className='grid p-0 md:grid-cols-2'>
          <form className='p-6 md:p-8' onSubmit={ handleUserLogin }>
            <FieldGroup>
              <div className='flex flex-col items-center gap-2 text-center'>
                <h1 className='text-2xl font-bold'>relay</h1>
                <p className='text-balance text-muted-foreground'>
                  control your business in a simple manner
                </p>
              </div>

              <Field>
                <FieldLabel htmlFor='email'>Email</FieldLabel>
                <div>
                  <Input
                    id='email'
                    type='email'
                    { ...register('email') }
                    placeholder='email@email.com'
                  />
                  { formState.errors.email && (
                    <small className='text-destructive'>
                      { formState.errors.email.message as string }
                    </small>
                  ) }
                </div>
              </Field>
              <Field>
                <div className='flex items-center'>
                  <FieldLabel htmlFor='password'>Password</FieldLabel>
                  <a
                    href='#'
                    className='ml-auto text-sm underline-offset-2 hover:underline'
                  >
                    Forgot your password?
                  </a>
                </div>
                <div>
                  <Input
                    id='password'
                    type='password'
                    { ...register('password') }
                  />
                  { formState.errors.password && (
                    <small className='text-destructive'>
                      { formState.errors.password.message as string }
                    </small>
                  ) }
                </div>
              </Field>

              <Field>
                <Button type='submit' className='cursor-pointer' disabled={ isLoading }>{
                  isLoading ? 'Logging in...' : 'Login'
                }</Button>
              </Field>
              <FieldSeparator className='*:data-[slot=field-separator-content]:bg-card'>
                Or continue with
              </FieldSeparator>
              <Field className='grid grid-cols-3 gap-4'>
                <Button variant='outline' type='button'>
                  <svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'>
                    <path
                      d='M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z'
                      fill='currentColor'
                    />
                  </svg>
                  <span className='sr-only'>Login with Google</span>
                </Button>
                <Button variant='outline' type='button'>
                  <svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'>
                    <path
                      d='M12 0C5.37 0 0 5.37 0 12c0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61-.546-1.385-1.335-1.755-1.335-1.755-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23A11.52 11.52 0 0 1 12 5.803c1.02.005 2.047.138 3.006.404 2.29-1.552 3.297-1.23 3.297-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.605-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 21.795 24 17.298 24 12c0-6.63-5.37-12-12-12z'
                      fill='currentColor'
                    />
                  </svg>
                  <span className='sr-only'>Login with GitHub</span>
                </Button>
              </Field>
              <FieldDescription className='text-center'>
                Don&apos;t have an account? <a href='#'>Sign up</a>
              </FieldDescription>
            </FieldGroup>
          </form>
          <div className='relative hidden bg-muted md:block'>
            <img
              src='https://ui.shadcn.com/placeholder.svg'
              alt='Image'
              className='absolute inset-0 h-full w-full object-cover dark:brightness-[0.2] dark:grayscale'
            />
          </div>
        </CardContent>
      </Card>
      <FieldDescription className='px-6 text-center'>
        By clicking continue, you agree to our <a href='#'>Terms of Service</a>{ ' ' }
        and <a href='#'>Privacy Policy</a>.
      </FieldDescription>
    </div>
  )
}
