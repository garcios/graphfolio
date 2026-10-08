import { useState } from 'react';
import { AdminLayout, type AdminTab } from './components/AdminLayout';
import { AssetManagement } from './components/AssetManagement';
import { PriceManagement } from './components/PriceManagement';
import { FXManagement } from './components/FXManagement';
import { IngestionPipeline } from './components/IngestionPipeline';
import './App.css';

export default function App() {
  const [currentTab, setCurrentTab] = useState<AdminTab>('assets');
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const notify = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => {
      setToastMessage((current) => (current === msg ? null : current));
    }, 4000);
  };

  return (
    <div className="admin-app-root">
      <AdminLayout currentTab={currentTab} onTabChange={setCurrentTab}>
        {currentTab === 'assets' && <AssetManagement onNotify={notify} />}
        {currentTab === 'prices' && <PriceManagement onNotify={notify} />}
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
