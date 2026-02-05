import { AppSidebar, CustomSidebarTrigger } from "@/components/app-sidebar";
import { SidebarProvider } from "@/components/ui/sidebar";
import { authMiddleware } from "@/middleware/auth";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_protected")({
	component: RouteComponent,
	server: {
		middleware: [authMiddleware],
	},
});

function RouteComponent() {
	return (
		<SidebarProvider>
			<AppSidebar />
			<main className="w-full flex flex-col">
				<CustomSidebarTrigger />
				<div className="px-6 py-3">
					<Outlet />
				</div>
			</main>
		</SidebarProvider>
	);
}
