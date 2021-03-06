package heroku

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/remind101/empire"
	"github.com/remind101/empire/pkg/heroku"
)

// Named matching heroku's error codes. See
// https://devcenter.heroku.com/articles/platform-api-reference#error-responses
var (
	ErrBadRequest = &ErrorResource{
		Status:  http.StatusBadRequest,
		ID:      "bad_request",
		Message: "Request invalid, validate usage and try again",
	}
	ErrUnauthorized = &ErrorResource{
		Status:  http.StatusUnauthorized,
		ID:      "unauthorized",
		Message: "Request not authenticated, API token is missing, invalid or expired",
	}
	ErrForbidden = &ErrorResource{
		Status:  http.StatusForbidden,
		ID:      "forbidden",
		Message: "Request not authorized, provided credentials do not provide access to specified resource",
	}
	ErrNotFound = &ErrorResource{
		Status:  http.StatusNotFound,
		ID:      "not_found",
		Message: "Request failed, the specified resource does not exist",
	}
	ErrTwoFactor = &ErrorResource{
		Status:  http.StatusUnauthorized,
		ID:      "two_factor",
		Message: "Two factor code is required.",
	}
	ErrSSLRemoved = &ErrorResource{
		Status:  http.StatusNotFound,
		ID:      "not_found",
		Message: "Support for uploading SSL certificates through Empire has been removed and replaced with certificate attachments.",
		URL:     "http://empire.readthedocs.org/en/latest/ssl_certs/",
	}
	ErrMessageRequired = &ErrorResource{
		Status:  http.StatusBadRequest,
		ID:      "message_required",
		Message: fmt.Sprintf("Header '%s' is required", heroku.CommitMessageHeader),
	}
)

// ErrorResource represents the error response format that we return.
type ErrorResource struct {
	Status  int    `json:"-"`
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

func newError(err error) *ErrorResource {
	if err == gorm.RecordNotFound {
		return ErrNotFound
	}

	switch err := err.(type) {
	case *ErrorResource:
		return err
	case *empire.MessageRequiredError:
		return ErrMessageRequired
	case *empire.ValidationError:
		return ErrBadRequest
	default:
		return &ErrorResource{
			Message: err.Error(),
		}
	}
}

// Error implements error interface.
func (e *ErrorResource) Error() string {
	return e.Message
}

func errNotImplemented(message string) *ErrorResource {
	return &ErrorResource{
		Status:  http.StatusNotImplemented,
		ID:      "not_implemented",
		Message: message,
	}
}

// Returns an appropriate ErrorResource when a request is unauthorized.
func Unauthorized(reason error) *ErrorResource {
	if reason == nil {
		return ErrUnauthorized
	}

	return &ErrorResource{
		Status:  http.StatusUnauthorized,
		ID:      "unauthorized",
		Message: reason.Error(),
	}
}

// SAMLUnauthorized can be used in place of Unauthorized to return a link to
// login via SAML.
func SAMLUnauthorized(loginURL string) func(error) *ErrorResource {
	return func(reason error) *ErrorResource {
		if reason == nil {
			reason = errors.New("Request not authenticated, API token is missing, invalid or expired")
		}

		return &ErrorResource{
			Status:  http.StatusUnauthorized,
			ID:      "saml_unauthorized",
			Message: fmt.Sprintf("%s. Login at %s", reason, loginURL),
		}
	}
}

// errHandler returns a handler that responds with the given error.
func errHandler(err error) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return err
	}
}
