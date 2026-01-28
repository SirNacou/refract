import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { type Url } from '@/gen/api/types.gen'
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'

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
  const table = useReactTable({
    data: urls,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <div className='overflow-hidden rounded-md border'>
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                        header.column.columnDef.header,
                        header.getContext()
                      )}
                  </TableHead>
                )
              })}
            </TableRow>
          ))}
        </TableHeader>

        <TableBody>
          {table.getRowModel().rows?.length ? (
            table.getRowModel().rows.map((row) => (
              <TableRow
                key={row.id}
                data-state={row.getIsSelected() && "selected"}
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell colSpan={columns.length} className="h-24 text-center">
                No results.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  )
}

export default URLList