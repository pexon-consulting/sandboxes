package resolver

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/log"
)

func (*Resolver) UpdateSandbox(ctx context.Context) (*string, error) {
	logger := log.GetGlobalLogger(ctx).SetPackage("Resolver").SetResolver("UpdateSandbox").Builder()

	logger.Debug("call UpdateSandbox")

	s := "s"
	logger.Debug("return Result", s)
	return &s, nil
}
