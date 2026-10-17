import { useState } from 'react';
import { AdminAuthProvider, useAdminAuth } from './context/AdminAuthContext';
import { AdminLayout, type AdminTab } from './components/AdminLayout';
import { AdminLoginView } from './components/AdminLoginView';
import { AdminChangePasswordView } from './components/AdminChangePasswordView';
import { AccessDeniedView } from './components/AccessDeniedView';
import { AssetManagement } from './components/AssetManagement';
import { PriceManagement } from './components/PriceManagement';
import { MissingPricesReport } from './components/MissingPricesReport';
import { MissingFXRatesReport } from './components/MissingFXRatesReport';
import { FXManagement } from './components/FXManagement';
import { IngestionPipeline } from './components/IngestionPipeline';
import './App.css';

function AdminPortalContent() {
  const { user, isAuthenticated, isLoading } = useAdminAuth();
  const [currentTab, setCurrentTab] = useState<AdminTab>('assets');
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [missingPriceCount, setMissingPriceCount] = useState<number>(0);
  const [missingFXCount, setMissingFXCount] = useState<number>(0);

  const notify = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => {
      setToastMessage((current) => (current === msg ? null : current));
    }, 4000);
  };

  if (isLoading) {
    return (
      <div className="admin-loading-screen">
        <div className="admin-loading-spinner" />
        <p className="admin-loading-text">Authenticating GraphFolio Admin Session...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <AdminLoginView />;
  }

  if (user?.role !== 'ADMIN') {
    return <AccessDeniedView />;
  }

  if (user?.mustChangePassword) {
    return (
      <AdminChangePasswordView
        onSuccess={() => notify('Password updated successfully. Welcome to the Admin Portal!')}
      />
    );
  }

  return (
    <div className="admin-app-root">
      <AdminLayout
        currentTab={currentTab}
        onTabChange={setCurrentTab}
        missingPriceCount={missingPriceCount}
        missingFXCount={missingFXCount}
      >
        {currentTab === 'assets' && <AssetManagement onNotify={notify} />}
        {currentTab === 'prices' && <PriceManagement onNotify={notify} />}
        {currentTab === 'missing-prices' && (
          <MissingPricesReport onNotify={notify} onUpdateCount={setMissingPriceCount} />
        )}
        {currentTab === 'missing-fx' && (
          <MissingFXRatesReport onNotify={notify} onUpdateCount={setMissingFXCount} />
        )}
        {currentTab === 'fx' && <FXManagement onNotify={notify} />}
        {currentTab === 'ingestion' && <IngestionPipeline onNotify={notify} />}
      </AdminLayout>

      {toastMessage && (
        <div className="admin-toast">
          <span>ℹ️</span>
          <span>{toastMessage}</span>
        </div>
      )}
    </div>
  );
}

export default function App() {
  return (
    <AdminAuthProvider>
      <AdminPortalContent />
    </AdminAuthProvider>
  );
}
