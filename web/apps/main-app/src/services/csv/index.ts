import type { SupportedBroker, ParseResult } from './types';
import { splitCSVLines } from './utils';
import { detectBrokerFormat } from './detector';
import { parseCommSecCSV } from './parsers/commsec';
import { parseNabtradeCSV } from './parsers/nabtrade';

export * from './types';
export * from './utils';
export * from './detector';
export * from './parsers/commsec';
export * from './parsers/nabtrade';

export async function parseBrokerCSV(
  content: string,
  forcedBroker: SupportedBroker = 'auto'
): Promise<ParseResult> {
  const lines = splitCSVLines(content);
  if (lines.length === 0 || lines.every((l) => !l.trim())) {
    throw new Error('The selected CSV file is empty.');
  }

  const detection = detectBrokerFormat(lines);

  let broker: 'commsec' | 'nabtrade';
  let headerIndex: number;

  if (forcedBroker === 'auto') {
    if (!detection.detected) {
      throw new Error(
        'Unable to automatically detect broker format. Supported brokers: CommSec, nabtrade. Please select broker manually.'
      );
    }
    broker = detection.detected;
    headerIndex = detection.headerIndex;
  } else {
    broker = forcedBroker;
    headerIndex = detection.headerIndex !== -1 ? detection.headerIndex : 0;
  }

  if (broker === 'commsec') {
    return parseCommSecCSV(lines, headerIndex);
  } else {
    return parseNabtradeCSV(lines, headerIndex);
  }
}
