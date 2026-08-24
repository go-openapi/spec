// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/swag/conv"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

var schema = Schema{ //nolint:gochecknoglobals // test fixture
	VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-framework": "go-swagger"}},
	SchemaProps: SchemaProps{
		Ref:              MustCreateRef("Cat"),
		Type:             []string{"string"},
		Format:           "date",
		Description:      "the description of this schema",
		Title:            "the title",
		Default:          "blah",
		Maximum:          float64Ptr(100),
		ExclusiveMaximum: true,
		ExclusiveMinimum: true,
		Minimum:          float64Ptr(5),
		MaxLength:        int64Ptr(100),
		MinLength:        int64Ptr(5),
		Pattern:          "\\w{1,5}\\w+",
		MaxItems:         int64Ptr(100),
		MinItems:         int64Ptr(5),
		UniqueItems:      true,
		MultipleOf:       float64Ptr(5),
		Enum:             []any{"hello", "world"},
		MaxProperties:    int64Ptr(5),
		MinProperties:    int64Ptr(1),
		Required:         []string{"id", "name"},
		Items:            &SchemaOrArray{Schema: &Schema{SchemaProps: SchemaProps{Type: []string{"string"}}}},
		AllOf:            []Schema{{SchemaProps: SchemaProps{Type: []string{"string"}}}},
		Properties: map[string]Schema{
			"id":   {SchemaProps: SchemaProps{Type: []string{"integer"}, Format: "int64"}},
			"name": {SchemaProps: SchemaProps{Type: []string{"string"}}},
		},
		AdditionalProperties: &SchemaOrBool{Allows: true, Schema: &Schema{SchemaProps: SchemaProps{
			Type:   []string{"integer"},
			Format: "int32",
		}}},
	},
	SwaggerSchemaProps: SwaggerSchemaProps{
		Discriminator: "not this",
		ReadOnly:      true,
		XML:           &XMLObject{Name: "sch", Namespace: "io", Prefix: "sw", Attribute: true, Wrapped: true},
		ExternalDocs: &ExternalDocumentation{
			Description: "the documentation etc",
			URL:         "http://readthedocs.org/swagger",
		},
		Example: []any{
			map[string]any{
				"id":   1,
				"name": "a book",
			},
			map[string]any{
				"id":   2,
				"name": "the thing",
			},
		},
	},
}

//nolint:gochecknoglobals // test fixture
var schemaJSON = `{
	"x-framework": "go-swagger",
  "$ref": "Cat",
  "description": "the description of this schema",
  "maximum": 100,
  "minimum": 5,
  "exclusiveMaximum": true,
  "exclusiveMinimum": true,
  "maxLength": 100,
  "minLength": 5,
  "pattern": "\\w{1,5}\\w+",
  "maxItems": 100,
  "minItems": 5,
  "uniqueItems": true,
  "multipleOf": 5,
  "enum": ["hello", "world"],
  "type": "string",
  "format": "date",
  "title": "the title",
  "default": "blah",
  "maxProperties": 5,
  "minProperties": 1,
  "required": ["id", "name"],
  "items": {
    "type": "string"
  },
  "allOf": [
    {
      "type": "string"
    }
  ],
  "properties": {
    "id": {
      "type": "integer",
      "format": "int64"
    },
    "name": {
      "type": "string"
    }
  },
  "discriminator": "not this",
  "readOnly": true,
  "xml": {
    "name": "sch",
    "namespace": "io",
    "prefix": "sw",
    "wrapped": true,
    "attribute": true
  },
  "externalDocs": {
    "description": "the documentation etc",
    "url": "http://readthedocs.org/swagger"
  },
  "example": [
    {
      "id": 1,
      "name": "a book"
    },
    {
      "id": 2,
      "name": "the thing"
    }
  ],
  "additionalProperties": {
    "type": "integer",
    "format": "int32"
  }
}
`

func TestSchema(t *testing.T) {
	assert.JSONMarshalAsT(t, schemaJSON, schema)

	actual2 := Schema{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &actual2))

	assert.Equal(t, schema.Ref, actual2.Ref)
	assert.EqualT(t, schema.Description, actual2.Description)
	assert.Equal(t, schema.Maximum, actual2.Maximum)
	assert.Equal(t, schema.Minimum, actual2.Minimum)
	assert.EqualT(t, schema.ExclusiveMinimum, actual2.ExclusiveMinimum)
	assert.EqualT(t, schema.ExclusiveMaximum, actual2.ExclusiveMaximum)
	assert.Equal(t, schema.MaxLength, actual2.MaxLength)
	assert.Equal(t, schema.MinLength, actual2.MinLength)
	assert.EqualT(t, schema.Pattern, actual2.Pattern)
	assert.Equal(t, schema.MaxItems, actual2.MaxItems)
	assert.Equal(t, schema.MinItems, actual2.MinItems)
	assert.TrueT(t, actual2.UniqueItems)
	assert.Equal(t, schema.MultipleOf, actual2.MultipleOf)
	assert.Equal(t, schema.Enum, actual2.Enum)
	assert.Equal(t, schema.Type, actual2.Type)
	assert.EqualT(t, schema.Format, actual2.Format)
	assert.EqualT(t, schema.Title, actual2.Title)
	assert.Equal(t, schema.MaxProperties, actual2.MaxProperties)
	assert.Equal(t, schema.MinProperties, actual2.MinProperties)
	assert.Equal(t, schema.Required, actual2.Required)
	assert.Equal(t, schema.Items, actual2.Items)
	assert.Equal(t, schema.AllOf, actual2.AllOf)
	assert.Equal(t, schema.Properties, actual2.Properties)
	assert.EqualT(t, schema.Discriminator, actual2.Discriminator)
	assert.EqualT(t, schema.ReadOnly, actual2.ReadOnly)
	assert.Equal(t, schema.XML, actual2.XML)
	assert.Equal(t, schema.ExternalDocs, actual2.ExternalDocs)
	assert.Equal(t, schema.AdditionalProperties, actual2.AdditionalProperties)
	assert.Equal(t, schema.Extensions, actual2.Extensions)

	examples, ok := actual2.Example.([]any)
	require.TrueT(t, ok, "expected []any for actual2.Example")
	expEx, ok := schema.Example.([]any)
	require.TrueT(t, ok, "expected []any for schema.Example")
	ex1, ok := examples[0].(map[string]any)
	require.TrueT(t, ok, "expected map[string]any for examples[0]")
	ex2, ok := examples[1].(map[string]any)
	require.TrueT(t, ok, "expected map[string]any for examples[1]")
	exp1, ok := expEx[0].(map[string]any)
	require.TrueT(t, ok, "expected map[string]any for expEx[0]")
	exp2, ok := expEx[1].(map[string]any)
	require.TrueT(t, ok, "expected map[string]any for expEx[1]")

	assert.EqualValues(t, exp1["id"], ex1["id"])
	assert.Equal(t, exp1["name"], ex1["name"])
	assert.EqualValues(t, exp2["id"], ex2["id"])
	assert.Equal(t, exp2["name"], ex2["name"])
}

func BenchmarkSchemaUnmarshal(b *testing.B) {
	for b.Loop() {
		sch := &Schema{}
		_ = sch.UnmarshalJSON([]byte(schemaJSON))
	}
}

func TestSchemaWithValidation(t *testing.T) {
	s := new(Schema).WithValidations(SchemaValidations{CommonValidations: CommonValidations{MaxLength: conv.Pointer(int64(15))}})
	assert.Equal(t, conv.Pointer(int64(15)), s.MaxLength)

	val := mkVal()
	s = new(Schema).WithValidations(val)

	assert.Equal(t, val, s.Validations())
}

func TestSchemaPropertyConstructors(t *testing.T) {
	t.Run("aliases yield the same schema as their canonical constructor", func(t *testing.T) {
		assert.Equal(t, BooleanProperty(), BoolProperty())
		assert.Equal(t, StringProperty(), CharProperty())
	})

	t.Run("sized integer properties carry their format", func(t *testing.T) {
		assert.EqualT(t, "int8", Int8Property().Format)
		assert.EqualT(t, "int16", Int16Property().Format)
		assert.Equal(t, StringOrArray{"integer"}, Int8Property().Type)
		assert.Equal(t, StringOrArray{"integer"}, Int16Property().Type)
	})

	t.Run("StrFmtProperty names the string format", func(t *testing.T) {
		s := StrFmtProperty("uuid")
		assert.Equal(t, StringOrArray{"string"}, s.Type)
		assert.EqualT(t, "uuid", s.Format)
	})

	t.Run("ArrayProperty without items is an untyped array", func(t *testing.T) {
		s := ArrayProperty(nil)
		assert.Equal(t, StringOrArray{"array"}, s.Type)
		assert.Nil(t, s.Items)
	})

	t.Run("ComposedSchema collects its members in allOf", func(t *testing.T) {
		s := ComposedSchema(*StringProperty(), *Int64Property())
		assert.Len(t, s.AllOf, 2)
		assert.Equal(t, *StringProperty(), s.AllOf[0])
		assert.Equal(t, *Int64Property(), s.AllOf[1])
	})
}

func TestSchemaURLUnmarshal(t *testing.T) {
	t.Run("should unmarshal a $schema member", func(t *testing.T) {
		var u SchemaURL
		require.NoError(t, json.Unmarshal([]byte(`{"$schema":"http://json-schema.org/draft-04/schema#"}`), &u))
		// the empty fragment is dropped when the URL is reassembled
		assert.EqualT(t, SchemaURL("http://json-schema.org/draft-04/schema"), u)
	})

	t.Run("should leave the URL empty when $schema is absent", func(t *testing.T) {
		var u SchemaURL
		require.NoError(t, json.Unmarshal([]byte(`{"other":"value"}`), &u))
		assert.EqualT(t, SchemaURL(""), u)
	})

	t.Run("should ignore a non-string $schema", func(t *testing.T) {
		var u SchemaURL
		require.NoError(t, json.Unmarshal([]byte(`{"$schema":12}`), &u))
		assert.EqualT(t, SchemaURL(""), u)
	})

	t.Run("should error on invalid JSON", func(t *testing.T) {
		var u SchemaURL
		require.Error(t, json.Unmarshal([]byte(`not json`), &u))
	})

	t.Run("should error on an unparseable URL", func(t *testing.T) {
		var u SchemaURL
		require.Error(t, json.Unmarshal([]byte(`{"$schema":"http://["}`), &u))
	})
}

func TestSchemaBuilder(t *testing.T) {
	t.Run("identity and documentation", func(t *testing.T) {
		s := new(Schema).
			WithID("the id").
			WithTitle("the title").
			WithDescription("the description").
			WithExample("an example").
			WithDiscriminator("kind")

		assert.EqualT(t, "the id", s.ID)
		assert.EqualT(t, "the title", s.Title)
		assert.EqualT(t, "the description", s.Description)
		assert.Equal(t, "an example", s.Example)
		assert.EqualT(t, "kind", s.Discriminator)
	})

	t.Run("external docs are removed when both params are empty", func(t *testing.T) {
		s := new(Schema).WithExternalDocs("the docs", "http://readthedocs.org/swagger")
		require.NotNil(t, s.ExternalDocs)
		assert.EqualT(t, "the docs", s.ExternalDocs.Description)
		assert.EqualT(t, "http://readthedocs.org/swagger", s.ExternalDocs.URL)

		s = s.WithExternalDocs("", "")
		assert.Nil(t, s.ExternalDocs)
	})

	t.Run("properties", func(t *testing.T) {
		s := new(Schema).WithProperties(map[string]Schema{"name": *StringProperty()})
		require.Len(t, s.Properties, 1)

		s = s.SetProperty("age", *Int32Property())
		require.Len(t, s.Properties, 2)
		assert.Equal(t, *Int32Property(), s.Properties["age"])

		// SetProperty allocates the map when the schema has none
		other := new(Schema).SetProperty("id", *Int64Property())
		require.Len(t, other.Properties, 1)

		s = s.WithMaxProperties(10).WithMinProperties(1)
		assert.Equal(t, int64Ptr(10), s.MaxProperties)
		assert.Equal(t, int64Ptr(1), s.MinProperties)
	})

	t.Run("required", func(t *testing.T) {
		s := new(Schema).WithRequired("id", "name")
		assert.Equal(t, []string{"id", "name"}, s.Required)

		s = s.AddRequired("age")
		assert.Equal(t, []string{"id", "name", "age"}, s.Required)
	})

	t.Run("allOf", func(t *testing.T) {
		s := new(Schema).WithAllOf(*StringProperty())
		require.Len(t, s.AllOf, 1)

		s = s.AddToAllOf(*Int64Property(), *BooleanProperty())
		assert.Len(t, s.AllOf, 3)
	})

	t.Run("types and items", func(t *testing.T) {
		s := new(Schema).Typed("integer", "int32").AddType("null", "")
		assert.Equal(t, StringOrArray{"integer", "null"}, s.Type)
		assert.EqualT(t, "int32", s.Format)

		s = s.AddType("string", "date")
		assert.EqualT(t, "date", s.Format)

		s = new(Schema).AsNullable()
		assert.TrueT(t, s.Nullable)

		s = new(Schema).CollectionOf(*StringProperty())
		assert.Equal(t, StringOrArray{"array"}, s.Type)
		require.NotNil(t, s.Items)
		assert.Equal(t, StringProperty(), s.Items.Schema)
	})

	t.Run("validations", func(t *testing.T) {
		s := new(Schema).
			WithDefault("a default").
			WithMaxLength(100).
			WithMinLength(5).
			WithPattern(`\w+`).
			WithMultipleOf(5).
			WithMaximum(100, true).
			WithMinimum(5, true).
			WithEnum("hello", "world").
			WithMaxItems(100).
			WithMinItems(5).
			UniqueValues()

		assert.Equal(t, "a default", s.Default)
		assert.Equal(t, int64Ptr(100), s.MaxLength)
		assert.Equal(t, int64Ptr(5), s.MinLength)
		assert.EqualT(t, `\w+`, s.Pattern)
		assert.Equal(t, float64Ptr(5), s.MultipleOf)
		assert.Equal(t, float64Ptr(100), s.Maximum)
		assert.TrueT(t, s.ExclusiveMaximum)
		assert.Equal(t, float64Ptr(5), s.Minimum)
		assert.TrueT(t, s.ExclusiveMinimum)
		assert.Equal(t, []any{"hello", "world"}, s.Enum)
		assert.Equal(t, int64Ptr(100), s.MaxItems)
		assert.Equal(t, int64Ptr(5), s.MinItems)
		assert.TrueT(t, s.UniqueItems)

		s = s.AllowDuplicates().WithMaximum(100, false).WithMinimum(5, false)
		assert.FalseT(t, s.UniqueItems)
		assert.FalseT(t, s.ExclusiveMaximum)
		assert.FalseT(t, s.ExclusiveMinimum)
	})

	t.Run("readOnly", func(t *testing.T) {
		s := new(Schema).AsReadOnly()
		assert.TrueT(t, s.ReadOnly)

		s = s.AsWritable()
		assert.FalseT(t, s.ReadOnly)
	})

	t.Run("XML settings allocate the XMLObject on first use", func(t *testing.T) {
		s := new(Schema).
			WithXMLName("the name").
			WithXMLNamespace("the namespace").
			WithXMLPrefix("the prefix").
			AsXMLAttribute().
			AsWrappedXML()

		require.NotNil(t, s.XML)
		assert.EqualT(t, "the name", s.XML.Name)
		assert.EqualT(t, "the namespace", s.XML.Namespace)
		assert.EqualT(t, "the prefix", s.XML.Prefix)
		assert.TrueT(t, s.XML.Attribute)
		assert.TrueT(t, s.XML.Wrapped)

		s = s.AsXMLElement().AsUnwrappedXML()
		assert.FalseT(t, s.XML.Attribute)
		assert.FalseT(t, s.XML.Wrapped)

		// each setter allocates on its own when XML is still nil
		assert.NotNil(t, new(Schema).WithXMLNamespace("ns").XML)
		assert.NotNil(t, new(Schema).WithXMLPrefix("p").XML)
		assert.NotNil(t, new(Schema).AsXMLAttribute().XML)
		assert.NotNil(t, new(Schema).AsXMLElement().XML)
		assert.NotNil(t, new(Schema).AsWrappedXML().XML)
		assert.NotNil(t, new(Schema).AsUnwrappedXML().XML)
	})
}
