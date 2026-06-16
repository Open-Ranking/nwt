package nwt

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/nbd-wtf/go-nostr"
)

func TestExtractEventHTTP(t *testing.T) {
	tests := []struct {
		request *http.Request
		event   *nostr.Event
		err     error
	}{
		{
			request: &http.Request{Header: http.Header{}},
			err:     ErrMissingHeader,
		},
		{
			request: &http.Request{
				Header: http.Header{"Authorization": []string{"invalid"}},
			},
			err: ErrInvalidHeader,
		},
		{
			request: &http.Request{
				Header: http.Header{"Authorization": []string{"Nostr invalidbase64"}},
			},
			err: ErrInvalidHeader,
		},
		{
			request: &http.Request{
				Header: http.Header{"Authorization": []string{"Nostr eyJraW5kIjoxLCJpZCI6ImMzZTM5YjU0MjgxMzk0NTk0NmM2YWI0MTk1ODliMjQzYjc4YjJhOGI0NTNiNTA2YTBhZjMwZTM0ZGRhYTFmYjciLCJwdWJrZXkiOiI3OWJlNjY3ZWY5ZGNiYmFjNTVhMDYyOTVjZTg3MGIwNzAyOWJmY2RiMmRjZTI4ZDk1OWYyODE1YjE2ZjgxNzk4IiwiY3JlYXRlZF9hdCI6MTc2Nzk1Njg1MSwidGFncyI6W10sImNvbnRlbnQiOiJoZWxsbyBmcm9tIHRoZSBub3N0ciBhcm15IGtuaWZlIiwic2lnIjoiM2Q2YjIxYjgzN2IwYWYzNWEwYWViN2QyODY5MjdhNDA4MzlmNTkwOTQ3ZjRjNjI1ZTdjOGQ2ZWM2Nzg4NWRkNDA2NmU1ZTNhMGNlY2U0NTA1ZmI4NmU1NzFlM2Y0Zjk1ZjNjZjgxNjRjZWRkNTJhYWQ4MTdiODE4ZDYwNjY3MzQifQ"}},
			},
			event: &nostr.Event{
				ID:        "c3e39b542813945946c6ab419589b243b78b2a8b453b506a0af30e34ddaa1fb7",
				PubKey:    "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
				CreatedAt: nostr.Timestamp(1767956851),
				Kind:      1,
				Tags:      nostr.Tags{},
				Content:   "hello from the nostr army knife",
				Sig:       "3d6b21b837b0af35a0aeb7d286927a40839f590947f4c625e7c8d6ec67885dd4066e5e3a0cece4505fb86e571e3f4f95f3cf8164cedd52aad817b818d6066734",
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			event, err := ExtractEventHTTP(test.request)
			if !errors.Is(err, test.err) {
				t.Fatalf("expected error %v, got %v", test.err, err)
			}
			if !reflect.DeepEqual(event, test.event) {
				t.Errorf("expected event %v, got %v", test.event, event)
			}
		})
	}
}

func TestSetAuthRoundtrip(t *testing.T) {
	event := &nostr.Event{
		ID:        "366458cb01dd1f42d66cb71d31cc2e1217c69606181c83cbcdeb878942776d73",
		PubKey:    "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
		CreatedAt: 1767957502,
		Kind:      Kind,
		Tags:      nostr.Tags{},
		Sig:       "7c9a84e33fa7aaf6d85c3d90b3103b4197d7f964f5ff31dabe49aa4952b74579e4cfe6c4c4635e2501f5dbd742fdc4750a5ce26aae395a9b256a27b5533575b9",
	}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if err := SetAuth(r, event); err != nil {
		t.Fatalf("SetAuth failed: %v", err)
	}

	got, err := ExtractEventHTTP(r)
	if err != nil {
		t.Fatalf("ExtractEventHTTP failed: %v", err)
	}

	if !reflect.DeepEqual(got, event) {
		t.Errorf("roundtrip mismatch:\nwant: %v\ngot:  %v", event, got)
	}
}

func TestValidateEvent(t *testing.T) {
	tests := []struct {
		event *nostr.Event
		err   error
	}{
		{
			event: &nostr.Event{Kind: 1},
			err:   ErrInvalidEventKind,
		},
		{
			event: &nostr.Event{Kind: Kind, CreatedAt: -1},
			err:   ErrInvalidEventCreatedAt,
		},
		{
			event: &nostr.Event{Kind: Kind, CreatedAt: 99999999999999999},
			err:   ErrInvalidEventCreatedAt,
		},
		{
			event: &nostr.Event{
				ID:        "366458c___invalid____31cc2e1217c69606181c83cbcdeb878942776d73",
				PubKey:    "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
				CreatedAt: 1767957502,
				Kind:      Kind,
				Sig:       "7c9a84e33fa7aaf6d85c3d90b3103b4197d7f964f5ff31dabe49aa4952b74579e4cfe6c4c4635e2501f5dbd742fdc4750a5ce26aae395a9b256a27b5533575b9",
			},
			err: ErrInvalidEventID,
		},
		{
			event: &nostr.Event{
				ID:        "366458cb01dd1f42d66cb71d31cc2e1217c69606181c83cbcdeb878942776d73",
				PubKey:    "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
				CreatedAt: 1767957502,
				Kind:      Kind,
				Sig:       "7c9a84e333____invalid____579e4cfe6c4c4635e2501f5dbd742fdc4750a5ce26aae395a9b256a27b5533575b9",
			},
			err: ErrInvalidEventSignature,
		},
		{
			event: &nostr.Event{
				ID:        "366458cb01dd1f42d66cb71d31cc2e1217c69606181c83cbcdeb878942776d73",
				PubKey:    "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
				CreatedAt: 1767957502,
				Kind:      Kind,
				Sig:       "7c9a84e33fa7aaf6d85c3d90b3103b4197d7f964f5ff31dabe49aa4952b74579e4cfe6c4c4635e2501f5dbd742fdc4750a5ce26aae395a9b256a27b5533575b9",
			},
			err: nil,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			err := ValidateEvent(test.event)
			if !errors.Is(err, test.err) {
				t.Fatalf("expected error %v, got %v", test.err, err)
			}
		})
	}
}
