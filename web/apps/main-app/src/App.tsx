import { AuthProvider, useAuth } from './context/AuthContext';
import { AuthView } from './components/AuthView';
import { Dashboard } from './components/Dashboard';
import './App.css';

function AppContent() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="auth-loading-screen" role="status" aria-live="polite">
        <div className="auth-loading-spinner" />
        <p>Securing investor session...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <AuthView />;
  }

  return (
    <div className="app-container">
      <Dashboard />
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}

export default App;
