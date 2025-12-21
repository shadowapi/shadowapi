// Production SSR Server
// Serves static files and handles SSR for HTML requests
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import express from 'express';
import { fetchDataForRoute } from './src/lib/data-fetching.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

async function createServer() {
  const app = express();

  // Serve static assets from the client build
  app.use(
    '/assets',
    express.static(path.resolve(__dirname, 'dist/client/assets'), {
      immutable: true,
      maxAge: '1y',
    })
  );

  // Serve other static files from client build
  app.use(express.static(path.resolve(__dirname, 'dist/client'), { index: false }));

  // Health check endpoint
  app.get('/health', (_req, res) => {
    res.json({ status: 'ok', mode: 'ssr-production' });
  });

  // Load the production template and server render module
  const template = fs.readFileSync(
    path.resolve(__dirname, 'dist/client/index.html'),
    'utf-8'
  );

  // Import the built SSR module
  const { render } = await import('./dist/server/entry-server.js');

  // Handle all routes with SSR
  app.use('*', async (req, res) => {
    const url = req.originalUrl;

    try {
      // Fetch data for this route
      const ssrData = await fetchDataForRoute(url);

      // Render the app HTML
      const { html: appHtml, styles } = await render(url, ssrData);

      // Inject rendered content into template
      const html = template
        .replace('<!--ssr-styles-->', styles)
        .replace('<!--ssr-outlet-->', appHtml)
        .replace(
          '<!--ssr-data-->',
          `<script>window.__SSR_DATA__=${JSON.stringify(ssrData)}</script>`
        );

      // Send the rendered HTML
      res.status(200).set({ 'Content-Type': 'text/html' }).end(html);
    } catch (e) {
      console.error(e);
      res.status(500).end((e as Error).message);
    }
  });

  const port = process.env.PORT || 3000;
  app.listen(port, () => {
    console.log(`Production SSR server running at http://localhost:${port}`);
  });
}

createServer();
