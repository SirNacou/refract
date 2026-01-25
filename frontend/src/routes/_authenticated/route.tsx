import { account } from '@/lib/appwrite'
import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated')({
  component: RouteComponent,
  beforeLoad: async () => {
    try {
      await account.get()
    } catch (error) {
      throw redirect({
        to: '/login'
      })
    }
  }
})

function RouteComponent() {
  return <Outlet />
}
