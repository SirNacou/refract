import { env } from '@/env'
import type { CreateClientConfig } from '../gen/api/client.gen'

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: env.VITE_API_URL
})