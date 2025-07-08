import './index.css'
import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import App from './App'

const init = async () => {
  const basename = import.meta.env.VITE_APP_BASENAME || '/ui'
  ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
      <BrowserRouter basename={basename}>
        <QueryClientProvider client={new QueryClient()}>
          <App />
        </QueryClientProvider>
      </BrowserRouter>
    </React.StrictMode>
  )
}

init()
