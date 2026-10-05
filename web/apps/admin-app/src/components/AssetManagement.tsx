import React, { useState, useEffect } from 'react';
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
import { AddInstrumentModal, type NewInstrumentData } from './AddInstrumentModal';
import './AssetManagement.css';

interface InstrumentRecord {
  id: string;
  symbol: string;
  name: string;
  currencyCode: string;
  assetClass: string;
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
  const [isModalOpen, setIsModalOpen] = useState(false);

  const fetchInstruments = () => {
    setLoading(true);
    client
      .query({
        instruments: {
          id: true,
          symbol: true,
          name: true,
          currencyCode: true,
          assetClass: true,
        },
      })
      .then((res) => {
        if (res.instruments) {
          // Normalize and enrich with active status
          const enriched: InstrumentRecord[] = res.instruments.map((item) => ({
            ...item,
            isActive: true, // Default active
          }));
          setInstruments(enriched);
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to query instruments:', err);
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchInstruments();
  }, []);

  const handleToggleStatus = (id: string, currentStatus: boolean) => {
    setInstruments((prev) =>
      prev.map((inst) =>
        inst.id === id ? { ...inst, isActive: !currentStatus } : inst
      )
    );
    const target = instruments.find((i) => i.id === id);
    onNotify(
      `Instrument ${target?.symbol || ''} ${currentStatus ? 'deactivated' : 'activated'} successfully.`
    );
  };

  const handleRegisterInstrument = async (data: NewInstrumentData) => {
    // Optimistically update list and simulate server registration
    const newRecord: InstrumentRecord = {
      id: `inst-${Date.now()}`,
      symbol: data.symbol,
      name: data.name,
      currencyCode: data.currencyCode,
      assetClass: data.assetClass,
      isActive: true,
    };
    setInstruments((prev) => [newRecord, ...prev]);
    onNotify(`Asset ${data.symbol} registered into master directory!`);
  };

  const filteredInstruments = instruments.filter((inst) => {
    const matchesSearch =
      inst.symbol.toLowerCase().includes(searchQuery.toLowerCase()) ||
      inst.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesClass =
      classFilter === 'ALL' || inst.assetClass.toUpperCase() === classFilter.toUpperCase();
    return matchesSearch && matchesClass;
  });

  return (
    <div className="asset-mgmt">
      <div className="asset-mgmt__toolbar">
        <div className="asset-mgmt__filters">
          <Input
            className="asset-mgmt__search"
            placeholder="Search by symbol or name..."
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
              { label: 'Cryptocurrencies', value: 'CRYPTO' },
              { label: 'Commodities', value: 'COMMODITY' },
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
              <TableHeaderCell>Asset Class</TableHeaderCell>
              <TableHeaderCell>Base Currency</TableHeaderCell>
              <TableHeaderCell>Status</TableHeaderCell>
              <TableHeaderCell align="right">Actions</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredInstruments.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  No instruments matched your criteria.
                </TableCell>
              </TableRow>
            ) : (
              filteredInstruments.map((inst) => (
                <TableRow key={inst.id}>
                  <TableCell>
                    <div className="asset-table-symbol">
                      <span>{inst.symbol}</span>
                    </div>
                  </TableCell>
                  <TableCell>{inst.name}</TableCell>
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
      />
    </div>
  );
};
