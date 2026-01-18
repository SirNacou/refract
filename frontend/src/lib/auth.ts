import { env } from '@/env'
import { createClientOnlyFn } from '@tanstack/react-start'
import { createZitadelAuth, type ZitadelAuth, type ZitadelConfig } from '@zitadel/react'

const config: ZitadelConfig = {
  authority: env.VITE_ZITADEL_ISSUER,
  client_id: env.VITE_ZITADEL_CLIENT_ID,
  redirect_uri: env.VITE_ZITADEL_REDIRECT_URI,
  post_logout_redirect_uri: env.VITE_ZITADEL_POST_LOGOUT_URL,
  response_type: 'code',
  scope: env.VITE_ZITADEL_SCOPES,
}

let zitadel: ZitadelAuth | null = null

export const getZitadel = createClientOnlyFn(() => {
  if (!zitadel) {
    zitadel = createZitadelAuth(config)
  }
  return zitadel
})