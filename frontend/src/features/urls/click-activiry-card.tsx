import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import dayjs from '@/lib/dayjs-config'
import { Clock, MousePointer2, Plus } from 'lucide-react'
import React from 'react'

export type Activity = {
  id: string
  action: 'Created' | 'Clicked'
  url: string
  time: Date
  location?: string
  device?: string
}

type Props = {
  activities: Activity[]
} & React.ComponentProps<typeof Card>

const ClickActivityCard = ({ activities, ...props }: Props) => {
  return (
    <Card {...props}>
      <CardHeader>
        <CardTitle className='text-lg'>
          Click Activity
        </CardTitle>
      </CardHeader>
      <CardContent>
        {activities.map((activity) => (
          <div key={activity.id} className="relative flex items-start gap-4">
            <div className={`relative z-10 flex h-7 w-7 shrink-0 items-center justify-center rounded-full ring-4 ring-white ${activity.action === 'Created' ? 'bg-emerald-100 text-emerald-600' : 'bg-blue-100 text-blue-600'
              }`}>
              {activity.action === 'Created' ? <Plus size={14} /> : <MousePointer2 size={14} />}
            </div>
            <div className="flex-1 space-y-1">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-slate-900">
                  {activity.action === 'Created' ? 'New Link Created' : 'Link Clicked'}
                </p>
                <span className="flex items-center text-xs text-slate-400">
                  <Clock size={12} className="mr-1" />
                  {dayjs(activity.time).fromNow()}
                </span>
              </div>
              <p className="text-sm text-slate-500">
                <span className="font-medium text-indigo-600">{activity.url}</span>
                {activity.action !== 'Created' && (
                  <span className="ml-1">
                    from {activity.location} via {activity.device}
                  </span>
                )}
              </p>
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}

export default ClickActivityCard