package main

import "testing"

// US3: serve flag validation. addr/token-file are HTTP-only; supplying them
// without --http is a usage error. (HTTP-requires-token is checked in ServeMCP,
// not here — see validateServeFlags doc.)

func TestValidateServeFlags(t *testing.T) {
	cases := []struct {
		name            string
		useHTTP         bool
		addr, tokenFile string
		wantErr         bool
	}{
		{"stdio default", false, "", "", false},
		{"http with addr and token", true, ":8080", "tok", false},
		{"addr without http", false, ":8080", "", true},
		{"token-file without http", false, "", "tok", true},
	}
	for _, c := range cases {
		err := validateServeFlags(c.useHTTP, c.addr, c.tokenFile)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: validateServeFlags(%v,%q,%q) err=%v, wantErr=%v",
				c.name, c.useHTTP, c.addr, c.tokenFile, err, c.wantErr)
		}
	}
}
