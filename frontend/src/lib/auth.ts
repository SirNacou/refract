import { env } from '@/env'
import { createZitadelAuth, type ZitadelConfig } from '@zitadel/react'

const config: ZitadelConfig = {
  authority: env.VITE_ZITADEL_ISSUER,
  client_id: env.VITE_ZITADEL_CLIENT_ID,
  redirect_uri: env.VITE_ZITADEL_REDIRECT_URI,
  post_logout_redirect_uri: env.VITE_ZITADEL_POST_LOGOUT_URL,
  response_type: 'code',
  scope: env.VITE_ZITADEL_SCOPES,
}

let zitadelInstance: ReturnType<typeof createZitadelAuth> | null = null


export const getZitadel = () => {
  // Only create the instance on the client side
  if (typeof window === 'undefined') {
    throw new Error('Zitadel auth can only be used on the client side')
  }

  if (!zitadelInstance) {
    zitadelInstance = createZitadelAuth(config)
  }

  return zitadelInstance
}
