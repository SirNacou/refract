import DataTable from '@/components/data-table'
import { type Url } from '@/gen/api/types.gen'
import { createColumnHelper } from '@tanstack/react-table'

type Props = {
  data: Url[]
  loading?: boolean
}

const columnHelper = createColumnHelper<Url>()
const columns = [
  columnHelper.accessor('OriginalURL', {
    header: 'Original URL',
  }),
  columnHelper.accessor('ShortCode', {
    header: 'Short Code',
  })
]

const URLList = ({ data, loading }: Props) => {
  return (
    <DataTable data={data} columns={columns} loading={loading} />
  )
}

export default URLList