import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'

import { Routes, Route, BrowserRouter } from 'react-router'

import AppLayout from './layouts/AppLayout.tsx'
import PageLayout from './layouts/PageLayout.tsx'
import App from './App.tsx'
import Pages from './Pages.tsx'
import DocumentationIndex from './pages/documentation/index.tsx'
import DatasourceIndex from './pages/documentation/datasource/index.tsx'
import GmailDocumentation from './pages/documentation/datasource/gmail.tsx'
import TelegramDocumentation from './pages/documentation/datasource/telegram.tsx'
import AboutPage from './pages/about/index.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<AppLayout><App /></AppLayout>} />
        <Route path="/page/" element={<PageLayout><Pages /></PageLayout>} />
        <Route path="/page/documentation" element={<PageLayout><DocumentationIndex /></PageLayout>} />
        <Route path="/page/documentation/datasource" element={<PageLayout><DatasourceIndex /></PageLayout>} />
        <Route path="/page/documentation/datasource/gmail" element={<PageLayout><GmailDocumentation /></PageLayout>} />
        <Route path="/page/documentation/datasource/telegram" element={<PageLayout><TelegramDocumentation /></PageLayout>} />
        <Route path="/page/about" element={<PageLayout><AboutPage /></PageLayout>} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
