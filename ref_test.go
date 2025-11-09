// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pin pointing go-swagger/go-swagger#1816 issue with cloning ref's
func TestCloneRef(t *testing.T) {
	var b bytes.Buffer
	src := MustCreateRef("#/definitions/test")
	require.NoError(t,
		gob.NewEncoder(&b).Encode(&src),
	)

	var dst Ref
	require.NoError(t,
		gob.NewDecoder(&b).Decode(&dst),
	)

	jazon, err := json.Marshal(dst)
	require.NoError(t, err)

	assert.JSONEq(t, `{"$ref":"#/definitions/test"}`, string(jazon))
}
