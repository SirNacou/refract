import { env } from '@/env'
import ky from 'ky'
import { account } from './appwrite'

const api = ky.create({
  prefixUrl: env.VITE_API_URL,
  hooks: {
    beforeRequest: [
      async (request: any) => {
        const jwt = await account.createJWT()
        console.log("JWT:", jwt)

        if (jwt) {
          request.headers.set('Authorization', 'Bearer ' + jwt.jwt)
        }
      }
    ]
  }
})

export { api }
