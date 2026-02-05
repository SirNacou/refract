import { cn } from '@/lib/utils'
import { Card, CardContent } from './ui/card'

type Props = {
  title: string
  value: string | number
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>
  iconColor: string
  className?: string
  badge?: {
    text: string
    color: string
    icon: React.ComponentType<React.SVGProps<SVGSVGElement>>
    iconColor: string
  }
}

const StatisticCard = ({ title, badge, icon: Icon, iconColor, value, className }: Props) => {
  return (
    <Card className={className}>
      <CardContent className="flex items-center h-full">
        {/* Title & Badge */}

        <div className="flex-1 flex flex-col justify-between grow">
          {/* Value */}
          <div className="text-base font-medium text-muted-foreground mb-1">{title}</div>
          <div className="text-3xl font-bold text-foreground">{value.toLocaleString()}</div>
        </div>

        <Icon className={cn('size-8', iconColor)} />
      </CardContent>
    </Card>
  )
}

export default StatisticCard
