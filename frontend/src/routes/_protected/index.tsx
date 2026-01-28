import AddURLDialog from '@/features/urls/add-url-dialog'
import URLList from '@/features/urls/url-list'
import { listUrlsOptions } from '@/gen/api/@tanstack/react-query.gen'
import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_protected/')({
  component: RouteComponent,
})

function RouteComponent() {
  const { data, isLoading } = useQuery({
    ...listUrlsOptions()
  })

  if (isLoading) {
    return <div>Loading...</div>
  }

  return <div className='flex flex-col gap-3'>
    <div className='flex justify-between'>
      <div className='text-2xl font-bold'>URL List</div>
      <AddURLDialog />
    </div>

    <URLList />
  </div>
}
