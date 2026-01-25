
import { env } from '@/env'
import { Account, Client, ID, Models } from 'appwrite'

export const client = new Client()

client.setEndpoint(env.VITE_APPWRITE_ENDPOINT)
  .setProject(env.VITE_APPWRITE_PROJECT)

export const account = new Account(client)
export { ID }
export type { Models }
