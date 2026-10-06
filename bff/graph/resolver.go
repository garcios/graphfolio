package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	pb "graphfolio/proto/portfolio/v1"
	userpb "graphfolio/proto/user/v1"
)

const defaultUserID = "018f0000-0000-7000-8000-000000000001"

type Resolver struct {
	PortfolioClient pb.PortfolioServiceClient
	UserClient      userpb.UserServiceClient
}
