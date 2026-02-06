package nwt

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

func TestToTags(t *testing.T) {
	tests := []struct {
		token Token
		tags  nostr.Tags
	}{
		{
			token: Token{},
			tags:  nostr.Tags{},
		},
		{
			token: Token{
				Issuer:   "issuer",
				Subject:  "subject",
				Audience: []string{"aud1", "aud2"},
			},
			tags: nostr.Tags{
				{"iss", "issuer"},
				{"sub", "subject"},
				{"aud", "aud1"},
				{"aud", "aud2"},
			},
		},
		{
			token: Token{
				Extra: nostr.Tags{{"role", "admin"}},
			},
			tags: nostr.Tags{
				{"role", "admin"},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tags := test.token.ToTags()
			if !reflect.DeepEqual(tags, test.tags) {
				t.Errorf("expected tags %v, got %v", test.tags, tags)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	tests := []struct {
		event *nostr.Event
		token Token
	}{
		{
			event: &nostr.Event{},
			token: Token{
				IssuedAt:   time.Unix(0, 0).UTC(),
				Expiration: MaxTime,
				NotBefore:  MinTime,
			},
		},
		{
			event: &nostr.Event{
				ID:        "id",
				PubKey:    "pubkey",
				CreatedAt: nostr.Timestamp(420),
				Tags: nostr.Tags{
					{"sub", "subject"},
					{"aud", "aud1"},
					{"aud", "aud2"},
					{"exp", "6969"},
				},
			},
			token: Token{
				ID:         "id",
				Signer:     "pubkey",
				Issuer:     "pubkey",
				Subject:    "subject",
				Audience:   []string{"aud1", "aud2"},
				IssuedAt:   time.Unix(420, 0).UTC(),
				Expiration: time.Unix(6969, 0).UTC(),
				NotBefore:  MinTime,
			},
		},
		{
			event: &nostr.Event{
				ID:        "id",
				PubKey:    "pubkey",
				CreatedAt: nostr.Timestamp(420),
				Tags: nostr.Tags{
					{"role", "admin"},
					{"permission", "read", "write"},
				},
			},
			token: Token{
				ID:         "id",
				Signer:     "pubkey",
				Issuer:     "pubkey",
				Subject:    "pubkey",
				IssuedAt:   time.Unix(420, 0).UTC(),
				Expiration: MaxTime,
				NotBefore:  MinTime,
				Extra: nostr.Tags{
					{"role", "admin"},
					{"permission", "read", "write"},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			token, err := ParseToken(test.event)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(token, test.token) {
				t.Errorf("\nexpected %v, \ngot %v", test.token, token)
			}
		})
	}
}
