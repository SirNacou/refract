import StatisticCard from '@/components/statistic-card'
import AddURLDialog from '@/features/urls/add-url-dialog'
import ClickActivityCard from '@/features/urls/click-activiry-card'
import ClickTrendsChartCard from '@/features/urls/click-trends-chart-card'
import { createFileRoute } from '@tanstack/react-router'
import { AxeIcon } from 'lucide-react'

const data = [
  {
    date: '2024-10-01', clicks: 120,
  },
  {
    date: '2024-10-02', clicks: 200,
  },
  {
    date: '2024-10-03', clicks: 150,
  },
  {
    date: '2024-10-04', clicks: 80,
  },
  {
    date: '2024-10-05', clicks: 70,
  },
]

export const Route = createFileRoute('/_protected/(main)/')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div className='flex flex-col gap-3'>
    <div className='flex justify-between'>
      <div className='text-2xl font-bold'>Dashboard</div>
      <AddURLDialog />
    </div>

    <div className='grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6'>
      <StatisticCard
        title='Test'
        value={'100'} icon={AxeIcon} iconColor={''} />
      <StatisticCard
        title='Test'
        value={'100'} icon={AxeIcon} iconColor={''} />
      <StatisticCard
        title='Test'
        value={'100'} icon={AxeIcon} iconColor={''} />
      <StatisticCard
        title='Test'
        value={'100'} icon={AxeIcon} iconColor={''} />

      <ClickTrendsChartCard data={data} className='col-span-4 md:col-span-2 xl:col-span-3' />

      <ClickActivityCard className='col-span-4 md:col-span-2 xl:col-span-1' activities={[
        {
          id: "1",
          action: 'Clicked',
          time: new Date(),
          url: 'https://example.com/abc',
          location: 'New York, USA',
          device: 'Chrome on Windows'
        }
      ]} />
    </div>
  </div>
}
