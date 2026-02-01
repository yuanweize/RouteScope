import { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Settings from './pages/Settings';
import Setup from './pages/Setup';
import Targets from './pages/Targets';
import AppLayout from './components/Layout';
import ErrorBoundary from './components/ErrorBoundary';
import { checkNeedSetup } from './api';

function AppContent() {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    const check = async () => {
      try {
        const res = await checkNeedSetup() as any;
        if (res.need_setup && window.location.pathname !== '/setup') {
          navigate('/setup');
        }
      } catch (e) { console.error(e); }
      finally { setChecking(false); }
    };
    check();
  }, [navigate]);

  if (checking) return null;

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/setup" element={<Setup />} />
      <Route
        path="/*"
        element={
          <AppLayout>
            <Routes>
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="targets" element={<Targets />} />
              <Route path="settings" element={<Settings />} />
              <Route path="" element={<Navigate to="/dashboard" replace />} />
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </Routes>
          </AppLayout>
        }
      />
    </Routes>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <AppContent />
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
