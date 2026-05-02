import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

// BACKEND_URL is set when running inside Docker so the proxy reaches the
// backend container by name instead of localhost.
const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    port: 5173,
    proxy: {
      '/api':          { target: backendUrl, changeOrigin: true },
      '/auth':         { target: backendUrl, changeOrigin: true },
      '/health':       { target: backendUrl, changeOrigin: true },
      '/openapi.yaml': { target: backendUrl, changeOrigin: true },
      '/docs':         { target: backendUrl, changeOrigin: true }
    }
  }
});
