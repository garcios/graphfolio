export default {
    "scalars": [
        0,
        2,
        3,
        7,
        8,
        11,
        12
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
                34
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
                34
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
                24,
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
                30
            ],
            "userPreferences": [
                36
            ],
            "supportedCurrencies": [
                37
            ],
            "currencyPairs": [
                40
            ],
            "currencyPairHistory": [
                44,
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
                42,
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
                34
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
                        25,
                        "CreateInstrumentInput!"
                    ]
                }
            ],
            "updateInstrument": [
                6,
                {
                    "input": [
                        26,
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
                28,
                {
                    "input": [
                        27,
                        "RecordPriceOverrideInput!"
                    ]
                }
            ],
            "triggerMarketSync": [
                31,
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
                33,
                {
                    "input": [
                        32,
                        "TriggerBackfillInput!"
                    ]
                }
            ],
            "updateUserPreferences": [
                39,
                {
                    "input": [
                        38,
                        "UpdateUserPreferencesInput!"
                    ]
                }
            ],
            "recordFXRateOverride": [
                46,
                {
                    "input": [
                        45,
                        "RecordFXRateOverrideInput!"
                    ]
                }
            ],
            "importTransactions": [
                22,
                {
                    "input": [
                        21,
                        "BatchImportTransactionsInput!"
                    ]
                }
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
                20
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
                34
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
                23
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
                23
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
                29
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
                35
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
                36
            ],
            "portfolio": [
                34
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
                41
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
                43
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
                41
            ],
            "valuationsRecomputed": [
                8
            ],
            "__typename": [
                2
            ]
        }
    }
}