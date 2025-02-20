import { defineConfig } from 'orval';

export default defineConfig({
  api: {
    output: {
      target: 'src/api/api.ts',
      schemas: 'src/models',
      client: 'axios',
      mock: false,
      override: {
        mutator: {
          path: './src/api/axios.ts',
          name: 'customInstance',
        }
      }
    },
    input: {
      target: '../spec/openapi.yaml',
    },
  },
});
