package savvy_errors

import "errors"

var ErrInvalidToken = errors.New("expired token")
var ErrMissingConfig = errors.New("missing config file")
