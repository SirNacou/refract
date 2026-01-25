import { api } from '@/lib/api-client'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import ky from 'ky'

// Define types here (or import from a shared types file)
export type ShortURL = {
  id: string
  original_url: string
  short_code: string
  title: string
  click_count: number
  is_active: boolean
  created_at: string
}

type CreateURLPayload = {
  destination: string,
  shortCode: string,
  title: string,
  notes?: string,
  expiration?: string,
}

const createURL = async (data: CreateURLPayload): Promise<{ short_code: string } | Error> => {
  const response = await api.post('/api/v1/urls', {
    body: JSON.stringify(data)
  }).json<{ short_code: string }>()

  return response
}

const fetchUrls = async (): Promise<ShortURL[]> => {
  const res = await api.get<ShortURL[]>('/api/v1/urls')
  if (res.status === 200) {
    return res.json()
  }
  // Return empty array [] to test the Empty State
  return [
    {
      id: '1',
      original_url: 'https://google.com',
      short_code: 'gg',
      title: "Google",
      click_count: 10,
      is_active: true,
      created_at: new Date().toISOString()
    }
  ]
}

export function useUrls() {
  const queryClient = useQueryClient()

  // 1. Query (Read)
  const query = useQuery({
    queryKey: ['urls'],
    queryFn: fetchUrls,
  })

  // 2. Mutation (Write) - Integrating the Create logic here
  const createUrlMutation = useMutation({
    mutationFn: async (params: CreateURLPayload) => {
      // Mock API call
      console.log("Saving...", params)
      return createURL(params)
    },
    // BEST PRACTICE: Automatic Cache Invalidation
    // When a new URL is created, tell React Query the 'urls' list is stale.
    // It will automatically refetch the list in the background.
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['urls'] })
    },
  })

  return {
    urls: query.data || [],
    isLoading: query.isLoading,
    isError: query.isError,
    createUrl: createUrlMutation.mutate,
    isCreating: createUrlMutation.isPending
  }
}