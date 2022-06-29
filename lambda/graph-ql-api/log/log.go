package log

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/settings"

	"github.com/sirupsen/logrus"
)

type AdditionalLogfields map[string]string

type LogConfig struct {
	Package   string
	Logfields AdditionalLogfields
}

func (l LogConfig) Builder() *logrus.Entry {
	fields := logrus.Fields{
		"Package": l.Package,
	}
	if l.Logfields != nil {
		for key, item := range l.Logfields {
			fields[key] = item
		}
	}
	return logrus.WithFields(fields)
}

func GetGlobalLogger(ctx context.Context) LogConfig {

	svcUntyped := ctx.Value(settings.LogConfig)
	logConfig, b := svcUntyped.(LogConfig)
	if b {
		return logConfig
	} else {
		return LogConfig{
			Logfields: AdditionalLogfields{
				"context": "no-lambda-context",
			},
		}
	}
}

func (l LogConfig) SetField(key, val string) LogConfig {
	if l.Logfields == nil {
		m := make(map[string]string)
		m[key] = val
		l.Logfields = m
		return l
	}
	l.Logfields[key] = val
	return l
}

func (l LogConfig) SetResolver(val string) LogConfig {
	return l.SetField("Resolver", val)
}

func (l LogConfig) SetFunction(val string) LogConfig {
	return l.SetField("Function", val)
}

func (l LogConfig) SetPackage(p string) LogConfig {
	l.Package = p
	return l
}
