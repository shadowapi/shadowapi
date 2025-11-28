import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'

import { Routes, Route, BrowserRouter } from 'react-router'

import AppLayout from './layouts/AppLayout.tsx'
import App from './App.tsx'

// Note: /page/* routes are handled by the SSR container
// Use regular <a> tags (not <Link>) when linking to /page/* routes
// to ensure full page navigation to the SSR server

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<AppLayout><App /></AppLayout>} />
        <Route path="/app/*" element={<AppLayout><App /></AppLayout>} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
