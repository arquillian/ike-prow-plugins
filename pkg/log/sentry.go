package log

import (
	"time"

	"github.com/evalphobia/logrus_sentry"
	"github.com/sirupsen/logrus"
)

// SentryTags encapsulate map of tags where both key and values are strings. Those tags are used by Sentry
// to categorize events reported by client apps.
type SentryTags map[string]string

// SentryConfiguration keeps basic configuration used by sentry hook
type SentryConfiguration struct {
	Dsn     string
	Tags    SentryTags
	Timeout time.Duration
}

// NewSentryConfiguration creates new instance of SentryConfiguration
// dsn - full sentry DSN url
// tags - extra tags to add to every event reported to Sentry
// timeout - timeout in milliseconds after which connection to Sentry server should be dropped
func NewSentryConfiguration(dsn string, tags map[string]string, timeout int) SentryConfiguration {
	return SentryConfiguration{
		Dsn:     dsn,
		Tags:    tags,
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
}

// AddSentryHook registers logrus hook integration logger with Sentry
func AddSentryHook(log *logrus.Entry, configuration SentryConfiguration) {
	if configuration.Dsn != "" {
		levels := []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		}

		hook, err := logrus_sentry.NewWithTagsSentryHook(configuration.Dsn, configuration.Tags, levels)

		if err == nil {
			hook.Timeout = configuration.Timeout
			logrus.AddHook(hook)
		} else {
			log.WithError(err).Error("failed to add sentry hook")
		}
	}

}
