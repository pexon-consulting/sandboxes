package log

import "github.com/sirupsen/logrus"

type AdditionalLogfields map[string]string

type LogConfig struct {
	Package   string
	Logfields *AdditionalLogfields
}

func GetLogger(input LogConfig) *logrus.Entry {
	fields := logrus.Fields{
		"Package": input.Package,
	}
	if input.Logfields != nil {
		for key, item := range *input.Logfields {
			fields[key] = item
		}
	}
	return logrus.WithFields(fields)
}
