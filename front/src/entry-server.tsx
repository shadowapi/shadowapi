import { renderToString } from 'react-dom/server';
import { StaticRouter } from 'react-router';
import { createCache, extractStyle, StyleProvider } from '@ant-design/cssinjs';
import { ConfigProvider } from 'antd';
import App from './App';

export interface RenderResult {
  html: string;
  styles: string;
}

export async function render(
  url: string,
  ssrData?: Record<string, unknown>
): Promise<RenderResult> {
  const cache = createCache();

  const html = renderToString(
    <StyleProvider cache={cache}>
      <ConfigProvider>
        <StaticRouter location={url}>
          <App ssrData={ssrData} />
        </StaticRouter>
      </ConfigProvider>
    </StyleProvider>
  );

  const styles = extractStyle(cache);

  return { html, styles };
}
