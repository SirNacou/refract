import { Skeleton } from '@/components/ui/skeleton'
import URLFormDialog from '@/features/urls/url-form-dialog'
import URLList from '@/features/urls/url-list'
import { useUrls } from '@/features/urls/use-urls'
import { createFileRoute } from '@tanstack/react-router'
import { Ghost, LayoutDashboard } from 'lucide-react'

export const Route = createFileRoute('/_authenticated/(urls)/dashboard')({
    component: RouteComponent,
})

function RouteComponent() {
    const { urls, isLoading, isError } = useUrls()

    return (
        <div className="min-h-screen bg-gray-50/50 p-8">
            <div className="mx-auto max-w-5xl space-y-8">

                {/* Page Header */}
                <div className="flex items-center space-x-2">
                    <div className="p-2 bg-primary/10 rounded-lg">
                        <LayoutDashboard className="h-6 w-6 text-primary" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
                        <p className="text-muted-foreground">Manage your links and view analytics.</p>
                    </div>
                </div>

                <main className="space-y-4">
                    <div className='flex flex-row justify-between'>
                        <h2 className="text-lg font-semibold">Your Recent Links</h2>

                        <URLFormDialog />
                    </div>

                    {isLoading ? (
                        <DashboardSkeleton />
                    ) : isError ? (
                        <div className="p-4 border border-red-200 bg-red-50 text-red-600 rounded-md">
                            Error loading data. Please try again.
                        </div>
                    ) : urls.length === 0 ? (
                        <EmptyState />
                    ) : (
                        <URLList data={urls} />
                    )}
                </main>
            </div>
        </div>
    )
}

// Sub-component: Empty State
function EmptyState() {
    return (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed p-8 text-center animate-in fade-in-50">
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
                <Ghost className="h-10 w-10 text-muted-foreground" />
            </div>
            <h3 className="mt-4 text-lg font-semibold">No links created yet</h3>
            <p className="mb-4 mt-2 text-sm text-muted-foreground max-w-sm">
                It looks like you haven't shortened any links. Use the form on the left to create your first tracking link.
            </p>
        </div>
    )
}

// Sub-component: Skeleton
function DashboardSkeleton() {
    return (
        <div className="space-y-4">
            <Skeleton className="h-30 w-full rounded-lg" />
            <Skeleton className="h-30 w-full rounded-lg" />
            <Skeleton className="h-30 w-full rounded-lg" />
        </div>
    )
}