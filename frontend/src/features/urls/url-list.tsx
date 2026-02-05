import DataTable from "@/components/data-table"
import ButtonCopy from "@/components/smoothui/button-copy"
import { Badge } from "@/components/ui/badge"
import type { Url } from "@/gen/api/types.gen"
import { type ColumnDef, createColumnHelper } from "@tanstack/react-table"

type Props = {
  data: Url[]
  loading?: boolean
}

const columnHelper = createColumnHelper<Url>()
const columns = [
  columnHelper.accessor("original_url", {
    header: "Original URL",
    cell: ({ getValue }) => (
      <span className="block max-w-[320px] truncate sm:max-w-105 lg:max-w-130">
        {getValue()}
      </span>
    ),
  }),
  columnHelper.accessor("short_url", {
    header: "Short URL",
    cell: ({ getValue }) => (
      <p className="flex items-center justify-between">
        {getValue()}
        <ButtonCopy
          onCopy={() => navigator.clipboard.writeText(getValue())}
          loadingDuration={0}
          duration={1000}
        />
      </p>
    ),
  }),
  columnHelper.accessor("status", {
    header: "Status",
    cell: ({ getValue }) => {
      const status = getValue()
      if (status === "active") {
        return (
          <Badge className="bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300 font-medium capitalize">
            {status}
          </Badge>
        )
      } else if (status === "expired") {
        return (
          <Badge className="bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300 font-medium capitalize">
            {status}
          </Badge>
        )
      } else {
        return (
          <Badge className="bg-orange-50 text-orange-700 dark:bg-orange-950 dark:text-orange-300 font-medium capitalize">
            {status}
          </Badge>
        )
      }
    },
  }),
  columnHelper.accessor("created_at", {
    header: "Created At",
    cell: ({ getValue }) =>
      new Intl.DateTimeFormat("en-GB", {}).format(getValue()),
    sortingFn: "datetime",
  }),
  columnHelper.accessor("expires_at", {
    header: "Expires At",
    cell: ({ getValue }) => {
      const value = getValue()
      return value
        ? new Intl.DateTimeFormat("en-GB", {}).format(value)
        : "Never"
    },
    sortingFn: "datetime",
  }),
] as ColumnDef<Url>[]

const URLList = ({ data, loading }: Props) => {
  return <DataTable data={data} columns={columns} loading={loading} />
}

export default URLList
