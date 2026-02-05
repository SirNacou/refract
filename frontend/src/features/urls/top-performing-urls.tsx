import DataTable from "@/components/data-table"
import ButtonCopy from "@/components/smoothui/button-copy"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer } from "@/components/ui/chart"
import type { TopUrl } from "@/gen/api"
import { type ColumnDef, createColumnHelper } from "@tanstack/react-table"
import dayjs from "dayjs"
import { useId } from "react"
import { Area, AreaChart } from "recharts"

const columnHelper = createColumnHelper<TopUrl>()

const columns = [
  columnHelper.accessor("original_url", {
    header: "Original URL",
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor("short_url", {
    header: "Short URL",
    cell: (info) => {
      const shortUrl = info.getValue()
      return (
        <p className="flex items-center gap-2">
          {shortUrl}
          <ButtonCopy
            loadingDuration={0}
            duration={1000}
            onCopy={() => navigator.clipboard.writeText(shortUrl)}
          />
        </p>
      )
    },
  }),
  columnHelper.accessor("clicks", {
    header: "Clicks",
    cell: (info) => info.getValue().toLocaleString(),
    size: 100,
  }),
  columnHelper.accessor("this_week_trends", {
    header: "This Week Trends",
    size: 100,
    cell: (info) => {
      const trends = [
        ...info.getValue().map((x) => ({ ...x, clicks: Number(x.clicks) })),
        { date: dayjs(Date.now()).add(1, "day"), clicks: 2 },
        { date: dayjs(Date.now()).add(2, "day"), clicks: 1 },
        { date: dayjs(Date.now()).add(3, "day"), clicks: 2 },
      ]

      return (
        <ChartContainer
          className="aspect-auto h-12.5 w-full"
          config={{ clicks: { label: "Clicks", color: "var(--chart-1)" } }}
        >
          <AreaChart
            style={{ width: "100%", height: 50 }}
            data={trends}
            margin={{
              top: 5,
              bottom: 5,
            }}
          >
            <defs>
              <linearGradient id={useId()} x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--chart-1)"
                  stopOpacity={0.3}
                />
                <stop offset="95%" stopColor="var(--chart-1)" stopOpacity={0} />
              </linearGradient>
            </defs>
            <Area
              type={"monotone"}
              dataKey="clicks"
              stroke="url(#colorClicks)"
              fill="url(#colorClicks)"
            />
          </AreaChart>
        </ChartContainer>
      )
    },
  }),
]

type Props = {
  data: TopUrl[]
}

const TopPerformingURLs = ({ data }: Props) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Top Performing URLs</CardTitle>
      </CardHeader>
      <CardContent>
        <DataTable columns={columns as ColumnDef<unknown>[]} data={data} />
      </CardContent>
    </Card>
  )
}

export default TopPerformingURLs
