import { cn } from "@/lib/utils";
import React from "react";

const Timeline = React.forwardRef<
	HTMLOListElement,
	React.HTMLAttributes<HTMLOListElement>
>(({ className, ...props }, ref) => (
	<ol
		ref={ref}
		className={cn("flex flex-col space-y-0", className)}
		{...props}
	/>
));
Timeline.displayName = "Timeline";

const TimelineItem = React.forwardRef<
	HTMLLIElement,
	React.LiHTMLAttributes<HTMLLIElement>
>(({ className, ...props }, ref) => (
	<li
		ref={ref}
		className={cn("relative flex gap-4 pb-8 last:pb-0", className)}
		{...props}
	/>
));
TimelineItem.displayName = "TimelineItem";

const TimelineConnector = React.forwardRef<
	HTMLDivElement,
	React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
	<div
		ref={ref}
		className={cn(
			"absolute left-4.75 top-10 bottom-0 w-px bg-border",
			"last-of-type:hidden",
			className,
		)}
		{...props}
	/>
));
TimelineConnector.displayName = "TimelineConnector";

const TimelineIcon = React.forwardRef<
	HTMLDivElement,
	React.HTMLAttributes<HTMLDivElement>
>(({ className, children, ...props }, ref) => (
	<div
		ref={ref}
		className={cn(
			"relative z-10 flex h-10 w-10 shrink-0 items-center justify-center rounded-full border bg-background shadow-sm",
			className,
		)}
		{...props}
	>
		{children}
	</div>
));
TimelineIcon.displayName = "TimelineIcon";

const TimelineContent = React.forwardRef<
	HTMLDivElement,
	React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
	<div
		ref={ref}
		className={cn("flex flex-1 flex-col justify-start pt-2", className)}
		{...props}
	/>
));
TimelineContent.displayName = "TimelineContent";

export {
  Timeline,
  TimelineItem,
  TimelineConnector,
  TimelineIcon,
  TimelineContent, 
};