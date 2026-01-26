import { cn } from "@/lib/utils"
import { Route as IndexRoute } from '@/routes/_protected/index'
import { AuthView } from "@daveyplate/better-auth-ui"
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/auth/$authView")({
  component: RouteComponent
})

function RouteComponent() {
  const { authView } = Route.useParams()
  const redirectTo = IndexRoute.fullPath

  return (
    <main className="container mx-auto flex grow flex-col items-center justify-center gap-3 self-center p-4 md:p-6 h-screen">
      <AuthView pathname={authView} socialLayout="auto" redirectTo={redirectTo} />

      <p className={cn(["callback", "sign-out"].includes(authView) && "hidden", "text-muted-foreground text-xs")}>
        Powered by{" "}
        <a className="text-warning underline" href="https://better-auth.com" target="_blank" rel="noreferrer">
          better-auth.
        </a>
      </p>
    </main>
  )
}