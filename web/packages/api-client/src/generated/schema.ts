// @ts-nocheck
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type Scalars = {
    Decimal: any,
    String: string,
    ID: string,
    Int: number,
    Boolean: boolean,
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
    __typename: 'Instrument'
}

export interface Query {
    portfolio: Portfolio
    instruments: Instrument[]
    portfolioHistory: PortfolioHistory
    transactions: TransactionsConnection
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
    __typename: 'Mutation'
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
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface QueryGenqlSelection{
    portfolio?: PortfolioGenqlSelection
    instruments?: InstrumentGenqlSelection
    portfolioHistory?: (PortfolioHistoryGenqlSelection & { __args: {timeframe: HistoryTimeframe} })
    transactions?: (TransactionsConnectionGenqlSelection & { __args?: {type?: (TransactionType | null), symbol?: (Scalars['String'] | null), page?: (Scalars['Int'] | null), pageSize?: (Scalars['Int'] | null)} })
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
