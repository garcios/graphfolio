export default {
    "scalars": [
        0,
        2,
        3,
        4,
        9,
        11,
        14,
        15,
        46,
        58,
        59,
        60,
        72,
        77,
        78,
        86,
        87
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
        "SortDirection": {},
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
            "frankingCredit": [
                0
            ],
            "frankingPct": [
                0
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
                44
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
            "frankingCredit": [
                0
            ],
            "frankingPct": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "UpdateTransactionPayload": {
            "success": [
                9
            ],
            "transaction": [
                18
            ],
            "portfolio": [
                44
            ],
            "__typename": [
                2
            ]
        },
        "Boolean": {},
        "Instrument": {
            "id": [
                11
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
                9
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
                44,
                {
                    "includeClosed": [
                        9
                    ]
                }
            ],
            "instruments": [
                10
            ],
            "portfolioHistory": [
                17,
                {
                    "timeframe": [
                        15,
                        "HistoryTimeframe!"
                    ]
                }
            ],
            "transactions": [
                19,
                {
                    "type": [
                        3
                    ],
                    "symbol": [
                        2
                    ],
                    "page": [
                        14
                    ],
                    "pageSize": [
                        14
                    ],
                    "sortDirection": [
                        4
                    ]
                }
            ],
            "exportTransactions": [
                18,
                {
                    "type": [
                        3
                    ],
                    "symbol": [
                        2
                    ]
                }
            ],
            "allInstruments": [
                10,
                {
                    "isActive": [
                        9
                    ],
                    "search": [
                        2
                    ]
                }
            ],
            "exchanges": [
                12
            ],
            "instrumentPrices": [
                32,
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
                        14
                    ],
                    "offset": [
                        14
                    ]
                }
            ],
            "ingestionStatus": [
                38
            ],
            "me": [
                47
            ],
            "userPreferences": [
                47
            ],
            "supportedCurrencies": [
                48
            ],
            "currencyPairs": [
                51
            ],
            "currencyPairHistory": [
                55,
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
                        15,
                        "HistoryTimeframe!"
                    ]
                }
            ],
            "fxRates": [
                53,
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
                        14
                    ],
                    "offset": [
                        14
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
                65,
                {
                    "filter": [
                        61,
                        "CashFlowFilterInput!"
                    ]
                }
            ],
            "missingPriceTransactions": [
                67,
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
                        14
                    ],
                    "offset": [
                        14
                    ]
                }
            ],
            "missingFXRateTransactions": [
                69,
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
                        14
                    ],
                    "offset": [
                        14
                    ]
                }
            ],
            "modifiedDietzReturn": [
                70,
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
                        9
                    ]
                }
            ],
            "benchmarks": [
                73
            ],
            "benchmarkComparison": [
                76,
                {
                    "benchmarkId": [
                        11,
                        "ID!"
                    ],
                    "timeframe": [
                        15,
                        "HistoryTimeframe!"
                    ],
                    "currencyMode": [
                        72
                    ]
                }
            ],
            "dividendCalendar": [
                79,
                {
                    "year": [
                        14,
                        "Int!"
                    ],
                    "month": [
                        14,
                        "Int!"
                    ]
                }
            ],
            "dividendYieldAnalytics": [
                82
            ],
            "simulateDRIP": [
                85,
                {
                    "input": [
                        83,
                        "DRIPSimulationInput!"
                    ]
                }
            ],
            "taxReport": [
                94,
                {
                    "financialYear": [
                        14
                    ],
                    "fromDate": [
                        2
                    ],
                    "toDate": [
                        2
                    ]
                }
            ],
            "taxLots": [
                88,
                {
                    "instrumentId": [
                        11
                    ],
                    "status": [
                        86
                    ]
                }
            ],
            "realizedCapitalGainsReport": [
                96,
                {
                    "financialYear": [
                        14
                    ],
                    "fromDate": [
                        2
                    ],
                    "toDate": [
                        2
                    ]
                }
            ],
            "exportTaxScheduleCSV": [
                2,
                {
                    "financialYear": [
                        14
                    ],
                    "fromDate": [
                        2
                    ],
                    "toDate": [
                        2
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
                16
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
                11
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
            "withholdingTax": [
                1
            ],
            "fxRateToBase": [
                0
            ],
            "externalRef": [
                2
            ],
            "frankingCredit": [
                1
            ],
            "frankingPct": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "TransactionsConnection": {
            "items": [
                18
            ],
            "totalCount": [
                14
            ],
            "page": [
                14
            ],
            "pageSize": [
                14
            ],
            "__typename": [
                2
            ]
        },
        "DeleteTransactionPayload": {
            "success": [
                9
            ],
            "portfolio": [
                44
            ],
            "__typename": [
                2
            ]
        },
        "DeleteInstrumentPayload": {
            "success": [
                9
            ],
            "id": [
                11
            ],
            "__typename": [
                2
            ]
        },
        "Mutation": {
            "addTransaction": [
                6,
                {
                    "input": [
                        5,
                        "AddTransactionInput!"
                    ]
                }
            ],
            "deleteTransaction": [
                20,
                {
                    "id": [
                        11,
                        "ID!"
                    ]
                }
            ],
            "updateTransaction": [
                8,
                {
                    "id": [
                        11,
                        "ID!"
                    ],
                    "input": [
                        7,
                        "UpdateTransactionInput!"
                    ]
                }
            ],
            "createInstrument": [
                10,
                {
                    "input": [
                        33,
                        "CreateInstrumentInput!"
                    ]
                }
            ],
            "updateInstrument": [
                10,
                {
                    "input": [
                        34,
                        "UpdateInstrumentInput!"
                    ]
                }
            ],
            "deleteInstrument": [
                21,
                {
                    "id": [
                        11,
                        "ID!"
                    ]
                }
            ],
            "recordPriceOverride": [
                36,
                {
                    "input": [
                        35,
                        "RecordPriceOverrideInput!"
                    ]
                }
            ],
            "triggerMarketSync": [
                39,
                {
                    "symbols": [
                        2,
                        "[String!]"
                    ],
                    "syncFx": [
                        9
                    ]
                }
            ],
            "triggerDividendSync": [
                41,
                {
                    "input": [
                        40
                    ]
                }
            ],
            "triggerBackfill": [
                43,
                {
                    "input": [
                        42,
                        "TriggerBackfillInput!"
                    ]
                }
            ],
            "updateUserPreferences": [
                50,
                {
                    "input": [
                        49,
                        "UpdateUserPreferencesInput!"
                    ]
                }
            ],
            "recordFXRateOverride": [
                57,
                {
                    "input": [
                        56,
                        "RecordFXRateOverrideInput!"
                    ]
                }
            ],
            "importTransactions": [
                30,
                {
                    "input": [
                        29,
                        "BatchImportTransactionsInput!"
                    ]
                }
            ],
            "register": [
                23,
                {
                    "input": [
                        26,
                        "RegisterInput!"
                    ]
                }
            ],
            "login": [
                23,
                {
                    "input": [
                        27,
                        "LoginInput!"
                    ]
                }
            ],
            "refreshToken": [
                23
            ],
            "logout": [
                9
            ],
            "changePassword": [
                24,
                {
                    "input": [
                        25,
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
                47
            ],
            "accessToken": [
                2
            ],
            "expiresIn": [
                14
            ],
            "__typename": [
                2
            ]
        },
        "ChangePasswordPayload": {
            "success": [
                9
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
                9
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
                9
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
            "frankingCredit": [
                0
            ],
            "frankingPct": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "BatchImportTransactionsInput": {
            "transactions": [
                28
            ],
            "skipDuplicates": [
                9
            ],
            "__typename": [
                2
            ]
        },
        "BatchImportTransactionsPayload": {
            "success": [
                9
            ],
            "importedCount": [
                14
            ],
            "skippedCount": [
                14
            ],
            "portfolio": [
                44
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
                11
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
                31
            ],
            "totalCount": [
                14
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
                11
            ],
            "name": [
                2
            ],
            "isActive": [
                9
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
                9
            ],
            "__typename": [
                2
            ]
        },
        "RecordPriceOverridePayload": {
            "price": [
                31
            ],
            "valuationsRecomputed": [
                9
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
                37
            ],
            "trackedInstruments": [
                14
            ],
            "trackedCurrencies": [
                14
            ],
            "latestPriceDate": [
                2
            ],
            "latestFxDate": [
                2
            ],
            "rateLimitRemaining": [
                14
            ],
            "rateLimitBudget": [
                14
            ],
            "pendingBackfillJobs": [
                14
            ],
            "__typename": [
                2
            ]
        },
        "MarketSyncPayload": {
            "success": [
                9
            ],
            "pricesSynced": [
                14
            ],
            "fxRatesSynced": [
                14
            ],
            "message": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "TriggerDividendSyncInput": {
            "symbols": [
                2
            ],
            "exchangeCode": [
                2
            ],
            "syncHistorical": [
                9
            ],
            "forceProvider": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "DividendSyncPayload": {
            "success": [
                9
            ],
            "schedulesSynced": [
                14
            ],
            "primarySourceCount": [
                14
            ],
            "fallbackSourceCount": [
                14
            ],
            "placeholdersUpgraded": [
                14
            ],
            "message": [
                2
            ],
            "durationMs": [
                14
            ],
            "failedSymbols": [
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
                9
            ],
            "backfillFx": [
                9
            ],
            "recomputeValuations": [
                9
            ],
            "__typename": [
                2
            ]
        },
        "BackfillPayload": {
            "success": [
                9
            ],
            "pricesSynced": [
                14
            ],
            "fxRatesSynced": [
                14
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
                45
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
                11
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
                9
            ],
            "isClosed": [
                9
            ],
            "__typename": [
                2
            ]
        },
        "UserRole": {},
        "UserPreferences": {
            "userId": [
                11
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
            "role": [
                46
            ],
            "mustChangePassword": [
                9
            ],
            "financialYearEndMonth": [
                14
            ],
            "financialYearEndDay": [
                14
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
            "financialYearEndMonth": [
                14
            ],
            "financialYearEndDay": [
                14
            ],
            "__typename": [
                2
            ]
        },
        "UpdateUserPreferencesPayload": {
            "preferences": [
                47
            ],
            "portfolio": [
                44
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
                14
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
                52
            ],
            "totalCount": [
                14
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
                54
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
                9
            ],
            "__typename": [
                2
            ]
        },
        "RecordFXRateOverridePayload": {
            "rate": [
                52
            ],
            "valuationsRecomputed": [
                9
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
                58
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
            "unrealizedFxGainLoss": [
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
                11
            ],
            "eventDate": [
                2
            ],
            "type": [
                3
            ],
            "flowDirection": [
                59
            ],
            "category": [
                60
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
                62
            ],
            "breakdown": [
                63
            ],
            "items": [
                64
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
                11
            ],
            "portfolioId": [
                11
            ],
            "instrumentId": [
                11
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
                66
            ],
            "totalCount": [
                14
            ],
            "uniqueInstrumentsCount": [
                14
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
                11
            ],
            "portfolioId": [
                11
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
                11
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
                68
            ],
            "totalCount": [
                14
            ],
            "uniquePairsCount": [
                14
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
                11
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
                14
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
                9
            ],
            "isComputable": [
                9
            ],
            "diagnosticCode": [
                2
            ],
            "cashFlowWeights": [
                71
            ],
            "__typename": [
                2
            ]
        },
        "ExternalCashFlowWeightItem": {
            "transactionId": [
                11
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
                11
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
                9
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
                9
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
                11
            ],
            "baseCurrency": [
                2
            ],
            "benchmark": [
                73
            ],
            "currencyMode": [
                72
            ],
            "startDate": [
                2
            ],
            "endDate": [
                2
            ],
            "elapsedDays": [
                14
            ],
            "points": [
                74
            ],
            "metrics": [
                75
            ],
            "__typename": [
                2
            ]
        },
        "DividendFrequency": {},
        "DividendStatus": {},
        "DividendCalendarEvent": {
            "id": [
                11
            ],
            "instrumentId": [
                11
            ],
            "symbol": [
                2
            ],
            "name": [
                2
            ],
            "exDate": [
                2
            ],
            "recordDate": [
                2
            ],
            "paymentDate": [
                2
            ],
            "sharesHeld": [
                0
            ],
            "dpsDeclared": [
                1
            ],
            "grossAmountBase": [
                1
            ],
            "netAmountBase": [
                1
            ],
            "frequency": [
                77
            ],
            "status": [
                78
            ],
            "isUserEntitled": [
                9
            ],
            "isEstimated": [
                9
            ],
            "frankingPct": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "MonthlyDividendProjection": {
            "year": [
                14
            ],
            "month": [
                14
            ],
            "projectedIncome": [
                1
            ],
            "dividendCount": [
                14
            ],
            "__typename": [
                2
            ]
        },
        "HoldingYieldItem": {
            "instrumentId": [
                11
            ],
            "symbol": [
                2
            ],
            "name": [
                2
            ],
            "quantity": [
                0
            ],
            "currentPrice": [
                1
            ],
            "averageBuyPrice": [
                1
            ],
            "currentMarketValue": [
                1
            ],
            "costBasis": [
                1
            ],
            "annualDps": [
                1
            ],
            "projectedAnnualCash": [
                1
            ],
            "currentYieldPct": [
                0
            ],
            "yieldOnCostPct": [
                0
            ],
            "dividendCagr3y": [
                0
            ],
            "frequency": [
                77
            ],
            "nextExDate": [
                2
            ],
            "nextPaymentDate": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "DividendYieldAnalytics": {
            "annualForwardIncome": [
                1
            ],
            "monthlyAverageIncome": [
                1
            ],
            "portfolioCurrentYield": [
                0
            ],
            "portfolioYieldOnCost": [
                0
            ],
            "pendingCashInflow": [
                1
            ],
            "monthlyProjections": [
                80
            ],
            "holdingsYield": [
                81
            ],
            "__typename": [
                2
            ]
        },
        "DRIPSimulationInput": {
            "horizonYears": [
                14
            ],
            "dividendGrowthRatePct": [
                0
            ],
            "capitalAppreciationPct": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "DRIPSimulationPoint": {
            "year": [
                14
            ],
            "dripPortfolioValue": [
                0
            ],
            "cashPortfolioValue": [
                0
            ],
            "dripAccumulatedShares": [
                0
            ],
            "cashAccumulatedCash": [
                0
            ],
            "dripAnnualDividend": [
                0
            ],
            "cashAnnualDividend": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "DRIPSimulationResult": {
            "startingPortfolioValue": [
                0
            ],
            "startingAnnualDividend": [
                0
            ],
            "baseCurrency": [
                2
            ],
            "horizonYears": [
                14
            ],
            "points": [
                84
            ],
            "terminalDripValue": [
                0
            ],
            "terminalCashValue": [
                0
            ],
            "dripBenefitAmount": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "TaxLotStatus": {},
        "HoldingTermClassification": {},
        "TaxLot": {
            "id": [
                11
            ],
            "instrumentId": [
                11
            ],
            "ticker": [
                2
            ],
            "name": [
                2
            ],
            "openTransactionId": [
                11
            ],
            "acquiredDate": [
                2
            ],
            "closedDate": [
                2
            ],
            "originalQuantity": [
                0
            ],
            "remainingQuantity": [
                0
            ],
            "costBasis": [
                1
            ],
            "costBasisBase": [
                1
            ],
            "unitCostBase": [
                1
            ],
            "currentPrice": [
                1
            ],
            "marketValueBase": [
                1
            ],
            "unrealizedPnlBase": [
                1
            ],
            "unrealizedPnlPercent": [
                0
            ],
            "holdingPeriodDays": [
                14
            ],
            "termClassification": [
                87
            ],
            "isCgtDiscountEligible": [
                9
            ],
            "isLossHarvestingCandidate": [
                9
            ],
            "hasWashSaleRisk": [
                9
            ],
            "__typename": [
                2
            ]
        },
        "RealizedDisposal": {
            "disposalId": [
                11
            ],
            "taxLotId": [
                11
            ],
            "sellTransactionId": [
                11
            ],
            "instrumentId": [
                11
            ],
            "ticker": [
                2
            ],
            "name": [
                2
            ],
            "acquiredDate": [
                2
            ],
            "disposedDate": [
                2
            ],
            "quantityDisposed": [
                0
            ],
            "costBasisRelievedBase": [
                1
            ],
            "proceedsBase": [
                1
            ],
            "realizedPnlBase": [
                1
            ],
            "realizedPnlPercent": [
                0
            ],
            "holdingPeriodDays": [
                14
            ],
            "termClassification": [
                87
            ],
            "isCgtDiscountEligible": [
                9
            ],
            "discountedTaxablePnlBase": [
                1
            ],
            "__typename": [
                2
            ]
        },
        "TaxableIncomeSummary": {
            "dividendAndInterestIncome": [
                1
            ],
            "netCapitalGains": [
                1
            ],
            "frankingCredits": [
                1
            ],
            "totalAssessableIncome": [
                1
            ],
            "dividendInterestRatioPct": [
                0
            ],
            "capitalGainsRatioPct": [
                0
            ],
            "frankingCreditsRatioPct": [
                0
            ],
            "__typename": [
                2
            ]
        },
        "CapitalGainsSummary": {
            "totalCapitalGains": [
                1
            ],
            "lossesApplied": [
                1
            ],
            "cgtDiscountAmount": [
                1
            ],
            "netCapitalGain": [
                1
            ],
            "cgtDiscountRatePct": [
                0
            ],
            "gainDisposalsCount": [
                14
            ],
            "lossDisposalsCount": [
                14
            ],
            "__typename": [
                2
            ]
        },
        "TaxOffsetsSummary": {
            "frankingCreditsOffset": [
                1
            ],
            "foreignTaxOffsets": [
                1
            ],
            "totalAvailableOffsets": [
                1
            ],
            "__typename": [
                2
            ]
        },
        "TaxReportSummary": {
            "taxableIncome": [
                90
            ],
            "capitalGains": [
                91
            ],
            "taxOffsets": [
                92
            ],
            "fiscalYearLabel": [
                2
            ],
            "baseCurrency": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "TaxReport": {
            "summary": [
                93
            ],
            "disposals": [
                89
            ],
            "openLots": [
                88
            ],
            "fromDate": [
                2
            ],
            "toDate": [
                2
            ],
            "fiscalYearLabel": [
                2
            ],
            "baseCurrency": [
                2
            ],
            "__typename": [
                2
            ]
        },
        "RealizedGainsSummary": {
            "totalProceedsBase": [
                1
            ],
            "totalCostBasisRelieved": [
                1
            ],
            "grossRealizedGainsBase": [
                1
            ],
            "grossRealizedLossesBase": [
                1
            ],
            "netRealizedPnlBase": [
                1
            ],
            "shortTermGainsBase": [
                1
            ],
            "longTermGainsBase": [
                1
            ],
            "discountedTaxableGainBase": [
                1
            ],
            "totalDisposalCount": [
                14
            ],
            "__typename": [
                2
            ]
        },
        "RealizedGainsReport": {
            "summary": [
                95
            ],
            "disposals": [
                89
            ],
            "fromDate": [
                2
            ],
            "toDate": [
                2
            ],
            "baseCurrency": [
                2
            ],
            "__typename": [
                2
            ]
        }
    }
}