export interface MoneyValue {
  amount: string;
  currencyCode?: string | null;
}

export function formatMoney(
  value?: MoneyValue | null | string,
  currencyCode = 'USD',
  decimals = 2
): string {
  if (value === null || value === undefined) {
    return '$0.00';
  }

  let amountStr: string;
  let code = currencyCode;

  if (typeof value === 'string') {
    amountStr = value;
  } else {
    amountStr = value.amount;
    if (value.currencyCode) {
      code = value.currencyCode;
    }
  }

  const n = Number(amountStr);
  if (isNaN(n)) return '$0.00';

  try {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: code,
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    }).format(n);
  } catch {
    return `${code} ${n.toFixed(decimals)}`;
  }
}

export function formatPercent(
  percentStr?: string | number | null,
  decimals = 2,
  includeSign = true
): string {
  if (percentStr === null || percentStr === undefined || percentStr === '') {
    return '0.00%';
  }

  const n = Number(percentStr);
  if (isNaN(n)) return '0.00%';

  const sign = includeSign && n > 0 ? '+' : '';
  return `${sign}${n.toFixed(decimals)}%`;
}

export function formatQuantity(
  qty?: string | number | null,
  maxDecimals = 6
): string {
  if (qty === null || qty === undefined || qty === '') {
    return '-';
  }

  const n = Number(qty);
  if (isNaN(n)) return String(qty);

  return new Intl.NumberFormat('en-US', {
    maximumFractionDigits: maxDecimals,
  }).format(n);
}

export function formatDecimal(
  val?: string | number | null,
  decimals = 2
): string {
  if (val === null || val === undefined || val === '') {
    return '0.00';
  }

  const n = Number(val);
  if (isNaN(n)) return '0.00';

  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(n);
}

export function formatDate(dateStr?: string | null): string {
  if (!dateStr) return '-';
  // Parse YYYY-MM-DD or ISO string without timezone drift
  const parts = dateStr.split('T')[0].split('-');
  if (parts.length === 3) {
    const year = parseInt(parts[0], 10);
    const month = parseInt(parts[1], 10) - 1;
    const day = parseInt(parts[2], 10);
    const d = new Date(Date.UTC(year, month, day));
    return d.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      timeZone: 'UTC',
    });
  }
  return dateStr;
}

export function isPositive(val?: MoneyValue | string | number | null): boolean {
  if (val === null || val === undefined) return false;
  if (typeof val === 'number') return val >= 0;
  const raw = typeof val === 'string' ? val : val.amount;
  return Number(raw) >= 0;
}
