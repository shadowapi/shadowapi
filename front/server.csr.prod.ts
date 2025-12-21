// Production CSR Server
// Serves static files for the client-side rendered SPA
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import express from 'express';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

async function createServer() {
  const app = express();

  // Serve static assets with long cache (hashed filenames)
  app.use(
    '/assets',
    express.static(path.resolve(__dirname, 'dist/assets'), {
      immutable: true,
      maxAge: '1y',
    })
  );

  // Serve other static files
  app.use(express.static(path.resolve(__dirname, 'dist'), { index: false }));

  // Health check endpoint
  app.get('/health', (_req, res) => {
    res.json({ status: 'ok', mode: 'csr-production' });
  });

  // Load the index.html template
  const indexHtml = fs.readFileSync(
    path.resolve(__dirname, 'dist/index.html'),
    'utf-8'
  );

  // SPA fallback - all routes return index.html
  app.use('*', (_req, res) => {
    res.status(200).set({ 'Content-Type': 'text/html' }).end(indexHtml);
  });

  const port = process.env.PORT || 5173;
  app.listen(port, () => {
    console.log(`Production CSR server running at http://localhost:${port}`);
  });
}

createServer();
