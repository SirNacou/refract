import StatisticCard from "@/components/statistic-card"
import AddURLDialog from "@/features/urls/add-url-dialog"
import ClickActivityCard, {
  type Activity,
} from "@/features/urls/click-activiry-card"
import ClickTrendsChartCard from "@/features/urls/click-trends-chart-card"
import TopPerformingURLs from "@/features/urls/top-performing-urls"
import { dashboardOptions } from "@/gen/api/@tanstack/react-query.gen"
import { useQuery } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import FluentOpen32Filled from "~icons/fluent/open-32-filled"
import LucideChartColumn from "~icons/lucide/chart-column"
import LucideLink from "~icons/lucide/link"
import LucideMousePointerClick from "~icons/lucide/mouse-pointer-click"

export const Route = createFileRoute("/_protected/(main)/")({
  component: RouteComponent,
})

function RouteComponent() {
  const { data } = useQuery({
    ...dashboardOptions(),
  })

  return (
    <div className="flex flex-col gap-3">
      <div className="flex justify-between">
        <div className="text-2xl font-bold">Dashboard</div>
        <AddURLDialog />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        <StatisticCard
          title="Total URLs"
          value={Number(data?.total_urls ?? 0)}
          icon={LucideLink}
          iconColor={"text-blue-500"}
        />
        <StatisticCard
          title="Total Clicks"
          value={Number(data?.total_clicks ?? 0)}
          icon={LucideMousePointerClick}
          iconColor={"text-purple-500"}
        />
        <StatisticCard
          title="Active URLs"
          value={Number(data?.active_urls ?? 0)}
          icon={FluentOpen32Filled}
          iconColor={"text-green-500"}
        />
        <StatisticCard
          title="Clicks This Week"
          value={Number(data?.clicks_this_week ?? 0)}
          icon={LucideChartColumn}
          iconColor={"text-orange-500"}
        />

        <ClickTrendsChartCard
          data={
            data?.click_trends.map((x) => ({
              date: x.date,
              clicks: Number(x.clicks),
            })) ?? []
          }
          className="col-span-4 md:col-span-2 xl:col-span-3"
        />

        <ClickActivityCard
          className="col-span-4 md:col-span-2 xl:col-span-1"
          activities={[
            {
              id: "1",
              action: "Clicked",
              time: new Date(),
              url: "https://example.com/abc",
              location: "New York, USA",
              device: "Chrome on Windows",
            },
            ...(data?.recent_activities.map(
              (x) =>
                ({
                  id: x.timestamp.toISOString(),
                  action: "Clicked",
                  time: x.timestamp,
                  url: x.activity,
                  location: "Unknown",
                  device: "Unknown",
                }) as Activity,
            ) ?? []),
          ]}
        />

        <div className="col-span-1 md:col-span-2 xl:col-span-4">
          <TopPerformingURLs data={data?.top_urls ?? []} />
        </div>
      </div>
    </div>
  )
}
