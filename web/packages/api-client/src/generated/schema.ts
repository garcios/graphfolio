// @ts-nocheck
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type Scalars = {
    Decimal: any,
    String: string,
    ID: string,
    Boolean: boolean,
    Int: number,
}

export interface Money {
    amount: Scalars['Decimal']
    currencyCode: Scalars['String']
    __typename: 'Money'
}

export type TransactionType = 'BUY' | 'SELL' | 'DIVIDEND' | 'DEPOSIT' | 'WITHDRAWAL' | 'INTEREST' | 'FEE' | 'TAX' | 'TRANSFER_IN' | 'TRANSFER_OUT' | 'FX_CONVERSION' | 'SPLIT'

export interface AddTransactionPayload {
    transactionId: Scalars['String']
    portfolio: Portfolio
    __typename: 'AddTransactionPayload'
}

export interface Instrument {
    id: Scalars['ID']
    symbol: Scalars['String']
    name: Scalars['String']
    currencyCode: Scalars['String']
    assetClass: Scalars['String']
    exchangeCode: Scalars['String']
    isin: (Scalars['String'] | null)
    isActive: Scalars['Boolean']
    __typename: 'Instrument'
}


/** Trading venue reference data identified by its ISO 10383 MIC code. */
export interface Exchange {
    code: Scalars['String']
    name: Scalars['String']
    country: Scalars['String']
    timezone: Scalars['String']
    __typename: 'Exchange'
}

export interface Query {
    portfolio: Portfolio
    instruments: Instrument[]
    portfolioHistory: PortfolioHistory
    transactions: TransactionsConnection
    allInstruments: Instrument[]
    exchanges: Exchange[]
    instrumentPrices: InstrumentPricesConnection
    ingestionStatus: IngestionStatus
    userPreferences: UserPreferences
    supportedCurrencies: Currency[]
    currencyPairs: CurrencyPair[]
    currencyPairHistory: CurrencyPairHistory
    fxRates: FXRatesConnection
    checkTransactionDuplicates: Scalars['String'][]
    cashFlowReport: CashFlowReport
    __typename: 'Query'
}

export type HistoryTimeframe = 'TIMEFRAME_1D' | 'TIMEFRAME_1W' | 'TIMEFRAME_1M' | 'TIMEFRAME_1Y' | 'TIMEFRAME_ALL'

export interface ValuationPoint {
    date: Scalars['String']
    totalValue: Money
    marketValue: Money
    cashValue: Money
    twrIndex: Scalars['Decimal']
    dailyReturn: (Scalars['Decimal'] | null)
    __typename: 'ValuationPoint'
}

export interface PortfolioHistory {
    points: ValuationPoint[]
    startValue: Money
    endValue: Money
    returnAmount: Money
    returnPercent: Scalars['Decimal']
    __typename: 'PortfolioHistory'
}

export interface TransactionItem {
    id: Scalars['ID']
    type: TransactionType
    symbol: (Scalars['String'] | null)
    instrumentName: (Scalars['String'] | null)
    tradeDate: Scalars['String']
    quantity: (Scalars['Decimal'] | null)
    price: (Money | null)
    amount: Money
    fee: Money
    notes: (Scalars['String'] | null)
    createdAt: Scalars['String']
    __typename: 'TransactionItem'
}

export interface TransactionsConnection {
    items: TransactionItem[]
    totalCount: Scalars['Int']
    page: Scalars['Int']
    pageSize: Scalars['Int']
    __typename: 'TransactionsConnection'
}

export interface DeleteTransactionPayload {
    success: Scalars['Boolean']
    portfolio: Portfolio
    __typename: 'DeleteTransactionPayload'
}

export interface DeleteInstrumentPayload {
    success: Scalars['Boolean']
    id: Scalars['ID']
    __typename: 'DeleteInstrumentPayload'
}

export interface Mutation {
    addTransaction: AddTransactionPayload
    deleteTransaction: DeleteTransactionPayload
    createInstrument: Instrument
    updateInstrument: Instrument
    deleteInstrument: DeleteInstrumentPayload
    recordPriceOverride: RecordPriceOverridePayload
    triggerMarketSync: MarketSyncPayload
    triggerBackfill: BackfillPayload
    updateUserPreferences: UpdateUserPreferencesPayload
    recordFXRateOverride: RecordFXRateOverridePayload
    importTransactions: BatchImportTransactionsPayload
    __typename: 'Mutation'
}

export interface BatchImportTransactionsPayload {
    success: Scalars['Boolean']
    importedCount: Scalars['Int']
    skippedCount: Scalars['Int']
    portfolio: Portfolio
    message: Scalars['String']
    __typename: 'BatchImportTransactionsPayload'
}

export interface InstrumentPrice {
    id: Scalars['ID']
    symbol: Scalars['String']
    priceDate: Scalars['String']
    price: Money
    source: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'InstrumentPrice'
}

export interface InstrumentPricesConnection {
    items: InstrumentPrice[]
    totalCount: Scalars['Int']
    __typename: 'InstrumentPricesConnection'
}

export interface RecordPriceOverridePayload {
    price: InstrumentPrice
    valuationsRecomputed: Scalars['Boolean']
    __typename: 'RecordPriceOverridePayload'
}

export interface FeedHealthStatus {
    name: Scalars['String']
    status: Scalars['String']
    provider: Scalars['String']
    schedule: Scalars['String']
    lastRun: Scalars['String']
    details: Scalars['String']
    __typename: 'FeedHealthStatus'
}

export interface IngestionStatus {
    feeds: FeedHealthStatus[]
    trackedInstruments: Scalars['Int']
    trackedCurrencies: Scalars['Int']
    latestPriceDate: (Scalars['String'] | null)
    latestFxDate: (Scalars['String'] | null)
    rateLimitRemaining: Scalars['Int']
    rateLimitBudget: Scalars['Int']
    pendingBackfillJobs: Scalars['Int']
    __typename: 'IngestionStatus'
}

export interface MarketSyncPayload {
    success: Scalars['Boolean']
    pricesSynced: Scalars['Int']
    fxRatesSynced: Scalars['Int']
    message: Scalars['String']
    __typename: 'MarketSyncPayload'
}

export interface BackfillPayload {
    success: Scalars['Boolean']
    pricesSynced: Scalars['Int']
    fxRatesSynced: Scalars['Int']
    message: Scalars['String']
    warnings: Scalars['String'][]
    __typename: 'BackfillPayload'
}

export interface Portfolio {
    totalValue: Money
    todayReturnAmount: Money
    todayReturnPercent: Scalars['Decimal']
    annualizedReturnPercent: Scalars['Decimal']
    cashBalance: Money
    investments: Investment[]
    __typename: 'Portfolio'
}

export interface Investment {
    id: Scalars['ID']
    ticker: Scalars['String']
    name: Scalars['String']
    price: Money
    averageBuyPrice: (Money | null)
    quantity: Scalars['Decimal']
    totalValue: Money
    todayReturnAmount: Money
    todayReturnPercent: Scalars['Decimal']
    totalReturnAmount: Money
    totalReturnPercent: Scalars['Decimal']
    capitalGainAmount: Money
    capitalGainPercent: Scalars['Decimal']
    incomeAmount: Money
    incomeYieldPercent: Scalars['Decimal']
    currencyGainAmount: Money
    currencyGainPercent: Scalars['Decimal']
    isInternational: Scalars['Boolean']
    __typename: 'Investment'
}

export interface UserPreferences {
    userId: Scalars['ID']
    email: Scalars['String']
    displayName: Scalars['String']
    displayCurrency: Scalars['String']
    theme: Scalars['String']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'UserPreferences'
}

export interface Currency {
    code: Scalars['String']
    name: Scalars['String']
    symbol: Scalars['String']
    __typename: 'Currency'
}

export interface UpdateUserPreferencesPayload {
    preferences: UserPreferences
    portfolio: (Portfolio | null)
    __typename: 'UpdateUserPreferencesPayload'
}

export interface CurrencyPair {
    baseCurrency: Scalars['String']
    quoteCurrency: Scalars['String']
    pair: Scalars['String']
    latestRate: Scalars['Decimal']
    latestDate: Scalars['String']
    latestSource: Scalars['String']
    previousRate: (Scalars['Decimal'] | null)
    change1dAmount: (Scalars['Decimal'] | null)
    change1dPct: (Scalars['Decimal'] | null)
    totalRecords: Scalars['Int']
    firstDate: Scalars['String']
    lastDate: Scalars['String']
    __typename: 'CurrencyPair'
}

export interface FXRate {
    baseCurrency: Scalars['String']
    quoteCurrency: Scalars['String']
    pair: Scalars['String']
    rateDate: Scalars['String']
    rate: Scalars['Decimal']
    invertedRate: Scalars['Decimal']
    source: Scalars['String']
    __typename: 'FXRate'
}

export interface FXRatesConnection {
    items: FXRate[]
    totalCount: Scalars['Int']
    __typename: 'FXRatesConnection'
}

export interface FXHistoryPoint {
    date: Scalars['String']
    rate: Scalars['Decimal']
    invertedRate: Scalars['Decimal']
    source: Scalars['String']
    __typename: 'FXHistoryPoint'
}

export interface CurrencyPairHistory {
    baseCurrency: Scalars['String']
    quoteCurrency: Scalars['String']
    pair: Scalars['String']
    points: FXHistoryPoint[]
    startRate: Scalars['Decimal']
    endRate: Scalars['Decimal']
    periodChange: Scalars['Decimal']
    periodChangePct: Scalars['Decimal']
    periodHigh: Scalars['Decimal']
    periodLow: Scalars['Decimal']
    __typename: 'CurrencyPairHistory'
}

export interface RecordFXRateOverridePayload {
    rate: FXRate
    valuationsRecomputed: Scalars['Boolean']
    __typename: 'RecordFXRateOverridePayload'
}

export type CashFlowTimeframe = 'MTD' | 'YTD' | 'M1' | 'M3' | 'M6' | 'Y1' | 'ALL' | 'CUSTOM'

export type CashFlowDirection = 'INFLOW' | 'OUTFLOW'

export type CashFlowCategory = 'CAPITAL_DEPOSITS' | 'DIVIDENDS' | 'INTEREST' | 'SALE_PROCEEDS' | 'CAPITAL_WITHDRAWALS' | 'PURCHASES' | 'FEES' | 'TAXES'

export interface CashFlowSummary {
    startingCashBalance: Money
    totalInflows: Money
    totalOutflows: Money
    netCashFlow: Money
    endingCashBalance: Money
    __typename: 'CashFlowSummary'
}

export interface CashFlowCategoryBreakdown {
    deposits: Money
    dividends: Money
    interest: Money
    salesProceeds: Money
    withdrawals: Money
    purchases: Money
    fees: Money
    taxes: Money
    __typename: 'CashFlowCategoryBreakdown'
}

export interface CashFlowItem {
    id: Scalars['ID']
    eventDate: Scalars['String']
    type: TransactionType
    flowDirection: CashFlowDirection
    category: CashFlowCategory
    symbol: (Scalars['String'] | null)
    instrumentName: (Scalars['String'] | null)
    description: Scalars['String']
    netAmount: Money
    runningBalance: Money
    localAmount: Money
    fee: Money
    withholdingTax: Money
    __typename: 'CashFlowItem'
}

export interface CashFlowReport {
    summary: CashFlowSummary
    breakdown: CashFlowCategoryBreakdown
    items: CashFlowItem[]
    baseCurrency: Scalars['String']
    fromDate: Scalars['String']
    toDate: Scalars['String']
    __typename: 'CashFlowReport'
}

export interface MoneyGenqlSelection{
    amount?: boolean | number
    currencyCode?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface AddTransactionInput {type: TransactionType,symbol?: (Scalars['String'] | null),tradeDate: Scalars['String'],quantity?: (Scalars['Decimal'] | null),price?: (Scalars['Decimal'] | null),amount?: (Scalars['Decimal'] | null),currencyCode?: (Scalars['String'] | null),fee?: (Scalars['Decimal'] | null),feeCurrencyCode?: (Scalars['String'] | null),notes?: (Scalars['String'] | null)}

export interface AddTransactionPayloadGenqlSelection{
    transactionId?: boolean | number
    portfolio?: PortfolioGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface InstrumentGenqlSelection{
    id?: boolean | number
    symbol?: boolean | number
    name?: boolean | number
    currencyCode?: boolean | number
    assetClass?: boolean | number
    exchangeCode?: boolean | number
    isin?: boolean | number
    isActive?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}


/** Trading venue reference data identified by its ISO 10383 MIC code. */
export interface ExchangeGenqlSelection{
    code?: boolean | number
    name?: boolean | number
    country?: boolean | number
    timezone?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface QueryGenqlSelection{
    portfolio?: PortfolioGenqlSelection
    instruments?: InstrumentGenqlSelection
    portfolioHistory?: (PortfolioHistoryGenqlSelection & { __args: {timeframe: HistoryTimeframe} })
    transactions?: (TransactionsConnectionGenqlSelection & { __args?: {type?: (TransactionType | null), symbol?: (Scalars['String'] | null), page?: (Scalars['Int'] | null), pageSize?: (Scalars['Int'] | null)} })
    allInstruments?: (InstrumentGenqlSelection & { __args?: {isActive?: (Scalars['Boolean'] | null), search?: (Scalars['String'] | null)} })
    exchanges?: ExchangeGenqlSelection
    instrumentPrices?: (InstrumentPricesConnectionGenqlSelection & { __args?: {symbol?: (Scalars['String'] | null), fromDate?: (Scalars['String'] | null), toDate?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    ingestionStatus?: IngestionStatusGenqlSelection
    userPreferences?: UserPreferencesGenqlSelection
    supportedCurrencies?: CurrencyGenqlSelection
    currencyPairs?: CurrencyPairGenqlSelection
    currencyPairHistory?: (CurrencyPairHistoryGenqlSelection & { __args: {baseCurrency: Scalars['String'], quoteCurrency: Scalars['String'], timeframe: HistoryTimeframe} })
    fxRates?: (FXRatesConnectionGenqlSelection & { __args?: {baseCurrency?: (Scalars['String'] | null), quoteCurrency?: (Scalars['String'] | null), fromDate?: (Scalars['String'] | null), toDate?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    checkTransactionDuplicates?: { __args: {externalRefs: Scalars['String'][]} }
    cashFlowReport?: (CashFlowReportGenqlSelection & { __args: {filter: CashFlowFilterInput} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ValuationPointGenqlSelection{
    date?: boolean | number
    totalValue?: MoneyGenqlSelection
    marketValue?: MoneyGenqlSelection
    cashValue?: MoneyGenqlSelection
    twrIndex?: boolean | number
    dailyReturn?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PortfolioHistoryGenqlSelection{
    points?: ValuationPointGenqlSelection
    startValue?: MoneyGenqlSelection
    endValue?: MoneyGenqlSelection
    returnAmount?: MoneyGenqlSelection
    returnPercent?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TransactionItemGenqlSelection{
    id?: boolean | number
    type?: boolean | number
    symbol?: boolean | number
    instrumentName?: boolean | number
    tradeDate?: boolean | number
    quantity?: boolean | number
    price?: MoneyGenqlSelection
    amount?: MoneyGenqlSelection
    fee?: MoneyGenqlSelection
    notes?: boolean | number
    createdAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TransactionsConnectionGenqlSelection{
    items?: TransactionItemGenqlSelection
    totalCount?: boolean | number
    page?: boolean | number
    pageSize?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface DeleteTransactionPayloadGenqlSelection{
    success?: boolean | number
    portfolio?: PortfolioGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface DeleteInstrumentPayloadGenqlSelection{
    success?: boolean | number
    id?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MutationGenqlSelection{
    addTransaction?: (AddTransactionPayloadGenqlSelection & { __args: {input: AddTransactionInput} })
    deleteTransaction?: (DeleteTransactionPayloadGenqlSelection & { __args: {id: Scalars['ID']} })
    createInstrument?: (InstrumentGenqlSelection & { __args: {input: CreateInstrumentInput} })
    updateInstrument?: (InstrumentGenqlSelection & { __args: {input: UpdateInstrumentInput} })
    deleteInstrument?: (DeleteInstrumentPayloadGenqlSelection & { __args: {id: Scalars['ID']} })
    recordPriceOverride?: (RecordPriceOverridePayloadGenqlSelection & { __args: {input: RecordPriceOverrideInput} })
    triggerMarketSync?: (MarketSyncPayloadGenqlSelection & { __args?: {symbols?: (Scalars['String'][] | null), syncFx?: (Scalars['Boolean'] | null)} })
    triggerBackfill?: (BackfillPayloadGenqlSelection & { __args: {input: TriggerBackfillInput} })
    updateUserPreferences?: (UpdateUserPreferencesPayloadGenqlSelection & { __args: {input: UpdateUserPreferencesInput} })
    recordFXRateOverride?: (RecordFXRateOverridePayloadGenqlSelection & { __args: {input: RecordFXRateOverrideInput} })
    importTransactions?: (BatchImportTransactionsPayloadGenqlSelection & { __args: {input: BatchImportTransactionsInput} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ImportTransactionInput {externalRef: Scalars['String'],symbol: Scalars['String'],type: TransactionType,tradeDate: Scalars['String'],settleDate?: (Scalars['String'] | null),quantity: Scalars['Decimal'],price: Scalars['Decimal'],amount: Scalars['Decimal'],fee: Scalars['Decimal'],feeCurrencyCode?: (Scalars['String'] | null),currencyCode?: (Scalars['String'] | null),notes?: (Scalars['String'] | null)}

export interface BatchImportTransactionsInput {transactions: ImportTransactionInput[],skipDuplicates?: (Scalars['Boolean'] | null)}

export interface BatchImportTransactionsPayloadGenqlSelection{
    success?: boolean | number
    importedCount?: boolean | number
    skippedCount?: boolean | number
    portfolio?: PortfolioGenqlSelection
    message?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface InstrumentPriceGenqlSelection{
    id?: boolean | number
    symbol?: boolean | number
    priceDate?: boolean | number
    price?: MoneyGenqlSelection
    source?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface InstrumentPricesConnectionGenqlSelection{
    items?: InstrumentPriceGenqlSelection
    totalCount?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CreateInstrumentInput {symbol: Scalars['String'],exchangeCode: Scalars['String'],name: Scalars['String'],assetClass: Scalars['String'],currencyCode: Scalars['String'],isin?: (Scalars['String'] | null)}

export interface UpdateInstrumentInput {id: Scalars['ID'],name?: (Scalars['String'] | null),isActive?: (Scalars['Boolean'] | null),isin?: (Scalars['String'] | null)}

export interface RecordPriceOverrideInput {symbol: Scalars['String'],priceDate: Scalars['String'],price: Scalars['Decimal'],reason?: (Scalars['String'] | null),recomputeValuations?: (Scalars['Boolean'] | null)}

export interface RecordPriceOverridePayloadGenqlSelection{
    price?: InstrumentPriceGenqlSelection
    valuationsRecomputed?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FeedHealthStatusGenqlSelection{
    name?: boolean | number
    status?: boolean | number
    provider?: boolean | number
    schedule?: boolean | number
    lastRun?: boolean | number
    details?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface IngestionStatusGenqlSelection{
    feeds?: FeedHealthStatusGenqlSelection
    trackedInstruments?: boolean | number
    trackedCurrencies?: boolean | number
    latestPriceDate?: boolean | number
    latestFxDate?: boolean | number
    rateLimitRemaining?: boolean | number
    rateLimitBudget?: boolean | number
    pendingBackfillJobs?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MarketSyncPayloadGenqlSelection{
    success?: boolean | number
    pricesSynced?: boolean | number
    fxRatesSynced?: boolean | number
    message?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TriggerBackfillInput {fromDate: Scalars['String'],toDate: Scalars['String'],symbols?: (Scalars['String'][] | null),currencyPairs?: (Scalars['String'][] | null),backfillAssets?: (Scalars['Boolean'] | null),backfillFx?: (Scalars['Boolean'] | null),recomputeValuations?: (Scalars['Boolean'] | null)}

export interface BackfillPayloadGenqlSelection{
    success?: boolean | number
    pricesSynced?: boolean | number
    fxRatesSynced?: boolean | number
    message?: boolean | number
    warnings?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PortfolioGenqlSelection{
    totalValue?: MoneyGenqlSelection
    todayReturnAmount?: MoneyGenqlSelection
    todayReturnPercent?: boolean | number
    annualizedReturnPercent?: boolean | number
    cashBalance?: MoneyGenqlSelection
    investments?: InvestmentGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface InvestmentGenqlSelection{
    id?: boolean | number
    ticker?: boolean | number
    name?: boolean | number
    price?: MoneyGenqlSelection
    averageBuyPrice?: MoneyGenqlSelection
    quantity?: boolean | number
    totalValue?: MoneyGenqlSelection
    todayReturnAmount?: MoneyGenqlSelection
    todayReturnPercent?: boolean | number
    totalReturnAmount?: MoneyGenqlSelection
    totalReturnPercent?: boolean | number
    capitalGainAmount?: MoneyGenqlSelection
    capitalGainPercent?: boolean | number
    incomeAmount?: MoneyGenqlSelection
    incomeYieldPercent?: boolean | number
    currencyGainAmount?: MoneyGenqlSelection
    currencyGainPercent?: boolean | number
    isInternational?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UserPreferencesGenqlSelection{
    userId?: boolean | number
    email?: boolean | number
    displayName?: boolean | number
    displayCurrency?: boolean | number
    theme?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CurrencyGenqlSelection{
    code?: boolean | number
    name?: boolean | number
    symbol?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UpdateUserPreferencesInput {displayName?: (Scalars['String'] | null),displayCurrency?: (Scalars['String'] | null),theme?: (Scalars['String'] | null)}

export interface UpdateUserPreferencesPayloadGenqlSelection{
    preferences?: UserPreferencesGenqlSelection
    portfolio?: PortfolioGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CurrencyPairGenqlSelection{
    baseCurrency?: boolean | number
    quoteCurrency?: boolean | number
    pair?: boolean | number
    latestRate?: boolean | number
    latestDate?: boolean | number
    latestSource?: boolean | number
    previousRate?: boolean | number
    change1dAmount?: boolean | number
    change1dPct?: boolean | number
    totalRecords?: boolean | number
    firstDate?: boolean | number
    lastDate?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FXRateGenqlSelection{
    baseCurrency?: boolean | number
    quoteCurrency?: boolean | number
    pair?: boolean | number
    rateDate?: boolean | number
    rate?: boolean | number
    invertedRate?: boolean | number
    source?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FXRatesConnectionGenqlSelection{
    items?: FXRateGenqlSelection
    totalCount?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FXHistoryPointGenqlSelection{
    date?: boolean | number
    rate?: boolean | number
    invertedRate?: boolean | number
    source?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CurrencyPairHistoryGenqlSelection{
    baseCurrency?: boolean | number
    quoteCurrency?: boolean | number
    pair?: boolean | number
    points?: FXHistoryPointGenqlSelection
    startRate?: boolean | number
    endRate?: boolean | number
    periodChange?: boolean | number
    periodChangePct?: boolean | number
    periodHigh?: boolean | number
    periodLow?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface RecordFXRateOverrideInput {baseCurrency: Scalars['String'],quoteCurrency: Scalars['String'],rateDate: Scalars['String'],rate: Scalars['Decimal'],reason?: (Scalars['String'] | null),recomputeValuations?: (Scalars['Boolean'] | null)}

export interface RecordFXRateOverridePayloadGenqlSelection{
    rate?: FXRateGenqlSelection
    valuationsRecomputed?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CashFlowFilterInput {timeframe: CashFlowTimeframe,fromDate?: (Scalars['String'] | null),toDate?: (Scalars['String'] | null),currency?: (Scalars['String'] | null)}

export interface CashFlowSummaryGenqlSelection{
    startingCashBalance?: MoneyGenqlSelection
    totalInflows?: MoneyGenqlSelection
    totalOutflows?: MoneyGenqlSelection
    netCashFlow?: MoneyGenqlSelection
    endingCashBalance?: MoneyGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CashFlowCategoryBreakdownGenqlSelection{
    deposits?: MoneyGenqlSelection
    dividends?: MoneyGenqlSelection
    interest?: MoneyGenqlSelection
    salesProceeds?: MoneyGenqlSelection
    withdrawals?: MoneyGenqlSelection
    purchases?: MoneyGenqlSelection
    fees?: MoneyGenqlSelection
    taxes?: MoneyGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CashFlowItemGenqlSelection{
    id?: boolean | number
    eventDate?: boolean | number
    type?: boolean | number
    flowDirection?: boolean | number
    category?: boolean | number
    symbol?: boolean | number
    instrumentName?: boolean | number
    description?: boolean | number
    netAmount?: MoneyGenqlSelection
    runningBalance?: MoneyGenqlSelection
    localAmount?: MoneyGenqlSelection
    fee?: MoneyGenqlSelection
    withholdingTax?: MoneyGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CashFlowReportGenqlSelection{
    summary?: CashFlowSummaryGenqlSelection
    breakdown?: CashFlowCategoryBreakdownGenqlSelection
    items?: CashFlowItemGenqlSelection
    baseCurrency?: boolean | number
    fromDate?: boolean | number
    toDate?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}


    const Money_possibleTypes: string[] = ['Money']
    export const isMoney = (obj?: { __typename?: any } | null): obj is Money => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMoney"')
      return Money_possibleTypes.includes(obj.__typename)
    }
    


    const AddTransactionPayload_possibleTypes: string[] = ['AddTransactionPayload']
    export const isAddTransactionPayload = (obj?: { __typename?: any } | null): obj is AddTransactionPayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isAddTransactionPayload"')
      return AddTransactionPayload_possibleTypes.includes(obj.__typename)
    }
    


    const Instrument_possibleTypes: string[] = ['Instrument']
    export const isInstrument = (obj?: { __typename?: any } | null): obj is Instrument => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isInstrument"')
      return Instrument_possibleTypes.includes(obj.__typename)
    }
    


    const Exchange_possibleTypes: string[] = ['Exchange']
    export const isExchange = (obj?: { __typename?: any } | null): obj is Exchange => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isExchange"')
      return Exchange_possibleTypes.includes(obj.__typename)
    }
    


    const Query_possibleTypes: string[] = ['Query']
    export const isQuery = (obj?: { __typename?: any } | null): obj is Query => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isQuery"')
      return Query_possibleTypes.includes(obj.__typename)
    }
    


    const ValuationPoint_possibleTypes: string[] = ['ValuationPoint']
    export const isValuationPoint = (obj?: { __typename?: any } | null): obj is ValuationPoint => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isValuationPoint"')
      return ValuationPoint_possibleTypes.includes(obj.__typename)
    }
    


    const PortfolioHistory_possibleTypes: string[] = ['PortfolioHistory']
    export const isPortfolioHistory = (obj?: { __typename?: any } | null): obj is PortfolioHistory => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPortfolioHistory"')
      return PortfolioHistory_possibleTypes.includes(obj.__typename)
    }
    


    const TransactionItem_possibleTypes: string[] = ['TransactionItem']
    export const isTransactionItem = (obj?: { __typename?: any } | null): obj is TransactionItem => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTransactionItem"')
      return TransactionItem_possibleTypes.includes(obj.__typename)
    }
    


    const TransactionsConnection_possibleTypes: string[] = ['TransactionsConnection']
    export const isTransactionsConnection = (obj?: { __typename?: any } | null): obj is TransactionsConnection => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTransactionsConnection"')
      return TransactionsConnection_possibleTypes.includes(obj.__typename)
    }
    


    const DeleteTransactionPayload_possibleTypes: string[] = ['DeleteTransactionPayload']
    export const isDeleteTransactionPayload = (obj?: { __typename?: any } | null): obj is DeleteTransactionPayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isDeleteTransactionPayload"')
      return DeleteTransactionPayload_possibleTypes.includes(obj.__typename)
    }
    


    const DeleteInstrumentPayload_possibleTypes: string[] = ['DeleteInstrumentPayload']
    export const isDeleteInstrumentPayload = (obj?: { __typename?: any } | null): obj is DeleteInstrumentPayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isDeleteInstrumentPayload"')
      return DeleteInstrumentPayload_possibleTypes.includes(obj.__typename)
    }
    


    const Mutation_possibleTypes: string[] = ['Mutation']
    export const isMutation = (obj?: { __typename?: any } | null): obj is Mutation => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMutation"')
      return Mutation_possibleTypes.includes(obj.__typename)
    }
    


    const BatchImportTransactionsPayload_possibleTypes: string[] = ['BatchImportTransactionsPayload']
    export const isBatchImportTransactionsPayload = (obj?: { __typename?: any } | null): obj is BatchImportTransactionsPayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isBatchImportTransactionsPayload"')
      return BatchImportTransactionsPayload_possibleTypes.includes(obj.__typename)
    }
    


    const InstrumentPrice_possibleTypes: string[] = ['InstrumentPrice']
    export const isInstrumentPrice = (obj?: { __typename?: any } | null): obj is InstrumentPrice => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isInstrumentPrice"')
      return InstrumentPrice_possibleTypes.includes(obj.__typename)
    }
    


    const InstrumentPricesConnection_possibleTypes: string[] = ['InstrumentPricesConnection']
    export const isInstrumentPricesConnection = (obj?: { __typename?: any } | null): obj is InstrumentPricesConnection => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isInstrumentPricesConnection"')
      return InstrumentPricesConnection_possibleTypes.includes(obj.__typename)
    }
    


    const RecordPriceOverridePayload_possibleTypes: string[] = ['RecordPriceOverridePayload']
    export const isRecordPriceOverridePayload = (obj?: { __typename?: any } | null): obj is RecordPriceOverridePayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRecordPriceOverridePayload"')
      return RecordPriceOverridePayload_possibleTypes.includes(obj.__typename)
    }
    


    const FeedHealthStatus_possibleTypes: string[] = ['FeedHealthStatus']
    export const isFeedHealthStatus = (obj?: { __typename?: any } | null): obj is FeedHealthStatus => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFeedHealthStatus"')
      return FeedHealthStatus_possibleTypes.includes(obj.__typename)
    }
    


    const IngestionStatus_possibleTypes: string[] = ['IngestionStatus']
    export const isIngestionStatus = (obj?: { __typename?: any } | null): obj is IngestionStatus => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isIngestionStatus"')
      return IngestionStatus_possibleTypes.includes(obj.__typename)
    }
    


    const MarketSyncPayload_possibleTypes: string[] = ['MarketSyncPayload']
    export const isMarketSyncPayload = (obj?: { __typename?: any } | null): obj is MarketSyncPayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMarketSyncPayload"')
      return MarketSyncPayload_possibleTypes.includes(obj.__typename)
    }
    


    const BackfillPayload_possibleTypes: string[] = ['BackfillPayload']
    export const isBackfillPayload = (obj?: { __typename?: any } | null): obj is BackfillPayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isBackfillPayload"')
      return BackfillPayload_possibleTypes.includes(obj.__typename)
    }
    


    const Portfolio_possibleTypes: string[] = ['Portfolio']
    export const isPortfolio = (obj?: { __typename?: any } | null): obj is Portfolio => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPortfolio"')
      return Portfolio_possibleTypes.includes(obj.__typename)
    }
    


    const Investment_possibleTypes: string[] = ['Investment']
    export const isInvestment = (obj?: { __typename?: any } | null): obj is Investment => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isInvestment"')
      return Investment_possibleTypes.includes(obj.__typename)
    }
    


    const UserPreferences_possibleTypes: string[] = ['UserPreferences']
    export const isUserPreferences = (obj?: { __typename?: any } | null): obj is UserPreferences => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isUserPreferences"')
      return UserPreferences_possibleTypes.includes(obj.__typename)
    }
    


    const Currency_possibleTypes: string[] = ['Currency']
    export const isCurrency = (obj?: { __typename?: any } | null): obj is Currency => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCurrency"')
      return Currency_possibleTypes.includes(obj.__typename)
    }
    


    const UpdateUserPreferencesPayload_possibleTypes: string[] = ['UpdateUserPreferencesPayload']
    export const isUpdateUserPreferencesPayload = (obj?: { __typename?: any } | null): obj is UpdateUserPreferencesPayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isUpdateUserPreferencesPayload"')
      return UpdateUserPreferencesPayload_possibleTypes.includes(obj.__typename)
    }
    


    const CurrencyPair_possibleTypes: string[] = ['CurrencyPair']
    export const isCurrencyPair = (obj?: { __typename?: any } | null): obj is CurrencyPair => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCurrencyPair"')
      return CurrencyPair_possibleTypes.includes(obj.__typename)
    }
    


    const FXRate_possibleTypes: string[] = ['FXRate']
    export const isFXRate = (obj?: { __typename?: any } | null): obj is FXRate => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFXRate"')
      return FXRate_possibleTypes.includes(obj.__typename)
    }
    


    const FXRatesConnection_possibleTypes: string[] = ['FXRatesConnection']
    export const isFXRatesConnection = (obj?: { __typename?: any } | null): obj is FXRatesConnection => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFXRatesConnection"')
      return FXRatesConnection_possibleTypes.includes(obj.__typename)
    }
    


    const FXHistoryPoint_possibleTypes: string[] = ['FXHistoryPoint']
    export const isFXHistoryPoint = (obj?: { __typename?: any } | null): obj is FXHistoryPoint => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFXHistoryPoint"')
      return FXHistoryPoint_possibleTypes.includes(obj.__typename)
    }
    


    const CurrencyPairHistory_possibleTypes: string[] = ['CurrencyPairHistory']
    export const isCurrencyPairHistory = (obj?: { __typename?: any } | null): obj is CurrencyPairHistory => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCurrencyPairHistory"')
      return CurrencyPairHistory_possibleTypes.includes(obj.__typename)
    }
    


    const RecordFXRateOverridePayload_possibleTypes: string[] = ['RecordFXRateOverridePayload']
    export const isRecordFXRateOverridePayload = (obj?: { __typename?: any } | null): obj is RecordFXRateOverridePayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRecordFXRateOverridePayload"')
      return RecordFXRateOverridePayload_possibleTypes.includes(obj.__typename)
    }
    


    const CashFlowSummary_possibleTypes: string[] = ['CashFlowSummary']
    export const isCashFlowSummary = (obj?: { __typename?: any } | null): obj is CashFlowSummary => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCashFlowSummary"')
      return CashFlowSummary_possibleTypes.includes(obj.__typename)
    }
    


    const CashFlowCategoryBreakdown_possibleTypes: string[] = ['CashFlowCategoryBreakdown']
    export const isCashFlowCategoryBreakdown = (obj?: { __typename?: any } | null): obj is CashFlowCategoryBreakdown => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCashFlowCategoryBreakdown"')
      return CashFlowCategoryBreakdown_possibleTypes.includes(obj.__typename)
    }
    


    const CashFlowItem_possibleTypes: string[] = ['CashFlowItem']
    export const isCashFlowItem = (obj?: { __typename?: any } | null): obj is CashFlowItem => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCashFlowItem"')
      return CashFlowItem_possibleTypes.includes(obj.__typename)
    }
    


    const CashFlowReport_possibleTypes: string[] = ['CashFlowReport']
    export const isCashFlowReport = (obj?: { __typename?: any } | null): obj is CashFlowReport => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCashFlowReport"')
      return CashFlowReport_possibleTypes.includes(obj.__typename)
    }
    

export const enumTransactionType = {
   BUY: 'BUY' as const,
   SELL: 'SELL' as const,
   DIVIDEND: 'DIVIDEND' as const,
   DEPOSIT: 'DEPOSIT' as const,
   WITHDRAWAL: 'WITHDRAWAL' as const,
   INTEREST: 'INTEREST' as const,
   FEE: 'FEE' as const,
   TAX: 'TAX' as const,
   TRANSFER_IN: 'TRANSFER_IN' as const,
   TRANSFER_OUT: 'TRANSFER_OUT' as const,
   FX_CONVERSION: 'FX_CONVERSION' as const,
   SPLIT: 'SPLIT' as const
}

export const enumHistoryTimeframe = {
   TIMEFRAME_1D: 'TIMEFRAME_1D' as const,
   TIMEFRAME_1W: 'TIMEFRAME_1W' as const,
   TIMEFRAME_1M: 'TIMEFRAME_1M' as const,
   TIMEFRAME_1Y: 'TIMEFRAME_1Y' as const,
   TIMEFRAME_ALL: 'TIMEFRAME_ALL' as const
}

export const enumCashFlowTimeframe = {
   MTD: 'MTD' as const,
   YTD: 'YTD' as const,
   M1: 'M1' as const,
   M3: 'M3' as const,
   M6: 'M6' as const,
   Y1: 'Y1' as const,
   ALL: 'ALL' as const,
   CUSTOM: 'CUSTOM' as const
}

export const enumCashFlowDirection = {
   INFLOW: 'INFLOW' as const,
   OUTFLOW: 'OUTFLOW' as const
}

export const enumCashFlowCategory = {
   CAPITAL_DEPOSITS: 'CAPITAL_DEPOSITS' as const,
   DIVIDENDS: 'DIVIDENDS' as const,
   INTEREST: 'INTEREST' as const,
   SALE_PROCEEDS: 'SALE_PROCEEDS' as const,
   CAPITAL_WITHDRAWALS: 'CAPITAL_WITHDRAWALS' as const,
   PURCHASES: 'PURCHASES' as const,
   FEES: 'FEES' as const,
   TAXES: 'TAXES' as const
}
