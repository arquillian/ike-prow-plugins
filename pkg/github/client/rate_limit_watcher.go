package ghclient

import (
	"github.com/arquillian/ike-prow-plugins/pkg/log"

	gogh "github.com/google/go-github/v41/github"
)

type rateLimitWatcher struct {
	client    Client
	logger    log.Logger
	threshold int
}

// NewRateLimitWatcher creates an instance of rateLimitWatcher that watches GH API rate limits
func NewRateLimitWatcher(c Client, logger log.Logger, threshold int) AroundFunctionCreator {
	return &rateLimitWatcher{client: c, logger: logger, threshold: threshold}
}

func (r rateLimitWatcher) createAroundFunction(earlierAround aroundFunction) aroundFunction {
	return func(doFunction doFunction) doFunction {
		return func(aroundContext aroundContext) (func(), *gogh.Response, error) {
			return r.logRateLimitsAfter(earlierAround(doFunction), aroundContext)
		}
	}
}

func (r rateLimitWatcher) logRateLimitsAfter(f doFunction, aroundContext aroundContext) (func(), *gogh.Response, error) {
	setValueFunc, response, err := f(aroundContext)
	r.logRateLimits()
	return setValueFunc, response, err
}

func (r rateLimitWatcher) logRateLimits() {
	limits, e := r.client.GetRateLimit()
	if e != nil {
		r.logger.Errorf("failed to load rate limits %s", e)
		return
	}
	core := limits.GetCore()
	if core.Remaining < r.threshold {
		r.logger.Warnf("reaching limit for GH API calls. %d/%d left. resetting at [%s]",
			core.Remaining, core.Limit, core.Reset.Format("2006-01-01 15:15:15"))
	}
}
