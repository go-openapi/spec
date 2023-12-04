package spec

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoader_Issue145(t *testing.T) {
	t.Run("with ExpandSpec", func(t *testing.T) {
		basePath := filepath.Join("fixtures", "bugs", "145", "Program Files (x86)", "AppName", "todos.json")
		todosDoc, err := jsonDoc(basePath)
		require.NoError(t, err)

		spec := new(Swagger)
		require.NoError(t, json.Unmarshal(todosDoc, spec))

		require.NoError(t, ExpandSpec(spec, &ExpandOptions{RelativeBase: basePath}))
	})

	t.Run("with ExpandSchema", func(t *testing.T) {
		basePath := filepath.Join("fixtures", "bugs", "145", "Program Files (x86)", "AppName", "ref.json")
		schemaDoc, err := jsonDoc(basePath)
		require.NoError(t, err)

		sch := new(Schema)
		require.NoError(t, json.Unmarshal(schemaDoc, sch))

		require.NoError(t, ExpandSchema(sch, nil, nil))
	})
}
