import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@radix-ui/react-dropdown-menu"
import { formatDistanceToNow } from 'date-fns'
import { BarChart3, CalendarDays, Copy, ExternalLink, MoreHorizontal } from "lucide-react"
import { toast } from "sonner"
import { Badge } from '../../components/ui/badge'
import { Button } from "../../components/ui/button"
import { Skeleton } from "../../components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../../components/ui/table"
import { type ShortURL } from "./use-urls"
type Props = {
  data: ShortURL[]
}

const URLList = ({ data }: Props) => {
  const copyToClipboard = (shortCode: string) => {
    const fullUrl = `${window.location.origin}/${shortCode}`
    navigator.clipboard.writeText(fullUrl)
    toast.message("Copied to clipboard")
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-62.5">Short Link</TableHead>
            <TableHead className="hidden md:table-cell">Destination</TableHead>
            <TableHead>Clicks</TableHead>
            <TableHead className="hidden md:table-cell">Status</TableHead>
            <TableHead className="hidden md:table-cell">Created</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data?.map((url) => (
            <TableRow key={url.id}>
              {/* Short Link Column */}
              <TableCell className="font-medium">
                <div className="flex flex-col">
                  <span className="text-sm font-semibold">{url.title || url.short_code}</span>
                  <span className="text-xs text-muted-foreground flex items-center gap-1">
                    /{url.short_code}
                  </span>
                </div>
              </TableCell>

              {/* Destination Column */}
              <TableCell className="hidden md:table-cell max-w-50">
                <div className="truncate text-muted-foreground text-sm" title={url.original_url}>
                  {url.original_url}
                </div>
              </TableCell>

              {/* Clicks Column */}
              <TableCell>
                <div className="flex items-center gap-2">
                  <BarChart3 className="h-4 w-4 text-muted-foreground" />
                  <span className="font-mono">{url.click_count.toLocaleString()}</span>
                </div>
              </TableCell>

              {/* Status Column */}
              <TableCell className="hidden md:table-cell">
                <Badge variant={url.is_active ? "secondary" : "destructive"}>
                  {url.is_active ? "Active" : "Archived"}
                </Badge>
              </TableCell>

              {/* Created At Column */}
              <TableCell className="hidden md:table-cell text-muted-foreground text-sm">
                <div className="flex items-center gap-2">
                  <CalendarDays className="h-3 w-3" />
                  {formatDistanceToNow(new Date(url.created_at), { addSuffix: true })}
                </div>
              </TableCell>

              {/* Actions Column */}
              <TableCell className="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" className="h-8 w-8 p-0">
                      <span className="sr-only">Open menu</span>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                    <DropdownMenuItem onClick={() => copyToClipboard(url.short_code)}>
                      <Copy className="mr-2 h-4 w-4" />
                      Copy Link
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => window.open(url.original_url, '_blank')}>
                      <ExternalLink className="mr-2 h-4 w-4" />
                      Visit Destination
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem className="text-red-600">
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

export function URLListSkeleton() {
  return (
    <div className="rounded-md border">
      <div className="p-4 space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex items-center justify-between space-x-4">
            <div className="space-y-2">
              <Skeleton className="h-4 w-50" />
              <Skeleton className="h-4 w-37.5" />
            </div>
            <Skeleton className="h-8 w-8 rounded-full" />
          </div>
        ))}
      </div>
    </div>
  )
}

export default URLList