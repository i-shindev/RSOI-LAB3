package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gateway/src/domain"
)

const usernameHeader = "X-User-Name"

type request struct {
	method   string
	path     string
	query    url.Values
	username string
	body     any
}

type client struct {
	name            string
	baseURL         string
	http            *http.Client
	circuit_breaker CircuitBreakerStrategy
}

func newClient(name, baseURL string, timeout time.Duration) client {
	return client{
		name:            name,
		baseURL:         strings.TrimRight(baseURL, "/"),
		http:            &http.Client{Timeout: timeout},
		circuit_breaker: DefaultCircuitBreaker(),
	}
}

func (c client) do(ctx context.Context, req request, out any) error {
	if IsOpen(c.circuit_breaker) {
		return domain.NewError(domain.ErrCircuitOpen, "%s unavailable", c.name)
	}
	target := c.baseURL + req.path
	if len(req.query) > 0 {
		target += "?" + req.query.Encode()
	}

	var body io.Reader
	if req.body != nil {
		payload, err := json.Marshal(req.body)
		if err != nil {
			return err
		}
		body = bytes.NewReader(payload)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, req.method, target, body)
	if err != nil {
		return err
	}
	if req.body != nil {
		httpRequest.Header.Set("Content-Type", "application/json")
	}
	if req.username != "" {
		httpRequest.Header.Set(usernameHeader, req.username)
	}

	response, err := c.http.Do(httpRequest)
	if err != nil {
		c.circuit_breaker.OnRequestCompleted(false)
		return c.unavailable(err)
	}
	defer response.Body.Close()
	err, alive := c.readResponse(response, req, out)
	c.circuit_breaker.OnRequestCompleted(alive)
	return err
}
func (c client) readResponse(response *http.Response, req request, out any) (error, bool) {
	switch status := response.StatusCode; {
	case status >= 200 && status < 300:
		if out == nil {
			return nil, true
		}
		if err := json.NewDecoder(response.Body).Decode(out); err != nil {
			return c.unavailable(fmt.Errorf("decode %s %s response: %w", req.method, req.path, err)), false
		}
		return nil, true
	case status == http.StatusBadRequest:
		return c.validationError(response.Body), true
	case status == http.StatusNotFound:
		return &domain.Error{Kind: domain.ErrNotFound, Message: c.message(response.Body)}, true
	case status == http.StatusConflict:
		return &domain.Error{Kind: domain.ErrConflict, Message: c.message(response.Body)}, true
	default:
		return c.unavailable(fmt.Errorf("%s %s answered %d", req.method, req.path, status)), false
	}
}

func (c client) unavailable(cause error) error {
	return &domain.Error{Kind: domain.ErrServiceUnavailable, Message: c.name + " unavailable", Cause: cause}
}

func (c client) message(body io.Reader) string {
	var decoded errorResponse
	if err := json.NewDecoder(body).Decode(&decoded); err != nil || decoded.Message == "" {
		return c.name + " rejected the request"
	}
	return decoded.Message
}

func (c client) validationError(body io.Reader) error {
	var decoded validationErrorResponse
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return c.unavailable(fmt.Errorf("decode validation error: %w", err))
	}
	fields := make([]domain.FieldError, 0, len(decoded.Errors))
	for _, e := range decoded.Errors {
		fields = append(fields, domain.FieldError{Field: e.Field, Message: e.Error})
	}
	return &domain.ValidationError{Fields: fields}
}

type errorResponse struct {
	Message string `json:"message"`
}

type validationErrorResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Field string `json:"field"`
		Error string `json:"error"`
	} `json:"errors"`
}
