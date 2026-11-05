import { resolve } from 'node:path';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

const root = resolve(process.cwd(), 'internal/system/assets');

export default defineConfig({
  root,
  publicDir: false,
  plugins: [tailwindcss()],
  server: {
    port: 5173,
    strictPort: true,
    origin: 'http://localhost:5173',
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        app: resolve(root, 'js/app.js'),
        tailwind: resolve(root, 'css/tailwind.css'),
      },
      output: {
        entryFileNames: 'js/[name].js',
        assetFileNames: 'css/[name][extname]',
      },
    },
  },
});
