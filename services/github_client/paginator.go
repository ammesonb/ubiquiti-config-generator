package github_client

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/google/go-github/v70/github"
)

var (
	errStopIteration = errors.Err("no more values to iterate")
)

// paginateGitHub takes a function that returns next page of a datatype, and returns channels for data, send next, and error
//
// expected use is to select from data and err, and send to next after each data entry is handled
func paginateGitHub[T any](getNextPage func(page int, ctx context.Context) ([]*T, *github.Response, error), ctx context.Context) (<-chan *T, chan<- bool, <-chan error) {
	results := make(chan *T)
	next := make(chan bool)
	errorChan := make(chan error)

	values := []*T{}

	go func() {
		defer close(results)
		defer close(next)
		defer close(errorChan)

		page := 1
		for page > 0 {
			select {
			case <-ctx.Done():
				page = 0
			case <-next:
				if len(values) == 0 {
					if page == 0 {
						// Exit iteration when results exhausted and no next page
						errorChan <- errStopIteration
						return
					}

					// Otherwise, retrieve values from the next page
					pageValues, response, err := getNextPage(page, ctx)
					if err != nil {
						errorChan <- err
						return
					}
					values = append(values, pageValues...)
					page = response.NextPage
				}

				// declare explicitly so can pop from values
				var value *T
				value, values = values[0], values[1:]
				results <- value
			}
		}
	}()

	// Immediately kick off the first request/value, to save callers the trouble
	next <- true

	return results, next, errorChan
}
