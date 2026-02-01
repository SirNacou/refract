import {
	Sidebar,
	SidebarContent,
	SidebarFooter,
	SidebarGroup,
	SidebarGroupContent,
	SidebarGroupLabel,
	SidebarHeader,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
	useSidebar,
} from "@/components/ui/sidebar"
import { cn } from "@/lib/utils"
import { Link, type LinkOptions, useLocation } from "@tanstack/react-router"
import LucideBarChart from "~icons/lucide/bar-chart-3"
import LucideKey from "~icons/lucide/key"
import LucideLink from "~icons/lucide/link"
import LucideLink2 from "~icons/lucide/link-2"
import LucideLogOut from "~icons/lucide/log-out"
import LucideMousePointer2 from "~icons/lucide/mouse-pointer-2"
import LucideSidebar from "~icons/lucide/sidebar"
import LucideTrendingUp from "~icons/lucide/trending-up"
import LucideUser from "~icons/lucide/user"
import { Button } from "./ui/button"

type NavOptions = {
	title: string
	icon: React.ComponentType<React.SVGProps<SVGSVGElement>>
} & LinkOptions

type LinkGroup = {
	title: string
	items: NavOptions[]
}

const navGroups: LinkGroup[] = [
	{
		title: "Main",
		items: [
			{
				title: "Dashboard",
				to: "/",
				icon: LucideBarChart,
			},
			{
				title: "My URLs",
				to: "/urls",
				icon: LucideLink2,
			},
		],
	},
	{
		title: "Analytics",
		items: [
			{
				title: "Click Statistics",
				to: "/",
				icon: LucideMousePointer2,
			},
			{
				title: "Top Performing",
				to: "/",
				icon: LucideTrendingUp,
			},
		],
	},
	{
		title: "Settings",
		items: [
			{
				title: "Account Settings",
				to: "/account/$accountView" as const,
				params: { accountView: "settings" },
				icon: LucideUser,
			},
			{
				title: "API Keys",
				to: "/account/$accountView" as const,
				params: { accountView: "api" },
				icon: LucideKey,
			},
		],
	},
]

const user = {
	name: "John Doe",
	email: "john@example.com",
	avatar: "JD",
}

export function AppSidebar() {
	const location = useLocation()

	return (
		<Sidebar>
			<SidebarHeader>
				<div className="flex items-center gap-2 px-2 py-4">
					<div className="flex items-center justify-center size-8 rounded-lg bg-primary">
						<LucideLink className="size-5 text-primary-foreground" />
					</div>
					<span className="font-semibold text-lg">Refract</span>
				</div>
			</SidebarHeader>
			<SidebarContent>
				{navGroups.map((group) => (
					<SidebarGroup key={group.title}>
						<SidebarGroupLabel>{group.title}</SidebarGroupLabel>
						<SidebarGroupContent>
							<SidebarMenu>
								{group.items.map((item) => {
									const isActive = location.pathname === item.to
									return (
										<SidebarMenuItem key={item.title}>
											<SidebarMenuButton
												asChild
												isActive={isActive}
												className={cn(
													"transition-all duration-200",
													isActive && "bg-primary/10 text-primary font-medium",
												)}
											>
												<Link
													to={item.to}
													params={item.params}
													className="flex items-center gap-3"
												>
													<item.icon className="size-4" />
													<span>{item.title}</span>
												</Link>
											</SidebarMenuButton>
										</SidebarMenuItem>
									)
								})}
							</SidebarMenu>
						</SidebarGroupContent>
					</SidebarGroup>
				))}
			</SidebarContent>
			<SidebarFooter>
				<div className="flex items-center gap-3 px-2 py-3 rounded-lg hover:bg-muted/50 transition-colors cursor-pointer">
					<div className="flex items-center justify-center size-9 rounded-full bg-muted text-sm font-medium">
						{user.avatar}
					</div>
					<div className="flex-1 min-w-0">
						<p className="text-sm font-medium truncate">{user.name}</p>
						<p className="text-xs text-muted-foreground truncate">
							{user.email}
						</p>
					</div>
					<Button
						variant="ghost"
						size="icon"
						className="size-8 text-muted-foreground hover:text-foreground"
					>
						<LucideLogOut className="size-4" />
					</Button>
				</div>
			</SidebarFooter>
		</Sidebar>
	)
}

export function CustomSidebarTrigger({
	className,
	onClick,
	...props
}: React.ComponentProps<typeof Button>) {
	const { toggleSidebar } = useSidebar()

	return (
		<Button
			data-sidebar="trigger"
			data-slot="sidebar-trigger"
			variant="ghost"
			size="icon"
			className={cn("size-10 md:size-12", className)}
			onClick={(event) => {
				onClick?.(event)
				toggleSidebar()
			}}
			{...props}
		>
			<LucideSidebar className="w-5 h-5 md:w-6 md:h-6" />
			<span className="sr-only">Toggle Sidebar</span>
		</Button>
	)
}
