// @ts-nocheck
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type Scalars = {
    Decimal: any,
    String: string,
    ID: string,
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
    __typename: 'Query'
}

export interface Mutation {
    addTransaction: AddTransactionPayload
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
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MutationGenqlSelection{
    addTransaction?: (AddTransactionPayloadGenqlSelection & { __args: {input: AddTransactionInput} })
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
