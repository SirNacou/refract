import { env } from '@/env'
import ky from 'ky'
import { getZitadel } from './auth'

const api = ky.create({
  prefixUrl: env.VITE_API_URL,
  hooks: {
    beforeRequest: [
      (request) => {
        const auth = getZitadel()
        auth.userManager.getUser().then((user) => {
          if (user && !user.expired) {
            request.headers.set('Authorization', 'Bearer ' + user.access_token)
          }
        })
      }
    ]
  }
})

export { api }
