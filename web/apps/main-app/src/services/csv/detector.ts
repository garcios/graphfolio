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
    const hasNabtrade = NABTRADE_SIGNATURES.some((sig) => cells.includes(sig) || lower.includes(sig));
    if (hasNabtrade) {
      return { detected: 'nabtrade', headerIndex: i, headers: cells };
    }

    const hasCommSec = COMMSEC_SIGNATURES.some((sig) => cells.includes(sig) || lower.includes(sig));
    if (hasCommSec) {
      return { detected: 'commsec', headerIndex: i, headers: cells };
    }

    // Check for NAB International: header Date,Description,Debit,Credit,Balance with broker trade/dividend rows
    const isCashStatement =
      cells.includes('description') &&
      cells.includes('balance') &&
      (cells.includes('debit') || cells.includes('credit'));

    if (isCashStatement) {
      const hasBrokerRows = lines.slice(i + 1, i + 10).some((l) => {
        const upper = l.toUpperCase();
        return (
          upper.includes('BUY ') ||
          upper.includes('SELL ') ||
          upper.includes('DIVIDEND ON ') ||
          upper.includes('.NAS') ||
          upper.includes('.NYS')
        );
      });

      if (hasBrokerRows) {
        return { detected: 'nabtrade', headerIndex: i, headers: cells };
      }
    }
  }

  return { detected: null, headerIndex: -1, headers: [] };
}
