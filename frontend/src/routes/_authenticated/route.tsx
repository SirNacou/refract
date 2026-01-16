import { getZitadel } from '@/lib/auth'
import { createFileRoute, Outlet } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/_authenticated')({
  component: RouteComponent,
  ssr: false,
})

function RouteComponent() {
  useEffect(() => {
    // Only initialize on client side
    const auth = getZitadel()

    // Check for existing user session
    auth.userManager.getUser().then((user) => {
      console.log(user)
      if (!user || user.expired) {
        auth.authorize()
      }
    })
  }, [])
  return <Outlet />
}
