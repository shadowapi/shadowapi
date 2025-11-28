import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import express from 'express';
import { createServer as createViteServer } from 'vite';
import { fetchDataForRoute } from './src/lib/data-fetching.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

async function createServer() {
  const app = express();

  // Create Vite server in middleware mode
  const vite = await createViteServer({
    server: { middlewareMode: true },
    appType: 'custom',
  });

  // Use Vite's connect instance as middleware
  app.use(vite.middlewares);

  // Handle /page/* routes with SSR
  app.use('/page*', async (req, res) => {
    const url = req.originalUrl;

    try {
      // Read the index.html template
      let template = fs.readFileSync(
        path.resolve(__dirname, 'index.html'),
        'utf-8'
      );

      // Apply Vite HTML transforms (injects HMR client, etc.)
      template = await vite.transformIndexHtml(url, template);

      // Load the server entry module
      const { render } = await vite.ssrLoadModule('/src/entry-server.tsx');

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
      // Fix stack trace for Vite
      vite.ssrFixStacktrace(e as Error);
      console.error(e);
      res.status(500).end((e as Error).message);
    }
  });

  // Health check endpoint
  app.get('/health', (_req, res) => {
    res.json({ status: 'ok', mode: 'ssr' });
  });

  const port = process.env.PORT || 3000;
  app.listen(port, () => {
    console.log(`SSR server running at http://localhost:${port}`);
    console.log('Handling /page/* routes with server-side rendering');
  });
}

createServer();
