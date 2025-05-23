import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import {resolve} from 'node:path';
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react({babel: {plugins: ["relay"]}}),
    tailwindcss(),
  ],
  server: {
    cors: {
      // the origin you will be accessing via browser
      origin: 'http://localhost:8080',
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
})
