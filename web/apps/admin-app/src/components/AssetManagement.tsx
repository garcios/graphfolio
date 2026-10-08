import React, { useState, useEffect, useCallback } from 'react';
import { client } from '@graphfolio/api-client';
import {
  Button,
  Badge,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableHeaderCell,
  TableCell,
  Input,
  Select,
} from '@graphfolio/ui';
import { AddInstrumentModal, type NewInstrumentData, type ExchangeOption } from './AddInstrumentModal';
import './AssetManagement.css';

interface InstrumentRecord {
  id: string;
  symbol: string;
  name: string;
  currencyCode: string;
  assetClass: string;
  exchangeCode: string;
  isin?: string | null;
  isActive: boolean;
}

interface AssetManagementProps {
  onNotify: (msg: string) => void;
}

export const AssetManagement: React.FC<AssetManagementProps> = ({ onNotify }) => {
  const [instruments, setInstruments] = useState<InstrumentRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [classFilter, setClassFilter] = useState('ALL');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [isModalOpen, setIsModalOpen] = useState(false);

  const [exchanges, setExchanges] = useState<ExchangeOption[]>([]);
  const [exchangesLoading, setExchangesLoading] = useState(false);
  const [exchangesError, setExchangesError] = useState<string | null>(null);

  useEffect(() => {
    if (!isModalOpen || exchanges.length > 0) return;
    setExchangesLoading(true);
    client
      .query({
        exchanges: {
          code: true,
          name: true,
        },
      })
      .then((res) => {
        if (res.exchanges) {
          setExchanges(res.exchanges);
        }
        setExchangesError(null);
      })
      .catch((err: any) => {
        console.error('Failed to query exchanges:', err);
        setExchangesError(err?.message || 'Failed to load exchanges');
      })
      .finally(() => {
        setExchangesLoading(false);
      });
  }, [isModalOpen, exchanges.length]);

  const fetchInstruments = useCallback(() => {
    setLoading(true);
    client
      .query({
        allInstruments: {
          id: true,
          symbol: true,
          name: true,
          currencyCode: true,
          assetClass: true,
          exchangeCode: true,
          isin: true,
          isActive: true,
        },
      })
      .then((res) => {
        if (res.allInstruments) {
          setInstruments(res.allInstruments);
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to query instruments:', err);
        onNotify(`Error loading instruments: ${err?.message || 'Server error'}`);
        setLoading(false);
      });
  }, [onNotify]);

  useEffect(() => {
    fetchInstruments();
  }, [fetchInstruments]);

  const handleToggleStatus = async (id: string, currentStatus: boolean) => {
    const target = instruments.find((i) => i.id === id);
    try {
      const res = await client.mutation({
        updateInstrument: {
          __args: {
            input: {
              id,
              isActive: !currentStatus,
            },
          },
          id: true,
          symbol: true,
          isActive: true,
        },
      });
      setInstruments((prev) =>
        prev.map((inst) =>
          inst.id === id ? { ...inst, isActive: res.updateInstrument.isActive } : inst
        )
      );
      onNotify(
        `Instrument ${target?.symbol || ''} ${res.updateInstrument.isActive ? 'activated' : 'deactivated'} successfully.`
      );
    } catch (err: any) {
      console.error('Failed to update instrument status:', err);
      onNotify(`Failed to update status for ${target?.symbol || ''}: ${err?.message || 'Server error'}`);
    }
  };

  const handleRegisterInstrument = async (data: NewInstrumentData) => {
    const res = await client.mutation({
      createInstrument: {
        __args: {
          input: {
            symbol: data.symbol,
            exchangeCode: data.exchangeCode,
            name: data.name,
            assetClass: data.assetClass,
            currencyCode: data.currencyCode,
            isin: data.isin,
          },
        },
        id: true,
        symbol: true,
        name: true,
        currencyCode: true,
        assetClass: true,
        exchangeCode: true,
        isin: true,
        isActive: true,
      },
    });

    if (res.createInstrument) {
      setInstruments((prev) => [res.createInstrument, ...prev]);
      onNotify(`Asset ${res.createInstrument.symbol} registered into master directory!`);
    }
  };

  const filteredInstruments = instruments.filter((inst) => {
    const matchesSearch =
      inst.symbol.toLowerCase().includes(searchQuery.toLowerCase()) ||
      inst.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (inst.isin && inst.isin.toLowerCase().includes(searchQuery.toLowerCase()));
    const matchesClass =
      classFilter === 'ALL' || inst.assetClass.toUpperCase() === classFilter.toUpperCase();
    const matchesStatus =
      statusFilter === 'ALL' ||
      (statusFilter === 'ACTIVE' && inst.isActive) ||
      (statusFilter === 'INACTIVE' && !inst.isActive);
    return matchesSearch && matchesClass && matchesStatus;
  });

  return (
    <div className="asset-mgmt">
      <div className="asset-mgmt__toolbar">
        <div className="asset-mgmt__filters">
          <Input
            className="asset-mgmt__search"
            placeholder="Search by symbol, name, or ISIN..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <Select
            className="asset-mgmt__class-filter"
            value={classFilter}
            onChange={(e) => setClassFilter(e.target.value)}
            options={[
              { label: 'All Asset Classes', value: 'ALL' },
              { label: 'Equities', value: 'EQUITY' },
              { label: 'ETFs', value: 'ETF' },
              { label: 'Mutual Funds', value: 'FUND' },
              { label: 'Fixed Income / Bonds', value: 'BOND' },
              { label: 'Cryptocurrencies', value: 'CRYPTO' },
              { label: 'Cash Equivalents', value: 'CASH_EQUIVALENT' },
            ]}
          />
          <Select
            className="asset-mgmt__status-filter"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            options={[
              { label: 'All Statuses', value: 'ALL' },
              { label: 'Active Only', value: 'ACTIVE' },
              { label: 'Inactive Only', value: 'INACTIVE' },
            ]}
          />
        </div>

        <Button variant="primary" onClick={() => setIsModalOpen(true)}>
          + Register New Asset
        </Button>
      </div>

      {loading ? (
        <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
          Loading master instruments...
        </div>
      ) : (
        <Table hoverable striped>
          <TableHead>
            <TableRow>
              <TableHeaderCell>Symbol</TableHeaderCell>
              <TableHeaderCell>Instrument Name</TableHeaderCell>
              <TableHeaderCell>Exchange</TableHeaderCell>
              <TableHeaderCell>Asset Class</TableHeaderCell>
              <TableHeaderCell>Base Currency</TableHeaderCell>
              <TableHeaderCell>Status</TableHeaderCell>
              <TableHeaderCell align="right">Actions</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredInstruments.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} align="center">
                  No instruments matched your criteria.
                </TableCell>
              </TableRow>
            ) : (
              filteredInstruments.map((inst) => (
                <TableRow key={inst.id}>
                  <TableCell>
                    <div className="asset-table-symbol">
                      <span>{inst.symbol}</span>
                      {inst.isin && (
                        <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                          {inst.isin}
                        </span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>{inst.name}</TableCell>
                  <TableCell>
                    <span style={{ fontFamily: 'var(--font-mono, monospace)', fontSize: '0.85rem' }}>
                      {inst.exchangeCode || '—'}
                    </span>
                  </TableCell>
                  <TableCell>
                    <Badge variant={inst.assetClass.toLowerCase()}>{inst.assetClass}</Badge>
                  </TableCell>
                  <TableCell>{inst.currencyCode}</TableCell>
                  <TableCell>
                    <Badge variant={inst.isActive ? 'active' : 'inactive'}>
                      {inst.isActive ? 'Active' : 'Inactive'}
                    </Badge>
                  </TableCell>
                  <TableCell align="right">
                    <div className="asset-actions" style={{ justifyContent: 'flex-end' }}>
                      <Button
                        size="sm"
                        variant={inst.isActive ? 'ghost' : 'secondary'}
                        onClick={() => handleToggleStatus(inst.id, inst.isActive)}
                      >
                        {inst.isActive ? 'Deactivate' : 'Activate'}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      )}

      <AddInstrumentModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleRegisterInstrument}
        exchanges={exchanges}
        exchangesLoading={exchangesLoading}
        exchangesError={exchangesError}
      />
    </div>
  );
};
