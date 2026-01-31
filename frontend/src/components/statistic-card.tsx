import { cn } from '@/lib/utils'
import { Badge } from './ui/badge'
import { Card, CardContent } from './ui/card'

type Props = {
  title: string
  value: string | number
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>
  iconColor: string
  badge?: {
    text: string
    color: string
    icon: React.ComponentType<React.SVGProps<SVGSVGElement>>
    iconColor: string
  }
}

const StatisticCard = ({ title, badge, icon: Icon, iconColor, value }: Props) => {
  return (
    <Card>
      <CardContent className="flex flex-col h-full">
        {/* Title & Badge */}
        <div className="flex items-center justify-between mb-8">
          <Icon className={cn('size-6', iconColor)} />

          {badge &&
            <Badge className={cn('px-2 py-1 rounded-full', badge.color)}>
              <badge.icon className={`w-3 h-3 ${badge.iconColor}`} />
              {badge.text}
            </Badge>}
        </div>

        {/* Value & Date Range */}
        <div className="flex-1 flex flex-col justify-between grow">
          {/* Value */}
          <div>
            <div className="text-base font-medium text-muted-foreground mb-1">{title}</div>
            <div className="text-3xl font-bold text-foreground mb-6">{value.toLocaleString()}</div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export default StatisticCard