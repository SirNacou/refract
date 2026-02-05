import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartConfig, ChartContainer } from "@/components/ui/chart"
import { Area, AreaChart, CartesianGrid, Tooltip, XAxis, YAxis } from "recharts"

const chartConfig = {
  clicks: {
    label: 'Clicks',
    color: 'var(--chart-1)',
  }
} as ChartConfig

type ClickTrend = {
  clicks: number
  date: Date
}

type Props = {
  data: ClickTrend[]
} & React.ComponentProps<typeof Card>

const ClickTrendsChartCard = ({ data, ...props }: Props) => {
  return (
    <Card {...props}>
      <CardHeader>
        <CardTitle className='text-lg'>
          Click Trends
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig}>
          <AreaChart data={data} margin={{ top: 10, right: 50, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="colorClicks" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" />
            <XAxis
              dataKey="date"
              tickFormatter={(v) => v instanceof Date ? Intl.DateTimeFormat("en-GB").format(v) : ''}
              axisLine={false}
              tickLine={false}
              tick={{ fill: '#64748b', fontSize: 14 }}
              dy={10}
              minTickGap={30}
            />
            <YAxis
              axisLine={false}
              tickLine={false}
              tick={{ fill: '#64748b', fontSize: 14 }}
            />
            <Tooltip
              contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
            />
            <Area
              type="monotone"
              dataKey="clicks"
              stroke="#6366f1"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#colorClicks)"
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}

export default ClickTrendsChartCard