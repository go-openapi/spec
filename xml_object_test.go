// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

func TestXmlObject_Serialize(t *testing.T) {
	obj1 := XMLObject{}
	actual, err := json.Marshal(obj1)
	require.NoError(t, err)
	assert.Equal(t, "{}", string(actual))

	obj2 := XMLObject{
		Name:      "the name",
		Namespace: "the namespace",
		Prefix:    "the prefix",
		Attribute: true,
		Wrapped:   true,
	}

	actual, err = json.Marshal(obj2)
	require.NoError(t, err)

	var ad map[string]any
	require.NoError(t, json.Unmarshal(actual, &ad))
	assert.Equal(t, obj2.Name, ad["name"])
	assert.Equal(t, obj2.Namespace, ad["namespace"])
	assert.Equal(t, obj2.Prefix, ad["prefix"])
	assert.True(t, ad["attribute"].(bool))
	assert.True(t, ad["wrapped"].(bool))
}

func TestXmlObject_Deserialize(t *testing.T) {
	expected := XMLObject{}
	actual := XMLObject{}
	require.NoError(t, json.Unmarshal([]byte("{}"), &actual))
	assert.Equal(t, expected, actual)

	completed := `{"name":"the name","namespace":"the namespace","prefix":"the prefix","attribute":true,"wrapped":true}`
	expected = XMLObject{
		Name:      "the name",
		Namespace: "the namespace",
		Prefix:    "the prefix",
		Attribute: true,
		Wrapped:   true,
	}

	actual = XMLObject{}
	require.NoError(t, json.Unmarshal([]byte(completed), &actual))
	assert.Equal(t, expected, actual)
}
