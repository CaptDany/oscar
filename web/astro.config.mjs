import { defineConfig } from 'astro/config';
import preact from '@astrojs/preact';
import node from '@astrojs/node';
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  output: 'static',
  adapter: node({
    mode: 'standalone'
  }),
  integrations: [
    preact({ compat: true }),
  ],
  compressHTML: true,
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
    build: {
      cssMinify: true,
      minify: true,
    },
  },
});