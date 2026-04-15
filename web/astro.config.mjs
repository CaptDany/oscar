import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import node from '@astrojs/node';
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  output: 'server',
  adapter: node({
    mode: 'standalone'
  }),
  integrations: [
    react(),
  ],
  server: {
    port: 4321,
    host: true,
  },
  vite: {
    plugins: [tailwindcss()],
    server: {
      proxy: {
        '/api/v1': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  },
});
