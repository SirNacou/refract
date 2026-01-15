import { createFileRoute } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/auth/callback')({
  component: RouteComponent,
})

function RouteComponent() {
  useEffect(() => {
    // zitadel.userManager.signinRedirectCallback()
    //   .then(user => console.log('Logged in user:', user))
  }, [])
  return <div>Hello "/auth/callback"!</div>
}
