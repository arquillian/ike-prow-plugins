package ghclient

import (
	gogh "github.com/google/go-github/github"
)

type paginationChecker struct {
}

// NewPaginationChecker creates an instance of paginationChecker that checks if there is a next page with additional results
func NewPaginationChecker() AroundFunctionCreator {
	return &paginationChecker{}
}

func (r paginationChecker) createAroundFunction(earlierAround aroundFunction) aroundFunction {
	return func(doFunction doFunction) doFunction {
		return func(doContext aroundContext) (func(), *gogh.Response, error) {
			return r.checkNextPage(doFunction, doContext)
		}
	}
}

func (r paginationChecker) checkNextPage(do doFunction, context aroundContext) (func(), *gogh.Response, error) {
	var (
		setValueFunc func()
		response     *gogh.Response
		err          error
	)
	context.pageNumber = 1

	for context.pageNumber != 0 {
		setValueFunc, response, err = do(context)
		setValueFunc()
		if err != nil {
			break
		}
		context.pageNumber = response.NextPage
	}
	return setValueFunc, response, err
}
