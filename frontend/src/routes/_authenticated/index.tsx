import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/')({
  component: RouteComponent,
  beforeLoad: () => {
    throw redirect({
      to: '/dashboard',
      replace: true,
    })
  }
})

function RouteComponent() {
  return <div>Hello "/_authenticated/"!</div>
}
