import { authMiddleware } from '@/middleware/auth'
import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/_protected')({
  component: RouteComponent,
  server: {
    middleware: [authMiddleware]
  }
})

function RouteComponent() {
  return <Outlet />
}
