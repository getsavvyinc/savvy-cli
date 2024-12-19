package authz

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrInvalidClient = errors.New("invalid client")


type AuthorizedRoundTripper struct {
	token        string
	savvyVersion string
}
func NewRoundTripper(token, savvyVersion string) *AuthorizedRoundTripper {
	return &AuthorizedRoundTripper{token: token, savvyVersion: savvyVersion}
}

func (a *AuthorizedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to ensure thread safety
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("Authorization", "Bearer "+a.token)
	clonedReq.Header.Set("X-Savvy-Version", a.savvyVersion)

	// Use the embedded Transport to perform the actual request
	res, err := http.DefaultTransport.RoundTrip(clonedReq)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrInvalidClient, err)
		return nil, err
	}

	// If we get a 401 Unauthorized, then the token is expired
	// and we need to refresh it
	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("%w: invalid token", ErrInvalidClient)
	}
	return res, err

