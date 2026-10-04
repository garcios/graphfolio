package internal

import (
	"context"
	pb "graphfolio/proto/portfolio/v1"
)

type PortfolioServer struct {
	pb.UnimplementedPortfolioServiceServer
}

func NewPortfolioServer() *PortfolioServer {
	return &PortfolioServer{}
}

func (s *PortfolioServer) GetPortfolio(ctx context.Context, req *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
	return &pb.GetPortfolioResponse{
		Portfolio: &pb.Portfolio{
			TotalValue:              124532.89,
			TodayReturnAmount:       1234.50,
			TodayReturnPercent:      1.00,
			AnnualizedReturnPercent: 14.2,
			CashBalance:             8450.00,
			Investments: []*pb.Investment{
				{
					Id:                 "1",
					Ticker:             "AAPL",
					Name:               "Apple Inc.",
					Price:              185.92,
					Quantity:           142.5,
					TotalValue:         26493.60,
					TodayReturnAmount:  555.75,
					TodayReturnPercent: 2.14,
					TotalReturnAmount:  4230.10,
					TotalReturnPercent: 18.9,
				},
				{
					Id:                 "2",
					Ticker:             "MSFT",
					Name:               "Microsoft",
					Price:              402.11,
					Quantity:           85.0,
					TotalValue:         34179.35,
					TodayReturnAmount:  510.00,
					TodayReturnPercent: 1.51,
					TotalReturnAmount:  8450.20,
					TotalReturnPercent: 32.8,
				},
				{
					Id:                 "3",
					Ticker:             "TSLA",
					Name:               "Tesla",
					Price:              210.45,
					Quantity:           50.0,
					TotalValue:         10522.50,
					TodayReturnAmount:  -336.00,
					TodayReturnPercent: -3.10,
					TotalReturnAmount:  -1250.00,
					TotalReturnPercent: -10.6,
				},
				{
					Id:                 "4",
					Ticker:             "NVDA",
					Name:               "NVIDIA Corp.",
					Price:              721.33,
					Quantity:           35.0,
					TotalValue:         25246.55,
					TodayReturnAmount:  1050.00,
					TodayReturnPercent: 4.34,
					TotalReturnAmount:  14500.00,
					TotalReturnPercent: 134.9,
				},
				{
					Id:                 "5",
					Ticker:             "V",
					Name:               "Visa Inc.",
					Price:              278.10,
					Quantity:           45.0,
					TotalValue:         12514.50,
					TodayReturnAmount:  12.50,
					TodayReturnPercent: 0.10,
					TotalReturnAmount:  1120.00,
					TotalReturnPercent: 9.8,
				},
			},
		},
	}, nil
}
