import { getOptions } from '@/gen/api/@tanstack/react-query.gen'
import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_protected/')({
  component: RouteComponent,
})

function RouteComponent() {
  const { data, isLoading } = useQuery({
    ...getOptions()
  })

  if (isLoading) {
    return <div>Loading...</div>
  }

  return <div>{data}</div>
}
