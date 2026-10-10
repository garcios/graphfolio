/**
 * Strips currency symbols, commas, and whitespace.
 * Handles accounting negatives: "(19.95)" -> "19.95", "-$19.95" -> "19.95"
 */
export function cleanCurrency(val: string | undefined | null): string {
  if (!val) return '0.00';
  let cleaned = val.trim().replace(/[$ AUD,]/gi, '');
  if (cleaned.startsWith('(') && cleaned.endsWith(')')) {
    cleaned = cleaned.slice(1, -1);
  }
  if (cleaned.startsWith('-')) {
    cleaned = cleaned.slice(1);
  }
  return cleaned.trim() || '0.00';
}

/**
 * Parses Australian DD/MM/YYYY, D/M/YYYY or ISO YYYY-MM-DD into ISO YYYY-MM-DD.
 */
export function parseAustralianDate(val: string): string {
  const trimmed = val.trim();
  if (/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) {
    return trimmed;
  }
  const parts = trimmed.split(/[/.-]/);
  if (parts.length === 3) {
    let [day, month, year] = parts;
    if (day.length === 4 && year.length <= 2) {
      return `${day}-${String(parseInt(month, 10)).padStart(2, '0')}-${String(parseInt(year, 10)).padStart(2, '0')}`;
    }
    if (year.length === 2) {
      year = `20${year}`;
    }
    const d = parseInt(day, 10);
    const m = parseInt(month, 10);
    const y = parseInt(year, 10);
    if (m >= 1 && m <= 12 && d >= 1 && d <= 31 && y >= 1990 && y <= 2100) {
      return `${year}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
    }
  }
  throw new Error(`Invalid date format: "${val}". Expected DD/MM/YYYY or YYYY-MM-DD.`);
}

/**
 * Generates an idempotent SHA-256 hash using Web Crypto API.
 */
export async function generateIdempotentHash(seed: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(seed);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('').slice(0, 32);
}

/**
 * Splits CSV content into lines, normalizing CRLF and LF.
 */
export function splitCSVLines(content: string): string[] {
  return content
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
    .split('\n');
}

/**
 * Robust CSV line tokenizer that respects quotes and double-quote escapes.
 */
export function parseCSVLine(line: string): string[] {
  const result: string[] = [];
  let current = '';
  let inQuotes = false;

  for (let i = 0; i < line.length; i++) {
    const char = line[i];
    if (char === '"') {
      if (inQuotes && line[i + 1] === '"') {
        current += '"';
        i++; // skip escaped quote
      } else {
        inQuotes = !inQuotes;
      }
    } else if (char === ',' && !inQuotes) {
      result.push(current.trim().replace(/^["']|["']$/g, ''));
      current = '';
    } else {
      current += char;
    }
  }
  result.push(current.trim().replace(/^["']|["']$/g, ''));
  return result;
}
