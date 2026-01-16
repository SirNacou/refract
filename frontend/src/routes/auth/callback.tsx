import { getZitadel } from '@/lib/auth'
import { createFileRoute } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/auth/callback')({
  component: RouteComponent,
  ssr: false
})

function RouteComponent() {
  const navigate = Route.useNavigate()
  useEffect(() => {
    (async () => {
      const zitadel = getZitadel()
      await zitadel.userManager.signinRedirectCallback()
      navigate({
        to: '/',
        replace: true
      })
    })()
  }, [])
  return <div>Hello "/auth/callback"!</div>
}
