import DataTable from '@/components/data-table'
import { type Url } from '@/gen/api/types.gen'
import { createColumnHelper } from '@tanstack/react-table'

type Props = {}

const urls: Url[] = [
  {
    ID: BigInt(1),
    OriginalURL: 'https://example.com/original1',
    ShortCode: 'https://short.ly/abc123',
    Domain: 'short.ly',
    CreatedAt: new Date('2024-01-01T00:00:00Z'),
    ExpiresAt: new Date('2024-12-31T23:59:59Z'),
    Status: 'active',
    Title: 'Example URL 1',
    Notes: 'First example URL',
    UpdatedAt: new Date('2024-01-02T00:00:00Z'),
    UserID: 'user-123',
  },
  {
    ID: BigInt(2),
    OriginalURL: 'https://example.com/original2',
    ShortCode: 'https://short.ly/def456',
    Domain: 'short.ly',
    CreatedAt: new Date('2024-02-01T00:00:00Z'),
    ExpiresAt: null,
    Status: 'active',
    Title: 'Example URL 2',
    Notes: 'Second example URL',
    UpdatedAt: new Date('2024-02-02T00:00:00Z'),
    UserID: 'user-456',
  },
]

const columnHelper = createColumnHelper<Url>()
const columns = [
  columnHelper.accessor('OriginalURL', {
    header: 'Original URL',
  }),
  columnHelper.accessor('ShortCode', {
    header: 'Short Code',
  })
]

const URLList = (_: Props) => {
  return (
    <DataTable data={urls} columns={columns} />
  )
}

export default URLList