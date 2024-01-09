// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"bytes"
	"encoding/gob"
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

	assert.Equal(t, `{"$ref":"#/definitions/test"}`, string(jazon))
}
