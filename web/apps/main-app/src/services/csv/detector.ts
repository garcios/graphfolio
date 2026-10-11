export interface DetectionResult {
  detected: 'commsec' | 'nabtrade' | null;
  headerIndex: number;
  headers: string[];
}

const COMMSEC_SIGNATURES = [
  'units',
  'unit price ($)',
  'unit price',
  'brokerage ($)',
  'total value ($)',
  'net consideration',
  'asx code',
];

const NABTRADE_SIGNATURES = [
  'confirmation number',
  'trade date',
  'settlement date',
  'total consideration',
  'security code',
];

export function detectBrokerFormat(lines: string[]): DetectionResult {
  for (let i = 0; i < Math.min(lines.length, 10); i++) {
    const rawLine = lines[i].trim();
    if (!rawLine) continue;

    const lower = rawLine.toLowerCase();
    const cells = lower.split(',').map((c) => c.replace(/^["']|["']$/g, '').trim());
    const normalizedCells = cells.map((c) => c.replace(/\s*\(\$\)\s*$/, '').trim());

    const hasNabtrade = NABTRADE_SIGNATURES.some((sig) => cells.includes(sig) || lower.includes(sig));
    if (hasNabtrade) {
      return { detected: 'nabtrade', headerIndex: i, headers: cells };
    }

    const hasCommSec = COMMSEC_SIGNATURES.some((sig) => cells.includes(sig) || lower.includes(sig));
    if (hasCommSec) {
      return { detected: 'commsec', headerIndex: i, headers: cells };
    }

    // Check for cash account statements (CommSec or nabtrade)
    const isCashStatement =
      (normalizedCells.includes('description') || normalizedCells.includes('details')) &&
      normalizedCells.includes('balance') &&
      (normalizedCells.includes('debit') || normalizedCells.includes('credit'));

    if (isCashStatement) {
      const candidateRows = lines.slice(i + 1, i + 15);

      const hasCommSecRows = candidateRows.some((l) => {
        const upper = l.toUpperCase();
        return (
          /,\s*[BS]\s+\d+\s+[A-Z0-9.]+\s+@/i.test(l) ||
          upper.includes('COMMSEC') ||
          upper.includes('COMMONWEALTH') ||
          upper.includes('REJ D/TSFER') ||
          /^[A-Z0-9/]+,[CRPJ]\d{7,}/i.test(l)
        );
      });

      if (hasCommSecRows) {
        return { detected: 'commsec', headerIndex: i, headers: cells };
      }

      const hasNabtradeRows = candidateRows.some((l) => {
        const upper = l.toUpperCase();
        return (
          upper.includes('BUY ') ||
          upper.includes('SELL ') ||
          upper.includes('DIVIDEND') ||
          upper.includes('FUNDS TRANSFER') ||
          upper.includes('INTEREST') ||
          upper.includes('NABTRADE') ||
          upper.includes('.NAS') ||
          upper.includes('.NYS') ||
          upper.includes('.ASX')
        );
      });

      if (hasNabtradeRows) {
        return { detected: 'nabtrade', headerIndex: i, headers: cells };
      }
    }
  }

  return { detected: null, headerIndex: -1, headers: [] };
}
