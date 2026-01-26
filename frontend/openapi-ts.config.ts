import { defineConfig } from '@hey-api/openapi-ts'

export default defineConfig({
  input: {
    path: 'http://localhost:8080/openapi.yaml',
    watch: true,
  },
  output: {
    path: 'src/gen/api',
    source: true,
    postProcess: ['biome:format']
  },
  plugins: [
    "@hey-api/typescript",
    {
      name: "@hey-api/sdk",
      auth: true,
      transformer: true,
      validator: 'zod'
    },
    {
      name: "@hey-api/transformers",
      dates: true
    },
    {
      name: "@hey-api/schemas",
      type: 'json'
    },
    {
      name: "@hey-api/client-ky",
      runtimeConfigPath: '@/lib/api-client.ts'
    },
    {
      exportFromIndex: true,
      name: 'zod'
    },
    {
      name: "@tanstack/react-query",
    }
  ]
})