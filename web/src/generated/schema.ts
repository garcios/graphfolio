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

export interface Query {
    portfolio: Portfolio
    __typename: 'Query'
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

export interface QueryGenqlSelection{
    portfolio?: PortfolioGenqlSelection
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
    


    const Query_possibleTypes: string[] = ['Query']
    export const isQuery = (obj?: { __typename?: any } | null): obj is Query => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isQuery"')
      return Query_possibleTypes.includes(obj.__typename)
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
    