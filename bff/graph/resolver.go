package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"context"

	"bff/internal/auth"

	pb "graphfolio/proto/portfolio/v1"
	userpb "graphfolio/proto/user/v1"
)

const defaultUserID = "018f0000-0000-7000-8000-000000000001"

type Resolver struct {
	PortfolioClient pb.PortfolioServiceClient
	UserClient      userpb.UserServiceClient
}

func (r *Resolver) resolveUserID(ctx context.Context) string {
	if uid, err := auth.GetUserID(ctx); err == nil && uid != "" {
		return uid
	}
	return defaultUserID
}

func (r *Resolver) resolveUserRole(ctx context.Context) string {
	if role, err := auth.GetUserRole(ctx); err == nil && role != "" {
		return role
	}
	if uid, err := auth.GetUserID(ctx); err == nil && uid != "" && r.UserClient != nil {
		resp, err := r.UserClient.GetUserPreferences(ctx, &userpb.GetUserPreferencesRequest{UserId: uid})
		if err == nil && resp.Preferences != nil {
			if resp.Preferences.Role == userpb.UserRole_USER_ROLE_ADMIN {
				return "ADMIN"
			}
			return "USER"
		}
	}
	return "USER"
}

