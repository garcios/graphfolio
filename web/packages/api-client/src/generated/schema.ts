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

export type TransactionType = 'BUY' | 'SELL' | 'DIVIDEND' | 'DEPOSIT' | 'WITHDRAWAL' | 'INTEREST' | 'FEE' | 'TAX' | 'TRANSFER_IN' | 'TRANSFER_OUT' | 'FX_CONVERSION'

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

export interface Query {
    portfolio: Portfolio
    instruments: Instrument[]
    portfolioHistory: PortfolioHistory
    transactions: TransactionsConnection
    allInstruments: Instrument[]
    instrumentPrices: InstrumentPricesConnection
    ingestionStatus: IngestionStatus
    userPreferences: UserPreferences
    supportedCurrencies: Currency[]
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

export interface Mutation {
    addTransaction: AddTransactionPayload
    deleteTransaction: DeleteTransactionPayload
    createInstrument: Instrument
    updateInstrument: Instrument
    recordPriceOverride: RecordPriceOverridePayload
    triggerMarketSync: MarketSyncPayload
    updateUserPreferences: UpdateUserPreferencesPayload
    __typename: 'Mutation'
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
    quantity: Scalars['Decimal']
    totalValue: Money
    todayReturnAmount: Money
    todayReturnPercent: Scalars['Decimal']
    totalReturnAmount: Money
    totalReturnPercent: Scalars['Decimal']
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

export interface MoneyGenqlSelection{
    amount?: boolean | number
    currencyCode?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface AddTransactionInput {type: TransactionType,symbol?: (Scalars['String'] | null),tradeDate: Scalars['String'],quantity?: (Scalars['Decimal'] | null),price?: (Scalars['Decimal'] | null),amount?: (Scalars['Decimal'] | null),currencyCode?: (Scalars['String'] | null),fee?: (Scalars['Decimal'] | null),notes?: (Scalars['String'] | null)}

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

export interface QueryGenqlSelection{
    portfolio?: PortfolioGenqlSelection
    instruments?: InstrumentGenqlSelection
    portfolioHistory?: (PortfolioHistoryGenqlSelection & { __args: {timeframe: HistoryTimeframe} })
    transactions?: (TransactionsConnectionGenqlSelection & { __args?: {type?: (TransactionType | null), symbol?: (Scalars['String'] | null), page?: (Scalars['Int'] | null), pageSize?: (Scalars['Int'] | null)} })
    allInstruments?: (InstrumentGenqlSelection & { __args?: {isActive?: (Scalars['Boolean'] | null), search?: (Scalars['String'] | null)} })
    instrumentPrices?: (InstrumentPricesConnectionGenqlSelection & { __args?: {symbol?: (Scalars['String'] | null), fromDate?: (Scalars['String'] | null), toDate?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    ingestionStatus?: IngestionStatusGenqlSelection
    userPreferences?: UserPreferencesGenqlSelection
    supportedCurrencies?: CurrencyGenqlSelection
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

export interface MutationGenqlSelection{
    addTransaction?: (AddTransactionPayloadGenqlSelection & { __args: {input: AddTransactionInput} })
    deleteTransaction?: (DeleteTransactionPayloadGenqlSelection & { __args: {id: Scalars['ID']} })
    createInstrument?: (InstrumentGenqlSelection & { __args: {input: CreateInstrumentInput} })
    updateInstrument?: (InstrumentGenqlSelection & { __args: {input: UpdateInstrumentInput} })
    recordPriceOverride?: (RecordPriceOverridePayloadGenqlSelection & { __args: {input: RecordPriceOverrideInput} })
    triggerMarketSync?: (MarketSyncPayloadGenqlSelection & { __args?: {symbols?: (Scalars['String'][] | null), syncFx?: (Scalars['Boolean'] | null)} })
    updateUserPreferences?: (UpdateUserPreferencesPayloadGenqlSelection & { __args: {input: UpdateUserPreferencesInput} })
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
    quantity?: boolean | number
    totalValue?: MoneyGenqlSelection
    todayReturnAmount?: MoneyGenqlSelection
    todayReturnPercent?: boolean | number
    totalReturnAmount?: MoneyGenqlSelection
    totalReturnPercent?: boolean | number
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
    


    const Mutation_possibleTypes: string[] = ['Mutation']
    export const isMutation = (obj?: { __typename?: any } | null): obj is Mutation => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMutation"')
      return Mutation_possibleTypes.includes(obj.__typename)
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
   FX_CONVERSION: 'FX_CONVERSION' as const
}

export const enumHistoryTimeframe = {
   TIMEFRAME_1D: 'TIMEFRAME_1D' as const,
   TIMEFRAME_1W: 'TIMEFRAME_1W' as const,
   TIMEFRAME_1M: 'TIMEFRAME_1M' as const,
   TIMEFRAME_1Y: 'TIMEFRAME_1Y' as const,
   TIMEFRAME_ALL: 'TIMEFRAME_ALL' as const
}
