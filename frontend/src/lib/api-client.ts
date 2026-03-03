import { env } from "@/env"
import { createIsomorphicFn } from "@tanstack/react-start"
import type { CreateClientConfig } from "../gen/api/client.gen"

export const apiUrl = createIsomorphicFn()
  .client(() => window.location.origin + (env.VITE_API_URL || "/server/api"))
  .server(() => `api${env.VITE_API_URL || "/server/api"}`)

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: apiUrl(),
})
