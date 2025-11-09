// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/swag/conv"
	"github.com/stretchr/testify/require"
)

func mkVal() SchemaValidations {
	return SchemaValidations{
		CommonValidations: CommonValidations{
			Maximum:          conv.Pointer(2.5),
			ExclusiveMaximum: true,
			Minimum:          conv.Pointer(3.4),
			ExclusiveMinimum: true,
			MaxLength:        conv.Pointer(int64(15)),
			MinLength:        conv.Pointer(int64(16)),
			Pattern:          "abc",
			MaxItems:         conv.Pointer(int64(17)),
			MinItems:         conv.Pointer(int64(18)),
			UniqueItems:      true,
			MultipleOf:       conv.Pointer(4.4),
			Enum:             []any{"a", 12.5},
		},
		PatternProperties: SchemaProperties{
			"x": *BooleanProperty(),
			"y": *BooleanProperty(),
		},
		MinProperties: conv.Pointer(int64(19)),
		MaxProperties: conv.Pointer(int64(20)),
	}
}

func TestValidations(t *testing.T) {

	var cv CommonValidations
	val := mkVal()
	cv.SetValidations(val)

	expectedCV := val.CommonValidations
	require.Equal(t, expectedCV, cv)

	require.True(t, cv.HasArrayValidations())
	require.True(t, cv.HasNumberValidations())
	require.True(t, cv.HasStringValidations())
	require.True(t, cv.HasEnum())

	cv.Enum = nil
	require.False(t, cv.HasEnum())

	cv.MaxLength = nil
	require.True(t, cv.HasStringValidations())
	cv.MinLength = nil
	require.True(t, cv.HasStringValidations())
	cv.Pattern = ""
	require.False(t, cv.HasStringValidations())

	cv.Minimum = nil
	require.True(t, cv.HasNumberValidations())
	cv.Maximum = nil
	require.True(t, cv.HasNumberValidations())
	cv.MultipleOf = nil
	require.False(t, cv.HasNumberValidations())

	cv.MaxItems = nil
	require.True(t, cv.HasArrayValidations())
	cv.MinItems = nil
	require.True(t, cv.HasArrayValidations())
	cv.UniqueItems = false
	require.False(t, cv.HasArrayValidations())

	val = mkVal()
	expectedSV := val
	expectedSV.PatternProperties = nil
	expectedSV.MinProperties = nil
	expectedSV.MaxProperties = nil

	val = mkVal()
	cv.SetValidations(val)
	require.Equal(t, expectedSV, cv.Validations())

	var sv SchemaValidations
	val = mkVal()
	sv.SetValidations(val)

	expectedSV = val
	require.Equal(t, expectedSV, sv)

	require.Equal(t, val, sv.Validations())

	require.True(t, sv.HasObjectValidations())
	sv.MinProperties = nil
	require.True(t, sv.HasObjectValidations())
	sv.MaxProperties = nil
	require.True(t, sv.HasObjectValidations())
	sv.PatternProperties = nil
	require.False(t, sv.HasObjectValidations())

	val = mkVal()
	cv.SetValidations(val)
	cv.ClearStringValidations()
	require.False(t, cv.HasStringValidations())

	cv.ClearNumberValidations()
	require.False(t, cv.HasNumberValidations())

	cv.ClearArrayValidations()
	require.False(t, cv.HasArrayValidations())

	sv.SetValidations(val)
	sv.ClearObjectValidations(func(validation string, _ any) {
		switch validation {
		case "minProperties", "maxProperties", "patternProperties":
			return
		default:
			t.Logf("unexpected validation %s", validation)
			t.Fail()
		}
	})
	require.Falsef(t, sv.HasObjectValidations(), "%#v", sv)
}
