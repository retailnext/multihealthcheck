// Copyright 2022 RetailNext, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package checker

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strconv"
)

type newRequestWithContext func(ctx context.Context) (*http.Request, error)

type checker struct {
	client          *http.Client
	makeRequest     newRequestWithContext
	checkStatusCode func(int) error
}

func (c checker) doCheck(ctx context.Context) error {
	request, makeRequestErr := c.makeRequest(ctx)
	if makeRequestErr != nil {
		return makeRequestErr
	}

	response, doRequestErr := c.client.Do(request)

	var readBodyErr error
	var statusCode int
	if response != nil {
		if response.Body != nil {
			_, readBodyErr = io.Copy(io.Discard, response.Body)
			if closeBodyErr := response.Body.Close(); closeBodyErr != nil {
				panic(closeBodyErr)
			}
		}
		statusCode = response.StatusCode
	}

	if doRequestErr != nil {
		return doRequestErr
	}
	if readBodyErr != nil {
		return readBodyErr
	}
	return c.checkStatusCode(statusCode)
}

func newGetChecker(url string) checker {
	return checker{
		client: insecureHttpClient,
		makeRequest: func(ctx context.Context) (*http.Request, error) {
			return http.NewRequestWithContext(ctx, "GET", url, nil)
		},
		checkStatusCode: expect2xx,
	}
}

var insecureHttpClient = func() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Transport: transport}
}()

func expect2xx(statusCode int) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	return non2xxStatusCode(statusCode)
}

type non2xxStatusCode int

func (e non2xxStatusCode) Error() string {
	return "HTTP status " + strconv.Itoa(int(e))
}
