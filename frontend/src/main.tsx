import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { App as AntdApp, ConfigProvider } from 'antd'

import App from './App'
import './index.css'

const gisOwlTheme = {
  token: {
    colorPrimary: '#5B6F2E',
    colorLink: '#5B6F2E',
    colorLinkHover: '#7A943E',
    colorSuccess: '#389e0d',
    colorError: '#cf1322',
    colorWarning: '#d48806',
    colorTextBase: '#434343',
    colorBgBase: '#ffffff',
    borderRadius: 6,
  },
  components: {
    Button: {
      colorPrimary: '#434343',
      colorPrimaryHover: '#595959',
      colorPrimaryActive: '#262626',
    },
    Badge: {
      colorBgContainer: '#5B6F2E',
      colorText: '#ffffff',
    },
    Menu: {
      itemSelectedBg: '#f5f5f5',
      itemSelectedColor: '#262626',
      itemHoverBg: '#fafafa',
      itemHoverColor: '#434343',
      itemActiveBg: '#f0f0f0',
    },
    Table: {
      headerBg: '#FDF8F0',
      headerColor: '#3D4A1A',
      rowHoverBg: '#FAFAF5',
    },
    Steps: {
      colorPrimary: '#52c41a',
      finishIconBorderColor: '#52c41a',
      colorText: '#434343',
    },
  },
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <ConfigProvider theme={gisOwlTheme}>
        <AntdApp>
          <App />
        </AntdApp>
      </ConfigProvider>
    </BrowserRouter>
  </StrictMode>
)
