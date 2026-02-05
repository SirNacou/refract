import {
	Timeline,
	TimelineConnector,
	TimelineContent,
	TimelineIcon,
	TimelineItem,
} from "@/components/time-line";
import dayjs from "@/lib/dayjs-config";
import { cn } from "@/lib/utils";
import LucideArrowRight from "~icons/lucide/arrow-right";
import LucideClock from "~icons/lucide/clock";
import LucideGlobe from "~icons/lucide/globe";
import LucideLinkedin from "~icons/lucide/linkedin";
import LucideMonitor from "~icons/lucide/monitor";
import LucideSearch from "~icons/lucide/search";
import LucideSmartphone from "~icons/lucide/smartphone";
import LucideTwitter from "~icons/lucide/twitter";

type Props = {
	events?: AnalyticsEvent[];
};

// --- MOCK DATA GENERATORS (For Simulation) ---

type TrafficSource = "google" | "twitter" | "linkedin" | "direct" | "other";
type DeviceType = "desktop" | "mobile";

export interface AnalyticsEvent {
	id: string;
	timestamp: Date;
	shortPath: string;
	destination: string;
	location: string;
	ip: string;
	source: TrafficSource;
	device: DeviceType;
}

const ClickActivityTimeline = ({ events }: Props) => {
	const getDeviceIcon = (device: DeviceType) => {
		return device === "mobile" ? (
			<LucideSmartphone className="h-4 w-4" />
		) : (
			<LucideMonitor className="h-4 w-4" />
		);
	};

	const getSourceIcon = (source: TrafficSource) => {
		switch (source) {
			case "google":
				return <LucideSearch className="h-3 w-3" />;
			case "twitter":
				return <LucideTwitter className="h-3 w-3" />;
			case "linkedin":
				return <LucideLinkedin className="h-3 w-3" />;
			case "direct":
				return <LucideArrowRight className="h-3 w-3" />;
			default:
				return <LucideGlobe className="h-3 w-3" />;
		}
	};

	return (
		<Timeline>
			{events?.map((event, index) => {
				const isLatest = index === 0;
				return (
					<TimelineItem key={event.id}>
						{index !== events.length - 1 && <TimelineConnector />}

						<TimelineIcon
							className={cn(
								"transition-all duration-500",
								isLatest
									? "bg-indigo-50 text-indigo-600 border-indigo-200 shadow-md scale-110"
									: "text-slate-500 bg-background",
							)}
						>
							{getDeviceIcon(event.device)}
						</TimelineIcon>

						<TimelineContent className="min-w-0">
							<div className="flex min-w-0 flex-col gap-1.5">
								{/* Header Row: Path + Time */}
								<div className="flex min-w-0 flex-wrap items-center justify-between gap-2">
									<span className="flex min-w-0 items-center gap-1.5 text-sm font-semibold text-foreground">
										<span className="truncate text-indigo-600">{event.shortPath}</span>
										<span className="text-muted-foreground text-[10px] font-normal px-1.5 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800">
											Redirected
										</span>
									</span>
									<span className="flex items-center gap-1 text-xs text-muted-foreground tabular-nums">
										<LucideClock className="h-3 w-3" />
										{dayjs(event.timestamp).fromNow()}
									</span>
								</div>

								{/* Details Row: Location & Source */}
								<div className="flex min-w-0 flex-wrap items-center gap-2 text-xs text-muted-foreground">
									<div className="flex min-w-0 items-center gap-1 rounded bg-secondary/50 px-2 py-1">
										<LucideGlobe className="h-3 w-3" />
										<span className="truncate">{event.location}</span>
									</div>

									<div className="flex items-center gap-1 rounded bg-secondary/50 px-2 py-1 capitalize">
										{getSourceIcon(event.source)}
										{event.source}
									</div>

									<div className="flex min-w-0 items-center gap-1 px-1 opacity-50">
										<span className="truncate">IP: {event.ip}</span>
									</div>
								</div>

								{/* Destination Hint */}
								<div className="max-w-full truncate text-[10px] text-muted-foreground/60">
									To: {event.destination}
								</div>
							</div>
						</TimelineContent>
					</TimelineItem>
				);
			})}
		</Timeline>
	);
};

export default ClickActivityTimeline;
