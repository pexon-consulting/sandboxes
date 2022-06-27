package resolver

import (
	"context"
)

func (*Resolver) UpdateSandbox(ctx context.Context) (*string, error) {
	logger.Debug("call UpdateSandbox")

	s := "s"
	logger.Debug("return Result", s)
	return &s, nil
}
