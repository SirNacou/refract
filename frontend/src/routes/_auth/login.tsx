import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { account } from '@/lib/appwrite'
import { Route as CallbackRoute } from '@/routes/_auth/auth/callback'
import { createFileRoute, useRouterState } from '@tanstack/react-router'
import { OAuthProvider } from 'appwrite'
import { useState } from 'react'
import LucideArrowRight from '~icons/lucide/arrow-right'
import LucideEye from '~icons/lucide/eye'
import LucideEyeOff from '~icons/lucide/eye-off'
import LucideLockKeyhole from '~icons/lucide/lock-keyhole'
import LucideMail from '~icons/lucide/mail'
import MdiGithub from '~icons/mdi/github'

export const Route = createFileRoute('/_auth/login')({
  component: RouteComponent,
})

function RouteComponent() {
  const [isVisible, setIsVisible] = useState<boolean>(false)

  const toggleVisibility = () => setIsVisible((prevState) => !prevState)

  const routerState = useRouterState()

  function socialLogin() {
    try {
      account.createOAuth2Session({
        provider: OAuthProvider.Github,
        success: new URL(CallbackRoute.path, routerState.location.url.origin).href,
        failure: new URL(Route.path, routerState.location.url.origin).href,
      })
    } catch (error) {
      console.log('Error during social login:', error)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="mx-auto w-full max-w-xs space-y-6">
        <div className="space-y-2 text-center">
          {/* <Logo className="mx-auto h-16 w-16" /> */}
          <h1 className="text-3xl font-semibold">Welcome back</h1>
          <p className="text-muted-foreground">
            Sign in to access to your dashboard, settings and projects.
          </p>
        </div>

        <div className="space-y-5">
          <Button variant="outline" className="w-full justify-center gap-2" onClick={(e) => {
            e.preventDefault()
            socialLogin()
          }}>
            <MdiGithub className="h-4 w-4" />
            Sign in with Github
          </Button>

          <div className="flex items-center gap-2">
            <Separator className="flex-1" />
            <span className="text-sm text-muted-foreground">
              or sign in with email
            </span>
            <Separator className="flex-1" />
          </div>

          <div className="space-y-6">
            <div>
              <Label htmlFor="email">Email</Label>
              <div className="relative mt-2.5">
                <Input
                  id="email"
                  className="peer ps-9"
                  placeholder="ephraim@blocks.so"
                  type="email"
                />
                <div className="text-muted-foreground/80 pointer-events-none absolute inset-y-0 start-0 flex items-center justify-center ps-3 peer-disabled:opacity-50">
                  <LucideMail className='size-4' aria-hidden="true" />
                </div>
              </div>
            </div>

            <div>
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Password</Label>
                <a href="#" className="text-sm text-primary hover:underline">
                  Forgot Password?
                </a>
              </div>
              <div className="relative mt-2.5">
                <Input
                  id="password"
                  className="ps-9 pe-9"
                  placeholder="Enter your password"
                  type={isVisible ? "text" : "password"}
                />
                <div className="text-muted-foreground/80 pointer-events-none absolute inset-y-0 start-0 flex items-center justify-center ps-3 peer-disabled:opacity-50">
                  <LucideLockKeyhole className='size-4' aria-hidden="true" />
                </div>
                <button
                  className="text-muted-foreground/80 hover:text-foreground focus-visible:border-ring focus-visible:ring-ring/50 absolute inset-y-0 end-0 flex h-full w-9 items-center justify-center rounded-e-md transition-[color,box-shadow] outline-none focus:z-10 focus-visible:ring-[3px] disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50"
                  type="button"
                  onClick={toggleVisibility}
                  aria-label={isVisible ? "Hide password" : "Show password"}
                  aria-pressed={isVisible}
                  aria-controls="password"
                >
                  {isVisible ? (
                    <LucideEyeOff className='size-4' aria-hidden="true" />
                  ) : (
                    <LucideEye className='size-4' aria-hidden="true" />
                  )}
                </button>
              </div>
            </div>

            <div className="flex items-center gap-2 pt-1">
              <Checkbox id="remember-me" />
              <Label htmlFor="remember-me">Remember</Label>
            </div>
          </div>

          <Button className="w-full">
            Sign in
            <LucideArrowRight className="h-4 w-4" />
          </Button>

          <div className="text-center text-sm">
            No account?{" "}
            <a href="#" className="text-primary font-medium hover:underline">
              Create an account
            </a>
          </div>
        </div>
      </div>
    </div>
  )
}
