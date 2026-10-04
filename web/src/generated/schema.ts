// @ts-nocheck
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type Scalars = {
    Float: number,
    ID: string,
    String: string,
    Boolean: boolean,
}

export interface Query {
    portfolio: Portfolio
    __typename: 'Query'
}

export interface Portfolio {
    totalValue: Scalars['Float']
    todayReturnAmount: Scalars['Float']
    todayReturnPercent: Scalars['Float']
    annualizedReturnPercent: Scalars['Float']
    cashBalance: Scalars['Float']
    investments: Investment[]
    __typename: 'Portfolio'
}

export interface Investment {
    id: Scalars['ID']
    ticker: Scalars['String']
    name: Scalars['String']
    price: Scalars['Float']
    quantity: Scalars['Float']
    totalValue: Scalars['Float']
    todayReturnAmount: Scalars['Float']
    todayReturnPercent: Scalars['Float']
    totalReturnAmount: Scalars['Float']
    totalReturnPercent: Scalars['Float']
    __typename: 'Investment'
}

export interface QueryGenqlSelection{
    portfolio?: PortfolioGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PortfolioGenqlSelection{
    totalValue?: boolean | number
    todayReturnAmount?: boolean | number
    todayReturnPercent?: boolean | number
    annualizedReturnPercent?: boolean | number
    cashBalance?: boolean | number
    investments?: InvestmentGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface InvestmentGenqlSelection{
    id?: boolean | number
    ticker?: boolean | number
    name?: boolean | number
    price?: boolean | number
    quantity?: boolean | number
    totalValue?: boolean | number
    todayReturnAmount?: boolean | number
    todayReturnPercent?: boolean | number
    totalReturnAmount?: boolean | number
    totalReturnPercent?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
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
    