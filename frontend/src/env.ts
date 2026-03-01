import { createEnv } from '@t3-oss/env-core'
import { z } from 'zod'

export const env = createEnv({
  server: {
    SERVER_URL: z.url().optional(),
    DATABASE_URL: z.url(),
    BETTER_AUTH_URL: z.url(),
    BETTER_AUTH_SECRET: z.string(),
    GITHUB_CLIENT_ID: z.string(),
    GITHUB_CLIENT_SECRET: z.string(),
  },

  clientPrefix: 'VITE_',

  client: {
    VITE_APP_TITLE: z.string().min(1).optional(),
    VITE_API_URL: z.url().optional(),
    VITE_DEFAULT_REDIRECT_URL: z.url().optional(),
    VITE_ENVIRONMENT: z.enum(['development', 'production', 'test'] as const).default('development'),
  },

  runtimeEnv: {
    VITE_API_URL: (typeof window !== "undefined" && window._env_?.API_URL) ? window._env_.API_URL : process.env.VITE_API_URL,
    VITE_DEFAULT_REDIRECT_URL: (typeof window !== "undefined" && window._env_?.DEFAULT_REDIRECT_URL) ? window._env_.DEFAULT_REDIRECT_URL : process.env.VITE_DEFAULT_REDIRECT_URL,
    VITE_ENVIRONMENT: (typeof window !== "undefined" && window._env_?.ENVIRONMENT) ? window._env_.ENVIRONMENT : process.env.VITE_ENVIRONMENT || 'development',
    ...process.env,
    ...import.meta.env,
  },

  emptyStringAsUndefined: true,
})
