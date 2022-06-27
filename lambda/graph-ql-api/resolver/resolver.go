package resolver

import (
	"lambda/aws-sandbox/graph-ql-api/log"

	"github.com/sirupsen/logrus"
)

type Resolver struct{}

var logger *logrus.Entry = log.GetLogger(log.LogConfig{Package: "resolver"})
