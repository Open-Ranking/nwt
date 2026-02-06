package nwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// Kind is the Nostr event kind for Nostr Web Tokens (NWT).
const Kind = 27519

// Parsing errors
var (
	ErrMissingHeader    = errors.New("missing Authorization header")
	ErrInvalidHeader    = errors.New("invalid Authorization header format")
	ErrInvalidEventJSON = errors.New("invalid event JSON")
)

// Event validation errors
var (
	ErrInvalidEventKind      = fmt.Errorf("event kind must be %d", Kind)
	ErrInvalidEventCreatedAt = errors.New("invalid event created at")
	ErrInvalidEventID        = errors.New("invalid event ID")
	ErrInvalidEventSignature = errors.New("invalid event signature")
)

// ExtractEventHTTP extracts the Nostr event from the Authorization header
// of the HTTP request without performing any validation.
func ExtractEventHTTP(r *http.Request) (*nostr.Event, error) {
	if r == nil {
		return nil, errors.New("request is nil")
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, ErrMissingHeader
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return nil, ErrInvalidHeader
	}

	if parts[0] != "Nostr" {
		return nil, ErrInvalidHeader
	}

	bytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidHeader, err)
	}

	event := &nostr.Event{}
	if err := json.Unmarshal(bytes, event); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidEventJSON, err)
	}
	return event, nil
}

// ValidateEvent checks whether the given Nostr event is a valid NWT event.
// It verifies the event kind, created at, ID, and signature but doesn't validate the token claims themselves.
func ValidateEvent(e *nostr.Event) error {
	if e == nil {
		return errors.New("event is nil")
	}
	if e.Kind != Kind {
		return ErrInvalidEventKind
	}
	if int64(e.CreatedAt) < MinTime.Unix() {
		return fmt.Errorf("%w: created at cannot be negative", ErrInvalidEventCreatedAt)
	}
	if int64(e.CreatedAt) > MaxTime.Unix() {
		return fmt.Errorf("%w: created at exceeds maximum time", ErrInvalidEventCreatedAt)
	}
	if len(e.Tags) > MaxClaims {
		return errors.New("too many tags in event")
	}
	if !e.CheckID() {
		return ErrInvalidEventID
	}
	match, err := e.CheckSignature()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidEventSignature, err)
	}
	if !match {
		return ErrInvalidEventSignature
	}
	return nil
}
