// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/go-openapi/swag/loading"
	"github.com/go-openapi/testify/v2/require"
)

var errBlockedAddress = errors.New("blocked non-public address")

// restrictedDialContext refuses to dial loopback, private, link-local or unspecified addresses.
// This mirrors the SSRF guard a caller injects via loading.WithHTTPClient (and that
// go-openapi/loads ships as RestrictedHTTPClient).
func restrictedDialContext(_ context.Context, _, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		// a hostname would need resolution then a re-check; this test only uses IP literals.
		return nil, fmt.Errorf("%w: %s", errBlockedAddress, addr)
	}

	ip = ip.Unmap()
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
		return nil, fmt.Errorf("%w: %s", errBlockedAddress, addr)
	}

	return nil, fmt.Errorf("%w: %s (test performs no real dial)", errBlockedAddress, addr)
}

// TestExpand_SSRFPosture validates that a caller can neutralize the SSRF vector by injecting an
// option-aware loader bound to a restricted HTTP client through PathLoaderWithOptions: a remote
// "$ref" to a cloud metadata endpoint is refused at dial time, before any connection is made.
//
// The loader selection is shared by every expansion/resolution entry point, so blocking it here
// blocks it for ExpandSpec, ExpandSchemaWithBasePath, ExpandResponse, ExpandParameter and the
// Resolve* functions alike.
func TestExpand_SSRFPosture(t *testing.T) {
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DialContext: restrictedDialContext},
	}
	loader := func(pth string, _ ...loading.Option) (json.RawMessage, error) {
		b, err := loading.LoadFromFileOrHTTP(pth, loading.WithHTTPClient(client))
		return json.RawMessage(b), err
	}

	// AWS IMDS endpoint, exactly as in the report's PoC
	raw := `{
		"swagger":"2.0","info":{"title":"x","version":"1"},"paths":{},
		"definitions":{
			"Victim":{"$ref":"http://169.254.169.254/latest/meta-data/iam/security-credentials/role"}
		}
	}`
	var sw Swagger
	require.NoError(t, json.Unmarshal([]byte(raw), &sw))

	err := ExpandSpec(&sw, &ExpandOptions{PathLoaderWithOptions: loader})

	// the metadata endpoint was refused at dial time: the fetch never happened
	require.Error(t, err)
	require.ErrorIs(t, err, errBlockedAddress)
	require.ErrorContains(t, err, "169.254.169.254")
}
