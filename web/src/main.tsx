import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import '@arco-design/web-react/dist/css/arco.css';

console.log('Frontend Version: v1.1.0-DarkFix');

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
