export default {
    "scalars": [
        0,
        2,
        3,
        7,
        8,
        11,
        12,
        50,
        51,
        52,
        64
    ],
    "types": {
        "Decimal": {},
        "Money": {
            "amount": [
                0
            ],
            "currencyCode": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "String": {},
        "TransactionType": {},
        "AddTransactionInput": {
            "type": [
                3
            ],
            "symbol": [
                2
            ],
            "tradeDate": [
                2
            ],
            "quantity": [
                0
            ],
            "price": [
                0
            ],
            "amount": [
                0
            ],
            "currencyCode": [
                2
            ],
            "fee": [
                0
            ],
            "feeCurrencyCode": [
                2
            ],
            "notes": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "AddTransactionPayload": {
            "transactionId": [
                2
            ],
            "portfolio": [
                37
            ],
            "__typename": [
                2
            ]
        },
        "Instrument": {
            "id": [
                7
            ],
            "symbol": [
                2
            ],
            "name": [
                2
            ],
            "currencyCode": [
                2
            ],
            "assetClass": [
                2
            ],
            "exchangeCode": [
                2
            ],
            "isin": [
                2
            ],
            "isActive": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "ID": {},
        "Boolean": {},
        "Exchange": {
            "code": [
                2
            ],
            "name": [
                2
            ],
            "country": [
                2
            ],
            "timezone": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "Query": {
            "portfolio": [
                37,
                {
                    "includeClosed": [
                        8
                    ]
                }
            ],
            "instruments": [
                6
            ],
            "portfolioHistory": [
                14,
                {
                    "timeframe": [
                        12,
                        "HistoryTimeframe!"
                    ]
                }
            ],
            "transactions": [
                16,
                {
                    "type": [
                        3
                    ],
                    "symbol": [
                        2
                    ],
                    "page": [
                        11
                    ],
                    "pageSize": [
                        11
                    ]
                }
            ],
            "allInstruments": [
                6,
                {
                    "isActive": [
                        8
                    ],
                    "search": [
                        2
                    ]
                }
            ],
            "exchanges": [
                9
            ],
            "instrumentPrices": [
                27,
                {
                    "symbol": [
                        2
                    ],
                    "fromDate": [
                        2
                    ],
                    "toDate": [
                        2
                    ],
                    "limit": [
                        11
                    ],
                    "offset": [
                        11
                    ]
                }
            ],
            "ingestionStatus": [
                33
            ],
            "me": [
                39
            ],
            "userPreferences": [
                39
            ],
            "supportedCurrencies": [
                40
            ],
            "currencyPairs": [
                43
            ],
            "currencyPairHistory": [
                47,
                {
                    "baseCurrency": [
                        2,
                        "String!"
                    ],
                    "quoteCurrency": [
                        2,
                        "String!"
                    ],
                    "timeframe": [
                        12,
                        "HistoryTimeframe!"
                    ]
                }
            ],
            "fxRates": [
                45,
                {
                    "baseCurrency": [
                        2
                    ],
                    "quoteCurrency": [
                        2
                    ],
                    "fromDate": [
                        2
                    ],
                    "toDate": [
                        2
                    ],
                    "limit": [
                        11
                    ],
                    "offset": [
                        11
                    ]
                }
            ],
            "checkTransactionDuplicates": [
                2,
                {
                    "externalRefs": [
                        2,
                        "[String!]!"
                    ]
                }
            ],
            "cashFlowReport": [
                57,
                {
                    "filter": [
                        53,
                        "CashFlowFilterInput!"
                    ]
                }
            ],
            "missingPriceTransactions": [
                59,
                {
                    "symbol": [
                        2
                    ],
                    "fromDate": [
                        2
                    ],
                    "toDate": [
                        2
                    ],
                    "limit": [
                        11
                    ],
                    "offset": [
                        11
                    ]
                }
            ],
            "missingFXRateTransactions": [
                61,
                {
                    "currencyPair": [
                        2
                    ],
                    "fromDate": [
                        2
                    ],
                    "toDate": [
                        2
                    ],
                    "type": [
                        3
                    ],
                    "limit": [
                        11
                    ],
                    "offset": [
                        11
                    ]
                }
            ],
            "modifiedDietzReturn": [
                62,
                {
                    "startDate": [
                        2,
                        "String!"
                    ],
                    "endDate": [
                        2,
                        "String!"
                    ],
                    "useMonthlyLinking": [
                        8
                    ]
                }
            ],
            "benchmarks": [
                65
            ],
            "benchmarkComparison": [
                68,
                {
                    "benchmarkId": [
                        7,
                        "ID!"
                    ],
                    "timeframe": [
                        12,
                        "HistoryTimeframe!"
                    ],
                    "currencyMode": [
                        64
                    ]
                }
            ],
            "__typename": [
                2
            ]
        },
        "Int": {},
        "HistoryTimeframe": {},
        "ValuationPoint": {
            "date": [
                2
            ],
            "totalValue": [
                1
            ],
            "marketValue": [
                1
            ],
            "cashValue": [
                1
            ],
            "twrIndex": [
                0
            ],
            "dailyReturn": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "PortfolioHistory": {
            "points": [
                13
            ],
            "startValue": [
                1
            ],
            "endValue": [
                1
            ],
            "returnAmount": [
                1
            ],
            "returnPercent": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "TransactionItem": {
            "id": [
                7
            ],
            "type": [
                3
            ],
            "symbol": [
                2
            ],
            "instrumentName": [
                2
            ],
            "tradeDate": [
                2
            ],
            "quantity": [
                0
            ],
            "price": [
                1
            ],
            "amount": [
                1
            ],
            "fee": [
                1
            ],
            "notes": [
                2
            ],
            "createdAt": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "TransactionsConnection": {
            "items": [
                15
            ],
            "totalCount": [
                11
            ],
            "page": [
                11
            ],
            "pageSize": [
                11
            ],
            "__typename": [
                2
            ]
        },
        "DeleteTransactionPayload": {
            "success": [
                8
            ],
            "portfolio": [
                37
            ],
            "__typename": [
                2
            ]
        },
        "DeleteInstrumentPayload": {
            "success": [
                8
            ],
            "id": [
                7
            ],
            "__typename": [
                2
            ]
        },
        "Mutation": {
            "addTransaction": [
                5,
                {
                    "input": [
                        4,
                        "AddTransactionInput!"
                    ]
                }
            ],
            "deleteTransaction": [
                17,
                {
                    "id": [
                        7,
                        "ID!"
                    ]
                }
            ],
            "createInstrument": [
                6,
                {
                    "input": [
                        28,
                        "CreateInstrumentInput!"
                    ]
                }
            ],
            "updateInstrument": [
                6,
                {
                    "input": [
                        29,
                        "UpdateInstrumentInput!"
                    ]
                }
            ],
            "deleteInstrument": [
                18,
                {
                    "id": [
                        7,
                        "ID!"
                    ]
                }
            ],
            "recordPriceOverride": [
                31,
                {
                    "input": [
                        30,
                        "RecordPriceOverrideInput!"
                    ]
                }
            ],
            "triggerMarketSync": [
                34,
                {
                    "symbols": [
                        2,
                        "[String!]"
                    ],
                    "syncFx": [
                        8
                    ]
                }
            ],
            "triggerBackfill": [
                36,
                {
                    "input": [
                        35,
                        "TriggerBackfillInput!"
                    ]
                }
            ],
            "updateUserPreferences": [
                42,
                {
                    "input": [
                        41,
                        "UpdateUserPreferencesInput!"
                    ]
                }
            ],
            "recordFXRateOverride": [
                49,
                {
                    "input": [
                        48,
                        "RecordFXRateOverrideInput!"
                    ]
                }
            ],
            "importTransactions": [
                25,
                {
                    "input": [
                        24,
                        "BatchImportTransactionsInput!"
                    ]
                }
            ],
            "register": [
                20,
                {
                    "input": [
                        21,
                        "RegisterInput!"
                    ]
                }
            ],
            "login": [
                20,
                {
                    "input": [
                        22,
                        "LoginInput!"
                    ]
                }
            ],
            "refreshToken": [
                20
            ],
            "logout": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "AuthPayload": {
            "user": [
                39
            ],
            "accessToken": [
                2
            ],
            "expiresIn": [
                11
            ],
            "__typename": [
                2
            ]
        },
        "RegisterInput": {
            "email": [
                2
            ],
            "password": [
                2
            ],
            "displayName": [
                2
            ],
            "displayCurrency": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "LoginInput": {
            "email": [
                2
            ],
            "password": [
                2
            ],
            "rememberMe": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "ImportTransactionInput": {
            "externalRef": [
                2
            ],
            "symbol": [
                2
            ],
            "type": [
                3
            ],
            "tradeDate": [
                2
            ],
            "settleDate": [
                2
            ],
            "quantity": [
                0
            ],
            "price": [
                0
            ],
            "amount": [
                0
            ],
            "fee": [
                0
            ],
            "feeCurrencyCode": [
                2
            ],
            "currencyCode": [
                2
            ],
            "notes": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "BatchImportTransactionsInput": {
            "transactions": [
                23
            ],
            "skipDuplicates": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "BatchImportTransactionsPayload": {
            "success": [
                8
            ],
            "importedCount": [
                11
            ],
            "skippedCount": [
                11
            ],
            "portfolio": [
                37
            ],
            "message": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "InstrumentPrice": {
            "id": [
                7
            ],
            "symbol": [
                2
            ],
            "priceDate": [
                2
            ],
            "price": [
                1
            ],
            "source": [
                2
            ],
            "updatedAt": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "InstrumentPricesConnection": {
            "items": [
                26
            ],
            "totalCount": [
                11
            ],
            "__typename": [
                2
            ]
        },
        "CreateInstrumentInput": {
            "symbol": [
                2
            ],
            "exchangeCode": [
                2
            ],
            "name": [
                2
            ],
            "assetClass": [
                2
            ],
            "currencyCode": [
                2
            ],
            "isin": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "UpdateInstrumentInput": {
            "id": [
                7
            ],
            "name": [
                2
            ],
            "isActive": [
                8
            ],
            "isin": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "RecordPriceOverrideInput": {
            "symbol": [
                2
            ],
            "priceDate": [
                2
            ],
            "price": [
                0
            ],
            "reason": [
                2
            ],
            "recomputeValuations": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "RecordPriceOverridePayload": {
            "price": [
                26
            ],
            "valuationsRecomputed": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "FeedHealthStatus": {
            "name": [
                2
            ],
            "status": [
                2
            ],
            "provider": [
                2
            ],
            "schedule": [
                2
            ],
            "lastRun": [
                2
            ],
            "details": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "IngestionStatus": {
            "feeds": [
                32
            ],
            "trackedInstruments": [
                11
            ],
            "trackedCurrencies": [
                11
            ],
            "latestPriceDate": [
                2
            ],
            "latestFxDate": [
                2
            ],
            "rateLimitRemaining": [
                11
            ],
            "rateLimitBudget": [
                11
            ],
            "pendingBackfillJobs": [
                11
            ],
            "__typename": [
                2
            ]
        },
        "MarketSyncPayload": {
            "success": [
                8
            ],
            "pricesSynced": [
                11
            ],
            "fxRatesSynced": [
                11
            ],
            "message": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "TriggerBackfillInput": {
            "fromDate": [
                2
            ],
            "toDate": [
                2
            ],
            "symbols": [
                2
            ],
            "currencyPairs": [
                2
            ],
            "backfillAssets": [
                8
            ],
            "backfillFx": [
                8
            ],
            "recomputeValuations": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "BackfillPayload": {
            "success": [
                8
            ],
            "pricesSynced": [
                11
            ],
            "fxRatesSynced": [
                11
            ],
            "message": [
                2
            ],
            "warnings": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "Portfolio": {
            "totalValue": [
                1
            ],
            "todayReturnAmount": [
                1
            ],
            "todayReturnPercent": [
                0
            ],
            "annualizedReturnPercent": [
                0
            ],
            "cashBalance": [
                1
            ],
            "investments": [
                38
            ],
            "inceptionDate": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "Investment": {
            "id": [
                7
            ],
            "ticker": [
                2
            ],
            "name": [
                2
            ],
            "price": [
                1
            ],
            "averageBuyPrice": [
                1
            ],
            "quantity": [
                0
            ],
            "totalValue": [
                1
            ],
            "todayReturnAmount": [
                1
            ],
            "todayReturnPercent": [
                0
            ],
            "totalReturnAmount": [
                1
            ],
            "totalReturnPercent": [
                0
            ],
            "capitalGainAmount": [
                1
            ],
            "capitalGainPercent": [
                0
            ],
            "incomeAmount": [
                1
            ],
            "incomeYieldPercent": [
                0
            ],
            "currencyGainAmount": [
                1
            ],
            "currencyGainPercent": [
                0
            ],
            "isInternational": [
                8
            ],
            "isClosed": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "UserPreferences": {
            "userId": [
                7
            ],
            "email": [
                2
            ],
            "displayName": [
                2
            ],
            "displayCurrency": [
                2
            ],
            "theme": [
                2
            ],
            "createdAt": [
                2
            ],
            "updatedAt": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "Currency": {
            "code": [
                2
            ],
            "name": [
                2
            ],
            "symbol": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "UpdateUserPreferencesInput": {
            "displayName": [
                2
            ],
            "displayCurrency": [
                2
            ],
            "theme": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "UpdateUserPreferencesPayload": {
            "preferences": [
                39
            ],
            "portfolio": [
                37
            ],
            "__typename": [
                2
            ]
        },
        "CurrencyPair": {
            "baseCurrency": [
                2
            ],
            "quoteCurrency": [
                2
            ],
            "pair": [
                2
            ],
            "latestRate": [
                0
            ],
            "latestDate": [
                2
            ],
            "latestSource": [
                2
            ],
            "previousRate": [
                0
            ],
            "change1dAmount": [
                0
            ],
            "change1dPct": [
                0
            ],
            "totalRecords": [
                11
            ],
            "firstDate": [
                2
            ],
            "lastDate": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "FXRate": {
            "baseCurrency": [
                2
            ],
            "quoteCurrency": [
                2
            ],
            "pair": [
                2
            ],
            "rateDate": [
                2
            ],
            "rate": [
                0
            ],
            "invertedRate": [
                0
            ],
            "source": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "FXRatesConnection": {
            "items": [
                44
            ],
            "totalCount": [
                11
            ],
            "__typename": [
                2
            ]
        },
        "FXHistoryPoint": {
            "date": [
                2
            ],
            "rate": [
                0
            ],
            "invertedRate": [
                0
            ],
            "source": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "CurrencyPairHistory": {
            "baseCurrency": [
                2
            ],
            "quoteCurrency": [
                2
            ],
            "pair": [
                2
            ],
            "points": [
                46
            ],
            "startRate": [
                0
            ],
            "endRate": [
                0
            ],
            "periodChange": [
                0
            ],
            "periodChangePct": [
                0
            ],
            "periodHigh": [
                0
            ],
            "periodLow": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "RecordFXRateOverrideInput": {
            "baseCurrency": [
                2
            ],
            "quoteCurrency": [
                2
            ],
            "rateDate": [
                2
            ],
            "rate": [
                0
            ],
            "reason": [
                2
            ],
            "recomputeValuations": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "RecordFXRateOverridePayload": {
            "rate": [
                44
            ],
            "valuationsRecomputed": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "CashFlowTimeframe": {},
        "CashFlowDirection": {},
        "CashFlowCategory": {},
        "CashFlowFilterInput": {
            "timeframe": [
                50
            ],
            "fromDate": [
                2
            ],
            "toDate": [
                2
            ],
            "currency": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "CashFlowSummary": {
            "startingCashBalance": [
                1
            ],
            "totalInflows": [
                1
            ],
            "totalOutflows": [
                1
            ],
            "netCashFlow": [
                1
            ],
            "endingCashBalance": [
                1
            ],
            "__typename": [
                2
            ]
        },
        "CashFlowCategoryBreakdown": {
            "deposits": [
                1
            ],
            "dividends": [
                1
            ],
            "interest": [
                1
            ],
            "salesProceeds": [
                1
            ],
            "withdrawals": [
                1
            ],
            "purchases": [
                1
            ],
            "fees": [
                1
            ],
            "taxes": [
                1
            ],
            "__typename": [
                2
            ]
        },
        "CashFlowItem": {
            "id": [
                7
            ],
            "eventDate": [
                2
            ],
            "type": [
                3
            ],
            "flowDirection": [
                51
            ],
            "category": [
                52
            ],
            "symbol": [
                2
            ],
            "instrumentName": [
                2
            ],
            "description": [
                2
            ],
            "netAmount": [
                1
            ],
            "runningBalance": [
                1
            ],
            "localAmount": [
                1
            ],
            "fee": [
                1
            ],
            "withholdingTax": [
                1
            ],
            "__typename": [
                2
            ]
        },
        "CashFlowReport": {
            "summary": [
                54
            ],
            "breakdown": [
                55
            ],
            "items": [
                56
            ],
            "baseCurrency": [
                2
            ],
            "fromDate": [
                2
            ],
            "toDate": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "MissingPriceTransactionItem": {
            "transactionId": [
                7
            ],
            "portfolioId": [
                7
            ],
            "instrumentId": [
                7
            ],
            "symbol": [
                2
            ],
            "assetName": [
                2
            ],
            "exchangeMic": [
                2
            ],
            "tradeDate": [
                2
            ],
            "type": [
                3
            ],
            "quantity": [
                0
            ],
            "executionPrice": [
                1
            ],
            "currencyCode": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "MissingPriceReport": {
            "items": [
                58
            ],
            "totalCount": [
                11
            ],
            "uniqueInstrumentsCount": [
                11
            ],
            "earliestTradeDate": [
                2
            ],
            "latestTradeDate": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "MissingFXRateTransactionItem": {
            "transactionId": [
                7
            ],
            "portfolioId": [
                7
            ],
            "portfolioName": [
                2
            ],
            "portfolioBaseCurrency": [
                2
            ],
            "tradeDate": [
                2
            ],
            "type": [
                3
            ],
            "instrumentId": [
                7
            ],
            "symbol": [
                2
            ],
            "assetName": [
                2
            ],
            "amount": [
                1
            ],
            "transactionCurrency": [
                2
            ],
            "requiredPair": [
                2
            ],
            "recordedFxRate": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "MissingFXRatesReport": {
            "items": [
                60
            ],
            "totalCount": [
                11
            ],
            "uniquePairsCount": [
                11
            ],
            "earliestTradeDate": [
                2
            ],
            "latestTradeDate": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "ModifiedDietzReturnPayload": {
            "portfolioId": [
                7
            ],
            "baseCurrency": [
                2
            ],
            "startDate": [
                2
            ],
            "endDate": [
                2
            ],
            "elapsedDays": [
                11
            ],
            "startValuation": [
                1
            ],
            "endValuation": [
                1
            ],
            "netExternalFlows": [
                1
            ],
            "netInvestmentGain": [
                1
            ],
            "weightedCapitalEmployed": [
                1
            ],
            "cumulativeReturnPercent": [
                0
            ],
            "annualizedReturnPercent": [
                0
            ],
            "isAnnualized": [
                8
            ],
            "isComputable": [
                8
            ],
            "diagnosticCode": [
                2
            ],
            "cashFlowWeights": [
                63
            ],
            "__typename": [
                2
            ]
        },
        "ExternalCashFlowWeightItem": {
            "transactionId": [
                7
            ],
            "date": [
                2
            ],
            "type": [
                3
            ],
            "amount": [
                1
            ],
            "weight": [
                0
            ],
            "weightedAmount": [
                1
            ],
            "localAmount": [
                1
            ],
            "__typename": [
                2
            ]
        },
        "BenchmarkCurrencyMode": {},
        "Benchmark": {
            "id": [
                7
            ],
            "symbol": [
                2
            ],
            "name": [
                2
            ],
            "exchangeCode": [
                2
            ],
            "currencyCode": [
                2
            ],
            "indexType": [
                2
            ],
            "isDefault": [
                8
            ],
            "__typename": [
                2
            ]
        },
        "BenchmarkComparisonPoint": {
            "date": [
                2
            ],
            "portfolioReturnPercent": [
                0
            ],
            "benchmarkReturnPercent": [
                0
            ],
            "activeReturnPercent": [
                0
            ],
            "portfolioValue": [
                1
            ],
            "benchmarkPriceAdjusted": [
                1
            ],
            "__typename": [
                2
            ]
        },
        "BenchmarkMetrics": {
            "portfolioCumulativeReturnPercent": [
                0
            ],
            "benchmarkCumulativeReturnPercent": [
                0
            ],
            "activeCumulativeReturnPercent": [
                0
            ],
            "portfolioAnnualizedReturnPercent": [
                0
            ],
            "benchmarkAnnualizedReturnPercent": [
                0
            ],
            "activeAnnualizedReturnPercent": [
                0
            ],
            "isAnnualized": [
                8
            ],
            "beta": [
                0
            ],
            "trackingErrorPercent": [
                0
            ],
            "informationRatio": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "BenchmarkComparisonPayload": {
            "portfolioId": [
                7
            ],
            "baseCurrency": [
                2
            ],
            "benchmark": [
                65
            ],
            "currencyMode": [
                64
            ],
            "startDate": [
                2
            ],
            "endDate": [
                2
            ],
            "elapsedDays": [
                11
            ],
            "points": [
                66
            ],
            "metrics": [
                67
            ],
            "__typename": [
                2
            ]
        }
    }
}