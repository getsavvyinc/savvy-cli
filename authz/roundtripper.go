package authz

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrInvalidAuthzClient = errors.New("invalid authz client")

type AuthorizedRoundTripper struct {
	token        string
	savvyVersion string
	// wrap error returned by RoundTrip
	wrapErr error
}

// NewRoundTripper returns a new AuthorizedRoundTripper
//
// Caller must provide non nil err to wrap the error returned by RoundTrip
func NewRoundTripper(token, savvyVersion string, err error) *AuthorizedRoundTripper {
	return &AuthorizedRoundTripper{token: token, savvyVersion: savvyVersion, wrapErr: err}
}

func (a *AuthorizedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to ensure thread safety
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("Authorization", "Bearer "+a.token)
	clonedReq.Header.Set("X-Savvy-Version", a.savvyVersion)

	// Use the embedded Transport to perform the actual request
	res, err := http.DefaultTransport.RoundTrip(clonedReq)
	if err != nil {
		err = fmt.Errorf("%w: %v", a.wrapErr, err)
		return nil, err
	}

	// If we get a 401 Unauthorized, then the token is expired
	// and we need to refresh it
	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("%w: invalid token", a.wrapErr)
	}
	return res, err
}
