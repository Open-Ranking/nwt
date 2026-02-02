package nwt

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// MaxClaims defines the maximum number of claims allowed in a NWT, to prevent abuse.
const MaxClaims = 512

// Registered claim names as per NWT specification.
const (
	ClaimIssuer     = "iss"
	ClaimSubject    = "sub"
	ClaimAudience   = "aud"
	ClaimIssuedAt   = "iat"
	ClaimExpiration = "exp"
	ClaimNotBefore  = "nbf"
)

// Token represents a parsed Nostr Web Token (NWT) from a Nostr event.
// It includes registered claims as well as any additional claims found in the event tags.
// Learn more about NWTs at: https://github.com/pippellia-btc/nostr-web-tokens
type Token struct {
	ID         string // The ID of the Nostr event
	Issuer     string
	Subject    string
	Audience   []string
	IssuedAt   time.Time
	Expiration time.Time
	NotBefore  time.Time

	// Additional custom claims, preserved as raw tags for roundtrip compatibility.
	Extra nostr.Tags
}

func (t Token) String() string {
	return fmt.Sprintf("Token\n"+
		"\tID: %s\n"+
		"\tIssuer: %s\n"+
		"\tSubject: %s\n"+
		"\tAudience: %v\n"+
		"\tIssuedAt: %s\n"+
		"\tExpiration: %s\n"+
		"\tNotBefore: %s\n"+
		"\tExtra: %v}",
		t.ID, t.Issuer, t.Subject, t.Audience, t.IssuedAt, t.Expiration, t.NotBefore, t.Extra)
}

// IsActive checks whether the token is currently active.
// It's a shorthand for Token.IsActiveAt(time.Now(), skew).
func (t Token) IsActive(skew time.Duration) bool {
	return t.IsActiveAt(time.Now(), skew)
}

// IsActiveAt checks whether the token is active at the specified time, which happens iff
//
//	NotBefore - skew <= now <= Expiration + skew
//
// Skew is used to account for clock differences between systems, and is typically a small duration like 60s.
func (t Token) IsActiveAt(now time.Time, skew time.Duration) bool {
	if now.Before(t.NotBefore.Add(-skew)) {
		return false
	}
	if now.After(t.Expiration.Add(skew)) {
		return false
	}
	return true
}

// ToTags converts the Token into a list of nostr tags suitable for inclusion in a Nostr event.
func (t Token) ToTags() nostr.Tags {
	size := 2 + len(t.Audience) + len(t.Extra)
	tags := make(nostr.Tags, 0, size)

	if t.Issuer != "" {
		tags = append(tags, nostr.Tag{ClaimIssuer, t.Issuer})
	}
	if t.Subject != "" {
		tags = append(tags, nostr.Tag{ClaimSubject, t.Subject})
	}
	for _, audience := range t.Audience {
		tags = append(tags, nostr.Tag{ClaimAudience, audience})
	}
	if !t.IssuedAt.IsZero() {
		tags = append(tags, nostr.Tag{ClaimIssuedAt, strconv.FormatInt(t.IssuedAt.Unix(), 10)})
	}
	if !t.Expiration.IsZero() {
		tags = append(tags, nostr.Tag{ClaimExpiration, strconv.FormatInt(t.Expiration.Unix(), 10)})
	}
	if !t.NotBefore.IsZero() {
		tags = append(tags, nostr.Tag{ClaimNotBefore, strconv.FormatInt(t.NotBefore.Unix(), 10)})
	}
	tags = append(tags, t.Extra...)
	return tags
}

// Parse a [Token] from the event found in the Authorization header of the request.
// It validates the event, but doesn't validate the token's claims.
// To validate the token, use a [Validator].
func Parse(r *http.Request) (Token, error) {
	event, err := ExtractEventHTTP(r)
	if err != nil {
		return Token{}, err
	}

	if err := ValidateEvent(event); err != nil {
		return Token{}, err
	}
	return ParseToken(event)
}

// ParseToken parses the Nostr event into a [Token] struct, without performing any validation.
// To validate the token, use a [Validator].
func ParseToken(event *nostr.Event) (Token, error) {
	var err error
	token := Token{
		ID:         event.ID,
		Issuer:     event.PubKey,
		Subject:    event.PubKey,
		IssuedAt:   event.CreatedAt.Time().UTC(),
		Expiration: MaxTime,
		NotBefore:  MinTime,
	}

	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}

		switch tag[0] {
		case ClaimIssuer:
			token.Issuer = tag[1]

		case ClaimSubject:
			token.Subject = tag[1]

		case ClaimAudience:
			token.Audience = append(token.Audience, tag[1])

		case ClaimIssuedAt:
			token.IssuedAt, err = parseUnixTime(tag[1])
			if err != nil {
				return Token{}, err
			}

		case ClaimExpiration:
			token.Expiration, err = parseUnixTime(tag[1])
			if err != nil {
				return Token{}, err
			}

		case ClaimNotBefore:
			token.NotBefore, err = parseUnixTime(tag[1])
			if err != nil {
				return Token{}, err
			}

		default:
			token.Extra = append(token.Extra, tag)
		}
	}
	return token, nil
}

func parseUnixTime(t string) (time.Time, error) {
	unix, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid unix time: %w", err)
	}
	return time.Unix(unix, 0).UTC(), nil
}
