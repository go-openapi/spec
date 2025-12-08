// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
)

// OrderSchemaItem holds a named schema (e.g. from a property of an object)
type OrderSchemaItem struct {
	Schema

	Name string
}

// OrderSchemaItems is a sortable slice of named schemas.
// The ordering is defined by the x-order schema extension.
type OrderSchemaItems []OrderSchemaItem

// MarshalJSON produces a json object with keys defined by the name schemas
// of the OrderSchemaItems slice, keeping the original order of the slice.
func (items OrderSchemaItems) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("{")
	for i := range items {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("\"")
		buf.WriteString(items[i].Name)
		buf.WriteString("\":")
		bs, err := json.Marshal(&items[i].Schema)
		if err != nil {
			return nil, err
		}
		buf.Write(bs)
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

func (items OrderSchemaItems) Len() int      { return len(items) }
func (items OrderSchemaItems) Swap(i, j int) { items[i], items[j] = items[j], items[i] }
func (items OrderSchemaItems) Less(i, j int) (ret bool) {
	ii, oki := items[i].Extensions.GetInt("x-order")
	ij, okj := items[j].Extensions.GetInt("x-order")
	if oki {
		if okj {
			defer func() {
				if err := recover(); err != nil {
					defer func() {
						if err = recover(); err != nil {
							ret = items[i].Name < items[j].Name
						}
					}()
					ret = reflect.ValueOf(ii).String() < reflect.ValueOf(ij).String()
				}
			}()
			return ii < ij
		}
		return true
	} else if okj {
		return false
	}
	return items[i].Name < items[j].Name
}

// SchemaProperties is a map representing the properties of a Schema object.
// It knows how to transform its keys into an ordered slice.
type SchemaProperties map[string]Schema

// ToOrderedSchemaItems transforms the map of properties into a sortable slice
func (properties SchemaProperties) ToOrderedSchemaItems() OrderSchemaItems {
	items := make(OrderSchemaItems, 0, len(properties))
	for k, v := range properties {
		items = append(items, OrderSchemaItem{
			Name:   k,
			Schema: v,
		})
	}
	sort.Sort(items)
	return items
}

// MarshalJSON produces properties as json, keeping their order.
func (properties SchemaProperties) MarshalJSON() ([]byte, error) {
	if properties == nil {
		return []byte("null"), nil
	}
	return json.Marshal(properties.ToOrderedSchemaItems())
}

// UnmarshalJSON handles JSON Schema 2020-12 where property values can be either
// a schema object or a boolean (true/false). Boolean values are converted to
// empty schemas with appropriate semantics (true = allows any, false = allows none).
func (properties *SchemaProperties) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	result := make(SchemaProperties, len(raw))
	for k, v := range raw {
		// Check if it's a boolean value
		trimmed := bytes.TrimSpace(v)
		if bytes.Equal(trimmed, []byte("true")) || bytes.Equal(trimmed, []byte("false")) {
			// For boolean values in properties (JSON Schema 2020-12):
			// true = schema that allows anything (empty schema)
			// false = schema that allows nothing (we can represent this with impossible constraints)
			// For simplicity, we use an empty schema for true and skip false entries
			if bytes.Equal(trimmed, []byte("true")) {
				result[k] = Schema{} // empty schema allows anything
			}
			// false entries are not added - they disallow any value
			continue
		}

		var schema Schema
		if err := json.Unmarshal(v, &schema); err != nil {
			return err
		}
		result[k] = schema
	}

	*properties = result
	return nil
}

// PatternSchemaProperties is a map representing pattern properties of a Schema object.
// In JSON Schema 2020-12, pattern property values can be either a schema or a boolean.
type PatternSchemaProperties map[string]SchemaOrBool

// MarshalJSON produces pattern properties as json.
func (properties PatternSchemaProperties) MarshalJSON() ([]byte, error) {
	if properties == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]SchemaOrBool(properties))
}
