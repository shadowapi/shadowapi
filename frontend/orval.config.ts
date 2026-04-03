import { defineConfig } from 'orval'

export default defineConfig({
  shadowapi: {
    input: {
      target: '../spec/openapi-bundled.yaml',
    },
    output: {
      mode: 'split',
      target: 'src/api/generated/index.ts',
      client: 'swr',
      httpClient: 'axios',
      baseUrl: '',
      override: {
        mutator: {
          path: 'src/api/mutator.ts',
          name: 'apiInstance',
        },
        query: {
          useSuspense: false,
          useQuery: true,
          useMutation: true,
          signal: true,
        },
      },
    },
    hooks: {
      afterAllFilesWrite: 'prettier --write src/api/generated/**/*.ts',
    },
  },
})
