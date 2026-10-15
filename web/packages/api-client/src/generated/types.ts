export default {
    "scalars": [
        0,
        2,
        3,
        8,
        10,
        13,
        14,
        54,
        55,
        56,
        68
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
                41
            ],
            "__typename": [
                2
            ]
        },
        "UpdateTransactionInput": {
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
        "UpdateTransactionPayload": {
            "success": [
                8
            ],
            "transaction": [
                17
            ],
            "portfolio": [
                41
            ],
            "__typename": [
                2
            ]
        },
        "Boolean": {},
        "Instrument": {
            "id": [
                10
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
                41,
                {
                    "includeClosed": [
                        8
                    ]
                }
            ],
            "instruments": [
                9
            ],
            "portfolioHistory": [
                16,
                {
                    "timeframe": [
                        14,
                        "HistoryTimeframe!"
                    ]
                }
            ],
            "transactions": [
                18,
                {
                    "type": [
                        3
                    ],
                    "symbol": [
                        2
                    ],
                    "page": [
                        13
                    ],
                    "pageSize": [
                        13
                    ]
                }
            ],
            "allInstruments": [
                9,
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
                11
            ],
            "instrumentPrices": [
                31,
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
                        13
                    ],
                    "offset": [
                        13
                    ]
                }
            ],
            "ingestionStatus": [
                37
            ],
            "me": [
                43
            ],
            "userPreferences": [
                43
            ],
            "supportedCurrencies": [
                44
            ],
            "currencyPairs": [
                47
            ],
            "currencyPairHistory": [
                51,
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
                        14,
                        "HistoryTimeframe!"
                    ]
                }
            ],
            "fxRates": [
                49,
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
                        13
                    ],
                    "offset": [
                        13
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
                61,
                {
                    "filter": [
                        57,
                        "CashFlowFilterInput!"
                    ]
                }
            ],
            "missingPriceTransactions": [
                63,
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
                        13
                    ],
                    "offset": [
                        13
                    ]
                }
            ],
            "missingFXRateTransactions": [
                65,
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
                        13
                    ],
                    "offset": [
                        13
                    ]
                }
            ],
            "modifiedDietzReturn": [
                66,
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
                69
            ],
            "benchmarkComparison": [
                72,
                {
                    "benchmarkId": [
                        10,
                        "ID!"
                    ],
                    "timeframe": [
                        14,
                        "HistoryTimeframe!"
                    ],
                    "currencyMode": [
                        68
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
                15
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
                10
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
                17
            ],
            "totalCount": [
                13
            ],
            "page": [
                13
            ],
            "pageSize": [
                13
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
                41
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
                10
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
                19,
                {
                    "id": [
                        10,
                        "ID!"
                    ]
                }
            ],
            "updateTransaction": [
                7,
                {
                    "id": [
                        10,
                        "ID!"
                    ],
                    "input": [
                        6,
                        "UpdateTransactionInput!"
                    ]
                }
            ],
            "createInstrument": [
                9,
                {
                    "input": [
                        32,
                        "CreateInstrumentInput!"
                    ]
                }
            ],
            "updateInstrument": [
                9,
                {
                    "input": [
                        33,
                        "UpdateInstrumentInput!"
                    ]
                }
            ],
            "deleteInstrument": [
                20,
                {
                    "id": [
                        10,
                        "ID!"
                    ]
                }
            ],
            "recordPriceOverride": [
                35,
                {
                    "input": [
                        34,
                        "RecordPriceOverrideInput!"
                    ]
                }
            ],
            "triggerMarketSync": [
                38,
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
                40,
                {
                    "input": [
                        39,
                        "TriggerBackfillInput!"
                    ]
                }
            ],
            "updateUserPreferences": [
                46,
                {
                    "input": [
                        45,
                        "UpdateUserPreferencesInput!"
                    ]
                }
            ],
            "recordFXRateOverride": [
                53,
                {
                    "input": [
                        52,
                        "RecordFXRateOverrideInput!"
                    ]
                }
            ],
            "importTransactions": [
                29,
                {
                    "input": [
                        28,
                        "BatchImportTransactionsInput!"
                    ]
                }
            ],
            "register": [
                22,
                {
                    "input": [
                        25,
                        "RegisterInput!"
                    ]
                }
            ],
            "login": [
                22,
                {
                    "input": [
                        26,
                        "LoginInput!"
                    ]
                }
            ],
            "refreshToken": [
                22
            ],
            "logout": [
                8
            ],
            "changePassword": [
                23,
                {
                    "input": [
                        24,
                        "ChangePasswordInput!"
                    ]
                }
            ],
            "__typename": [
                2
            ]
        },
        "AuthPayload": {
            "user": [
                43
            ],
            "accessToken": [
                2
            ],
            "expiresIn": [
                13
            ],
            "__typename": [
                2
            ]
        },
        "ChangePasswordPayload": {
            "success": [
                8
            ],
            "message": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "ChangePasswordInput": {
            "currentPassword": [
                2
            ],
            "newPassword": [
                2
            ],
            "revokeOtherSessions": [
                8
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
                27
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
                13
            ],
            "skippedCount": [
                13
            ],
            "portfolio": [
                41
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
                10
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
                30
            ],
            "totalCount": [
                13
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
                10
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
                30
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
                36
            ],
            "trackedInstruments": [
                13
            ],
            "trackedCurrencies": [
                13
            ],
            "latestPriceDate": [
                2
            ],
            "latestFxDate": [
                2
            ],
            "rateLimitRemaining": [
                13
            ],
            "rateLimitBudget": [
                13
            ],
            "pendingBackfillJobs": [
                13
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
                13
            ],
            "fxRatesSynced": [
                13
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
                13
            ],
            "fxRatesSynced": [
                13
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
                42
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
                10
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
                10
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
                43
            ],
            "portfolio": [
                41
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
                13
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
                48
            ],
            "totalCount": [
                13
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
                50
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
                48
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
                54
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
                10
            ],
            "eventDate": [
                2
            ],
            "type": [
                3
            ],
            "flowDirection": [
                55
            ],
            "category": [
                56
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
                58
            ],
            "breakdown": [
                59
            ],
            "items": [
                60
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
                10
            ],
            "portfolioId": [
                10
            ],
            "instrumentId": [
                10
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
                62
            ],
            "totalCount": [
                13
            ],
            "uniqueInstrumentsCount": [
                13
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
                10
            ],
            "portfolioId": [
                10
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
                10
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
                64
            ],
            "totalCount": [
                13
            ],
            "uniquePairsCount": [
                13
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
                10
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
                13
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
                67
            ],
            "__typename": [
                2
            ]
        },
        "ExternalCashFlowWeightItem": {
            "transactionId": [
                10
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
                10
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
                10
            ],
            "baseCurrency": [
                2
            ],
            "benchmark": [
                69
            ],
            "currencyMode": [
                68
            ],
            "startDate": [
                2
            ],
            "endDate": [
                2
            ],
            "elapsedDays": [
                13
            ],
            "points": [
                70
            ],
            "metrics": [
                71
            ],
            "__typename": [
                2
            ]
        }
    }
}