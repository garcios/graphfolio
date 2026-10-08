import React from 'react';
import { Modal, Button } from '@graphfolio/ui';
import './DeleteAssetModal.css';

export interface AssetToDelete {
  id: string;
  symbol: string;
  name: string;
  currencyCode: string;
  exchangeCode: string;
  assetClass?: string;
  isin?: string | null;
}

interface DeleteAssetModalProps {
  isOpen: boolean;
  onClose: () => void;
  asset: AssetToDelete | null;
  onConfirm: () => Promise<void>;
  isDeleting: boolean;
  error?: string | null;
}

export const DeleteAssetModal: React.FC<DeleteAssetModalProps> = ({
  isOpen,
  onClose,
  asset,
  onConfirm,
  isDeleting,
  error,
}) => {
  if (!asset) return null;

  return (
    <Modal
      isOpen={isOpen}
      onClose={isDeleting ? () => {} : onClose}
      title={`Delete Asset: ${asset.symbol}`}
      size="md"
    >
      <div className="delete-asset-modal">
        <div className="delete-asset__warning">
          <strong>Warning:</strong> Deleting this asset is irreversible. It will permanently remove{' '}
          <strong>{asset.symbol}</strong> from the master directory and atomically delete all of its
          associated historical closing market prices.
        </div>

        <div className="delete-asset__details">
          <div className="delete-asset__row">
            <span className="delete-asset__label">Symbol:</span>
            <span className="delete-asset__value delete-asset__value--mono">{asset.symbol}</span>
          </div>
          <div className="delete-asset__row">
            <span className="delete-asset__label">Name:</span>
            <span className="delete-asset__value">{asset.name}</span>
          </div>
          <div className="delete-asset__row">
            <span className="delete-asset__label">Exchange:</span>
            <span className="delete-asset__value delete-asset__value--mono">
              {asset.exchangeCode || '—'}
            </span>
          </div>
          <div className="delete-asset__row">
            <span className="delete-asset__label">Base Currency:</span>
            <span className="delete-asset__value">{asset.currencyCode}</span>
          </div>
          {asset.isin && (
            <div className="delete-asset__row">
              <span className="delete-asset__label">ISIN:</span>
              <span className="delete-asset__value delete-asset__value--mono">{asset.isin}</span>
            </div>
          )}
        </div>

        {error && <div className="delete-asset__error">{error}</div>}

        <div className="delete-asset__actions">
          <Button
            type="button"
            variant="secondary"
            disabled={isDeleting}
            onClick={onClose}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="danger"
            isLoading={isDeleting}
            onClick={onConfirm}
          >
            Delete Asset & Prices
          </Button>
        </div>
      </div>
    </Modal>
  );
};
