import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router';
import { ConfigProvider } from 'antd';
import App from './App';
import theme from './theme';
import './index.css';

const container = document.getElementById('root')!;

// Always use createRoot for simplicity - avoids hydration mismatch errors
// SSR pages still render on server (good for SEO), client re-renders on load
container.innerHTML = '';
createRoot(container).render(
  <StrictMode>
    <ConfigProvider theme={theme}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </ConfigProvider>
  </StrictMode>
);
