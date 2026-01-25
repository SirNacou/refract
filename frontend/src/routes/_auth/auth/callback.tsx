import { account } from '@/lib/appwrite'
import { createFileRoute } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/_auth/auth/callback')({
  component: RouteComponent,
  beforeLoad: async () => {
  }
})

function RouteComponent() {
  const navigate = Route.useNavigate()
  useEffect(() => {
    async function handleCallback() {
      try {
        await account.get()
        navigate({
          to: '/',
          replace: true,
        })

      } catch (error) {
        console.log('Error during OAuth callback processing:', error)
        navigate({
          to: '/login',
          replace: true,
        })
      }
    }
    handleCallback()
  }, [])
  return <div></div>
}
