import { describe, it, expect } from 'vitest';
import { parseCommSecCSV } from './commsec';

describe('CommSec CSV Parser', () => {
  it('parses standard buy and sell trade rows with commas and dollar signs', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '15/03/2024,Buy,BHP,"1,000","$45.20","$19.95","$45,219.95"',
      '16/03/2024,Sell,CBA,200,$115.50,$19.95,"$23,080.05"',
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.broker).toBe('commsec');
    expect(res.validRows.length).toBe(2);
    expect(res.errors.length).toBe(0);

    const row1 = res.validRows[0];
    expect(row1.type).toBe('BUY');
    expect(row1.symbol).toBe('BHP');
    expect(row1.tradeDate).toBe('2024-03-15');
    expect(row1.quantity).toBe('1000');
    expect(row1.price).toBe('45.20');
    expect(row1.fee).toBe('19.95');
    expect(row1.amount).toBe('45219.95');
    expect(row1.externalRef).toMatch(/^cs_[a-f0-9]{32}$/);

    const row2 = res.validRows[1];
    expect(row2.type).toBe('SELL');
    expect(row2.symbol).toBe('CBA');
    expect(row2.tradeDate).toBe('2024-03-16');
    expect(row2.quantity).toBe('200');
    expect(row2.price).toBe('115.50');
    expect(row2.amount).toBe('23080.05');
  });

  it('parses dividend rows where quantity/unit price are omitted', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '20/03/2024,Div,VAS,,,0.00,350.00',
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.validRows.length).toBe(1);
    expect(res.errors.length).toBe(0);

    const div = res.validRows[0];
    expect(div.type).toBe('DIVIDEND');
    expect(div.symbol).toBe('VAS');
    expect(div.amount).toBe('350.00');
    expect(div.fee).toBe('0.00');
  });

  it('reconstructs total value if missing for Buy and Sell', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '10/01/2024,Buy,TLS,100,4.00,10.00,',
      '11/01/2024,Sell,TLS,50,4.00,5.00,',
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.validRows.length).toBe(2);
    // Buy total = 100 * 4.00 + 10.00 = 410.00
    expect(res.validRows[0].amount).toBe('410.00');
    // Sell total = 50 * 4.00 - 5.00 = 195.00
    expect(res.validRows[1].amount).toBe('195.00');
  });

  it('handles accounting negative parentheses and strips .AX suffix', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '15/03/2024,Sell,MQG.AX,10,180.00,(19.95),(1780.05)',
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.validRows.length).toBe(1);
    expect(res.validRows[0].symbol).toBe('MQG');
    expect(res.validRows[0].fee).toBe('19.95');
    expect(res.validRows[0].amount).toBe('1780.05');
  });

  it('records malformed rows without crashing valid rows', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '32/13/2024,Buy,BHP,10,45.00,0,450.00', // Invalid date
      '15/03/2024,Buy,BHP,0,45.00,0,450.00',   // Zero quantity on buy
      '15/03/2024,Buy,BHP,10,45.00,0,450.00',   // Valid
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.validRows.length).toBe(1);
    expect(res.errors.length).toBe(2);
    expect(res.errors[0].rowNumber).toBe(2);
    expect(res.errors[0].message).toMatch(/date/i);
    expect(res.errors[1].rowNumber).toBe(3);
    expect(res.errors[1].message).toMatch(/quantity/i);
  });

  describe('CommSec Cash Account / Transaction Statement Parser', () => {
    it('parses trade buy/sell, deposit, and withdrawal rows from CommSec cash statement', async () => {
      const lines = [
        'Date,Reference,Details,Debit($),Credit($),Balance($)',
        '24/06/2026,P36741365,Direct Transfer - Payee MR TOM CRUZ,7960.00,,0.00',
        '22/06/2026,C176305381,S 125 NDQ @ 63.680000  ,,7960.00,-7960.00',
        '09/06/2026,R71985282,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,251.22,0.00',
        '04/06/2026,C175717445,B 4 NDQ @ 62.806028  ,251.22,,251.22',
        '19/07/2024,J21882301049,Rej D/Tsfer B/S 18/07/2024 DE REFER TO CUST,985.84,,2260.90',
        '29/08/2024,C154772543,B 23 NDQ @ 41.950000  ,966.85,,966.85',
      ];

      const res = await parseCommSecCSV(lines, 0);
      expect(res.broker).toBe('commsec');
      expect(res.validRows.length).toBe(6);
      expect(res.errors.length).toBe(0);

      // Rows are sorted chronologically ascending
      const sortedRows = res.validRows;

      // 1. Rej D/Tsfer on 2024-07-19 -> WITHDRAWAL
      const rejWth = sortedRows.find((r) => r.externalRef === 'commsec:cash:J21882301049');
      expect(rejWth).toBeDefined();
      expect(rejWth?.type).toBe('WITHDRAWAL');
      expect(rejWth?.amount).toBe('985.84');
      expect(rejWth?.symbol).toBe('');

      // 2. Buy with brokerage fee: 23 * 41.95 = 964.85, Debit = 966.85 -> fee = 2.00
      const buyWithFee = sortedRows.find((r) => r.externalRef === 'commsec:C154772543');
      expect(buyWithFee).toBeDefined();
      expect(buyWithFee?.type).toBe('BUY');
      expect(buyWithFee?.symbol).toBe('NDQ');
      expect(buyWithFee?.quantity).toBe('23');
      expect(buyWithFee?.price).toBe('41.95');
      expect(buyWithFee?.fee).toBe('2.00');
      expect(buyWithFee?.amount).toBe('964.85');

      // 3. Buy with zero fee / fractional price: 4 * 62.806028 = 251.22
      const buyZeroFee = sortedRows.find((r) => r.externalRef === 'commsec:C175717445');
      expect(buyZeroFee).toBeDefined();
      expect(buyZeroFee?.type).toBe('BUY');
      expect(buyZeroFee?.symbol).toBe('NDQ');
      expect(buyZeroFee?.quantity).toBe('4');
      expect(buyZeroFee?.price).toBe('62.806028');
      expect(buyZeroFee?.fee).toBe('0.00');
      expect(buyZeroFee?.amount).toBe('251.22');

      // 4. Deposit: 251.22
      const dep = sortedRows.find((r) => r.externalRef === 'commsec:cash:R71985282');
      expect(dep).toBeDefined();
      expect(dep?.type).toBe('DEPOSIT');
      expect(dep?.amount).toBe('251.22');

      // 5. Sell: 125 NDQ @ 63.68 = 7960.00
      const sell = sortedRows.find((r) => r.externalRef === 'commsec:C176305381');
      expect(sell).toBeDefined();
      expect(sell?.type).toBe('SELL');
      expect(sell?.symbol).toBe('NDQ');
      expect(sell?.quantity).toBe('125');
      expect(sell?.price).toBe('63.68');
      expect(sell?.fee).toBe('0.00');
      expect(sell?.amount).toBe('7960.00');

      // 6. Direct Transfer Out: 7960.00 -> WITHDRAWAL
      const wth = sortedRows.find((r) => r.externalRef === 'commsec:cash:P36741365');
      expect(wth).toBeDefined();
      expect(wth?.type).toBe('WITHDRAWAL');
      expect(wth?.amount).toBe('7960.00');
    });

    it('parses the full user CommSec statement with 84 rows and 0 errors', async () => {
      const csv = `Date,Reference,Details,Debit($),Credit($),Balance($)
24/06/2026,P36741365,Direct Transfer - Payee MR TOM CRUZ,7960.00,,0.00
22/06/2026,C176305381,S 125 NDQ @ 63.680000  ,,7960.00,-7960.00
09/06/2026,R71985282,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,251.22,0.00
04/06/2026,C175717445,B 4 NDQ @ 62.806028  ,251.22,,251.22
06/05/2026,R71543549,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,283.38,0.00
04/05/2026,C174591268,B 5 NDQ @ 56.676268  ,283.38,,283.38
09/04/2026,R71170325,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,259.55,0.00
07/04/2026,C173645330,B 5 NDQ @ 51.509316  ,259.55,,259.55
06/03/2026,R70711020,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,260.55,0.00
04/03/2026,C172407129,B 5 NDQ @ 51.710000  ,260.55,,260.55
06/02/2026,R70292157,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,267.60,0.00
04/02/2026,C171263335,B 5 NDQ @ 53.120000  ,267.60,,267.60
06/01/2026,R69793597,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,281.26,0.00
02/01/2026,C169953959,B 5 NDQ @ 55.851411  ,281.26,,281.26
04/12/2025,R69465061,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,287.38,0.00
02/12/2025,C169110510,B 5 NDQ @ 57.075164  ,287.38,,287.38
05/11/2025,R69031168,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,293.99,0.00
03/11/2025,C168043107,B 5 NDQ @ 58.397816  ,293.99,,293.99
06/10/2025,R68517602,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,278.89,0.00
02/10/2025,C166636646,B 5 NDQ @ 55.377142  ,278.89,,278.89
04/09/2025,R68083516,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,265.13,0.00
02/09/2025,C165528055,B 5 NDQ @ 52.626861  ,265.13,,265.13
06/08/2025,R67662099,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,261.71,0.00
04/08/2025,C164447105,B 5 NDQ @ 51.942653  ,261.71,,261.71
04/07/2025,R67259940,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,254.06,0.00
02/07/2025,C163434305,B 5 NDQ @ 50.412236  ,254.06,,254.06
04/06/2025,R66924936,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,295.81,0.00
02/06/2025,C162518949,B 6 NDQ @ 48.968952  ,295.81,,295.81
06/05/2025,R66607704,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,277.11,0.00
02/05/2025,C161708325,B 6 NDQ @ 45.851698  ,277.11,,277.11
29/04/2025,R66525700,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,264.74,0.00
24/04/2025,C161520123,B 6 NDQ @ 43.790000  ,264.74,,264.74
08/04/2025,R66229607,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,264.74,0.00
04/04/2025,C160839357,B 6 NDQ @ 43.790000  ,264.74,,264.74
04/04/2025,R66176333,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,278.49,0.00
02/04/2025,C160698861,B 6 NDQ @ 46.081099  ,278.49,,278.49
05/03/2025,R65805748,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,252.60,0.00
03/03/2025,C159826414,B 5 NDQ @ 50.120000  ,252.60,,252.60
05/02/2025,R65399522,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,255.63,0.00
03/02/2025,C158948804,B 5 NDQ @ 50.725495  ,255.63,,255.63
06/01/2025,R65040995,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,255.42,0.00
02/01/2025,C158137135,B 5 NDQ @ 50.684647  ,255.42,,255.42
04/12/2024,R64692951,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,289.07,0.00
02/12/2024,C157392283,B 6 NDQ @ 47.844276  ,289.07,,289.07
06/11/2024,R64386576,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,272.23,0.00
04/11/2024,C156584584,B 6 NDQ @ 45.038038  ,272.23,,272.23
04/10/2024,R64029434,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,257.87,0.00
02/10/2024,C155678072,B 6 NDQ @ 42.644995  ,257.87,,257.87
04/09/2024,R63726773,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,259.33,0.00
02/09/2024,C154848570,B 6 NDQ @ 42.887883  ,259.33,,259.33
02/09/2024,R63695711,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,966.85,0.00
29/08/2024,C154772543,B 23 NDQ @ 41.950000  ,966.85,,966.85
06/08/2024,R63403215,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,259.89,0.00
02/08/2024,C153984858,B 6 NDQ @ 42.981304  ,259.89,,259.89
31/07/2024,R63318172,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,4781.35,0.00
29/07/2024,C153842161,B 110 NDQ @ 43.380000  ,4781.35,,4781.35
29/07/2024,R63292115,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,998.59,0.00
25/07/2024,C153758674,B 23 NDQ @ 43.330000  ,998.59,,998.59
22/07/2024,R63213675,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,1275.06,0.00
19/07/2024,R63200384,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,985.84,1275.06
19/07/2024,J21882301049,Rej D/Tsfer B/S 18/07/2024 DE REFER TO CUST,985.84,,2260.90
18/07/2024,C153559317,B 29 NDQ @ 43.880000,1275.06,,1275.06
17/07/2024,R63167985,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,985.84,0.00
16/07/2024,R63153409,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,980.12,985.84
15/07/2024,C153433762,B 22 NDQ @ 44.720000,985.84,,1965.96
15/07/2024,R63140335,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,956.45,980.12
12/07/2024,C153394695,B 22 NDQ @ 44.460000,980.12,,1936.57
11/07/2024,C153365050,B 21 NDQ @ 45.450000,956.45,,956.45
11/07/2024,R63117631,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,996.40,0.00
10/07/2024,R63106492,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,87.85,996.40
09/07/2024,C153294883,B 22 NDQ @ 45.200000,996.40,,1084.25
08/07/2024,C153249652,B 44 NDQ @ 44.850000,1977.35,,87.85
08/07/2024,C153249330,S 15 IOO @ 145.800000,,2182.62,-1889.50
05/07/2024,C153222667,B 2 IOO @ 145.557614,293.12,,293.12
07/06/2024,R62804469,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,282.33,0.00
05/06/2024,C152436530,B 2 IOO @ 140.167200,282.33,,282.33
08/05/2024,R62494936,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,270.38,0.00
06/05/2024,C151643247,B 2 IOO @ 134.190664,270.38,,270.38
11/07/2023,R59750768,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,449.96,0.00
07/07/2023,C144256667,B 4 IOO @ 111.990000,449.96,,449.96
21/06/2023,R59577942,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,445.72,0.00
19/06/2023,C143767049,B 4 IOO @ 110.930000,445.72,,445.72
14/06/2023,R59518879,Direct Transfer 063676 10094893 Drawer MR TOM CRUZ,,112.00,0.00
09/06/2023,C143575879,B 1 IOO @ 110.000000,112.00,,112.00`;

      const lines = csv.trim().split('\n');
      const res = await parseCommSecCSV(lines, 0);

      expect(res.broker).toBe('commsec');
      expect(res.totalRowsRead).toBe(84);
      expect(res.validRows.length).toBe(84);
      expect(res.errors.length).toBe(0);

      const buys = res.validRows.filter((r) => r.type === 'BUY');
      const sells = res.validRows.filter((r) => r.type === 'SELL');
      const deposits = res.validRows.filter((r) => r.type === 'DEPOSIT');
      const withdrawals = res.validRows.filter((r) => r.type === 'WITHDRAWAL');

      expect(buys.length).toBe(40);
      expect(sells.length).toBe(2);
      expect(deposits.length).toBe(40);
      expect(withdrawals.length).toBe(2);

      // Verify share balance net:
      // IOO: Bought 15, Sold 15 -> Net 0
      const iooBuys = buys.filter((r) => r.symbol === 'IOO').reduce((sum, r) => sum + parseFloat(r.quantity), 0);
      const iooSells = sells.filter((r) => r.symbol === 'IOO').reduce((sum, r) => sum + parseFloat(r.quantity), 0);
      expect(iooBuys).toBe(15);
      expect(iooSells).toBe(15);

      // NDQ: Bought 450, Sold 125 -> Net 325
      const ndqBuys = buys.filter((r) => r.symbol === 'NDQ').reduce((sum, r) => sum + parseFloat(r.quantity), 0);
      const ndqSells = sells.filter((r) => r.symbol === 'NDQ').reduce((sum, r) => sum + parseFloat(r.quantity), 0);
      expect(ndqBuys).toBe(450);
      expect(ndqSells).toBe(125);
    });
  });
});
