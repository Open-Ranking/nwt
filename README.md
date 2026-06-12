# nwt

A golang library implementing [Nostr Web Tokens](https://github.com/Open-Ranking/nostr-web-tokens).

## Installation

```
go get github.com/pippellia-btc/nwt
```

## Usage

Parsing a token from the Authorization header of an `http.Request`

```golang
token, err := nwt.Parse(request)
if err != nil {
    slog.Info("failed to parse token", "error", err)
}
```

Validating the token using the default `StrictValidator` 

```golang
validator := nwt.StrictValidator{
    Identifier: "example.com"   // domain to be present in the audience claim
    ClockSkew: time.Minute      // tolerance for clock differences
}

if err := validator.Validate(token); err != nil {
    slog.Info("token is invalid", "reason", err)
}
```

Or create a custom validator by satisfying the `Validator` interface

```golang
// Validator wraps the Validate method for validating Tokens.
// The token is considered valid iff Validate returns nil.
//
// Implementations may enforce different policies for what constitutes a valid token,
// but are generally expected to at least validate the time-based claims with [ValidateTimeBounds].
//
// As an example, check out [StrictValidator].
type Validator interface {
	Validate(Token) error
}
```
