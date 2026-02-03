import DataTable from '@/components/data-table'
import ButtonCopy from '@/components/smoothui/button-copy'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { TopUrl } from '@/gen/api'
import { ColumnDef, createColumnHelper } from '@tanstack/react-table'

const columnHelper = createColumnHelper<TopUrl>()

const columns = [
  columnHelper.accessor('original_url', {
    header: 'Original URL',
    cell: info => info.getValue(),
  }),
  columnHelper.accessor('short_url', {
    header: 'Short URL',
    cell: info => {
      const shortUrl = info.getValue()
      return <p className='flex items-center gap-2'>
        {shortUrl}
        <ButtonCopy loadingDuration={0} duration={1000} onCopy={() => navigator.clipboard.writeText(shortUrl)} />
      </p>
    },
  }),
  columnHelper.accessor('clicks', {
    header: 'Clicks',
    cell: info => info.getValue().toLocaleString(),
    size: 100
  }),
  columnHelper.accessor('this_week_trends', {
    header: 'This Week Trends',
    cell: info => {
      const trends = info.getValue()
      const totalClicks = trends.reduce((acc, trend) => acc + Number(trend.clicks), 0)
      return `${totalClicks} clicks`
    },
    size: 200
  })
] as ColumnDef<any>[]

type Props = {
  data: TopUrl[]
}

const TopPerformingURLs = ({ data }: Props) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle className='text-lg'>
          Top Performing URLs
        </CardTitle>
      </CardHeader>
      <CardContent>
        <DataTable columns={columns} data={data} />
      </CardContent>
    </Card>
  )
}

export default TopPerformingURLs