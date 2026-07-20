// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// errNoExternalLoads is returned by the deny-all PathLoader used to prove that refusing
// external loads is not, on its own, a mitigation for the amplification attack.
var errNoExternalLoads = errors.New("no external loads allowed")

// buildAmplificationSpec builds a self-contained spec where each of n definitions references
// the next one twice via allOf. Without an expansion budget, expanding d0 inlines a tree with
// 2^(n-1) leaves from O(n) bytes of input (a $ref amplification / "billion laughs" attack).
func buildAmplificationSpec(t testing.TB, n int) []byte {
	t.Helper()

	defs := make(map[string]any, n)
	for i := range n {
		var sch any
		if i == n-1 {
			sch = map[string]any{"type": "string"}
		} else {
			next := fmt.Sprintf("#/definitions/d%d", i+1)
			sch = map[string]any{"allOf": []any{
				map[string]any{"$ref": next},
				map[string]any{"$ref": next},
			}}
		}
		defs[fmt.Sprintf("d%d", i)] = sch
	}

	doc := map[string]any{
		"swagger":     "2.0",
		"info":        map[string]any{"title": "x", "version": "1"},
		"paths":       map[string]any{},
		"definitions": defs,
	}
	raw, err := json.Marshal(doc)
	require.NoError(t, err)

	return raw
}

func TestMaxExpansionNodesTriState(t *testing.T) {
	// 0 (zero value): default budget, so every caller is protected out of the box.
	assert.EqualT(t, DefaultMaxExpansionNodes, (&ExpandOptions{}).maxExpansionNodes())

	// negative: unbounded.
	assert.EqualT(t, 0, (&ExpandOptions{MaxExpansionNodes: -1}).maxExpansionNodes())

	// positive: explicit budget.
	assert.EqualT(t, 1234, (&ExpandOptions{MaxExpansionNodes: 1234}).maxExpansionNodes())
}

func TestExpand_AmplificationBudget(t *testing.T) {
	// A deep amplification spec. Even a modest depth would explode without a budget.
	const depth = 40
	raw := buildAmplificationSpec(t, depth)

	t.Run("explicit budget trips the guard", func(t *testing.T) {
		var sw Swagger
		require.NoError(t, json.Unmarshal(raw, &sw))

		err := ExpandSpec(&sw, &ExpandOptions{MaxExpansionNodes: 2000})
		require.ErrorIs(t, err, ErrExpandTooManyNodes)
	})

	t.Run("deny-all PathLoader is not a mitigation on its own", func(t *testing.T) {
		// All refs are fragment-only and resolve against the in-memory root, so refusing
		// external loads does not prevent the blow-up: the budget is what stops it.
		var sw Swagger
		require.NoError(t, json.Unmarshal(raw, &sw))

		loaderCalled := false
		err := ExpandSpec(&sw, &ExpandOptions{
			MaxExpansionNodes: 2000,
			PathLoader: func(p string) (json.RawMessage, error) {
				loaderCalled = true
				return nil, fmt.Errorf("%w: %s", errNoExternalLoads, p)
			},
		})
		require.ErrorIs(t, err, ErrExpandTooManyNodes)
		assert.FalseT(t, loaderCalled, "expected no external load attempts")
	})

	t.Run("ContinueOnError does not suppress a budget breach", func(t *testing.T) {
		// The budget is a hard resource-exhaustion safeguard: unlike an unresolvable $ref,
		// it must surface even when the caller tolerates errors.
		var sw Swagger
		require.NoError(t, json.Unmarshal(raw, &sw))

		err := ExpandSpec(&sw, &ExpandOptions{MaxExpansionNodes: 2000, ContinueOnError: true})
		require.ErrorIs(t, err, ErrExpandTooManyNodes)
	})

	t.Run("negative budget disables the guard", func(t *testing.T) {
		// A shallow spec that stays well under any real budget must expand fully when unbounded.
		shallow := buildAmplificationSpec(t, 8)
		var sw Swagger
		require.NoError(t, json.Unmarshal(shallow, &sw))

		require.NoError(t, ExpandSpec(&sw, &ExpandOptions{MaxExpansionNodes: -1}))
	})
}

func TestExpand_BudgetAllowsLegitSpec(t *testing.T) {
	// A shallow amplification spec (small node count) must expand cleanly under the default budget.
	raw := buildAmplificationSpec(t, 8)
	var sw Swagger
	require.NoError(t, json.Unmarshal(raw, &sw))

	require.NoError(t, ExpandSpec(&sw, nil)) // nil options => default budget
	// d0 fully expanded: no $ref remains in the leaf chain.
	out, err := json.Marshal(sw.Definitions["d0"])
	require.NoError(t, err)
	assert.StringNotContainsT(t, string(out), `"$ref"`)
}

func TestExpand_BudgetErrorIsSentinel(t *testing.T) {
	raw := buildAmplificationSpec(t, 40)
	var sw Swagger
	require.NoError(t, json.Unmarshal(raw, &sw))

	err := ExpandSpec(&sw, &ExpandOptions{MaxExpansionNodes: 100})
	require.Error(t, err)
	assert.TrueT(t, errors.Is(err, ErrExpandTooManyNodes))
}
