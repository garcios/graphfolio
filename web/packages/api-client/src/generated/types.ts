export default {
    "scalars": [
        0,
        2,
        3,
        7,
        9,
        10,
        16
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
                18
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
            "__typename": [
                2
            ]
        },
        "ID": {},
        "Query": {
            "portfolio": [
                18
            ],
            "instruments": [
                6
            ],
            "portfolioHistory": [
                12,
                {
                    "timeframe": [
                        10,
                        "HistoryTimeframe!"
                    ]
                }
            ],
            "transactions": [
                14,
                {
                    "type": [
                        3
                    ],
                    "symbol": [
                        2
                    ],
                    "page": [
                        9
                    ],
                    "pageSize": [
                        9
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
                11
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
                13
            ],
            "totalCount": [
                9
            ],
            "page": [
                9
            ],
            "pageSize": [
                9
            ],
            "__typename": [
                2
            ]
        },
        "DeleteTransactionPayload": {
            "success": [
                16
            ],
            "portfolio": [
                18
            ],
            "__typename": [
                2
            ]
        },
        "Boolean": {},
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
                15,
                {
                    "id": [
                        7,
                        "ID!"
                    ]
                }
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
                19
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
        }
    }
}