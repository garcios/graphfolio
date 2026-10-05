import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
  },
  resolve: {
    alias: {
      '@graphfolio/ui': path.resolve(__dirname, '../../packages/ui/src'),
      '@graphfolio/api-client': path.resolve(__dirname, '../../packages/api-client/src'),
    },
  },
});
