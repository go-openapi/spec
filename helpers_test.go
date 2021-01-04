package spec

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var rex = regexp.MustCompile(`"\$ref":\s*"(.*)"`)

func jsonDoc(path string) (json.RawMessage, error) {
	data, err := swag.LoadFromFileOrHTTP(path)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func docAndOpts(t testing.TB, fixturePath string) ([]byte, *ExpandOptions) {
	doc, err := jsonDoc(fixturePath)
	require.NoError(t, err)

	specPath, _ := absPath(fixturePath)

	return doc, &ExpandOptions{
		RelativeBase: specPath,
	}
}

func expandThisSchemaOrDieTrying(t testing.TB, fixturePath string) string {
	doc, opts := docAndOpts(t, fixturePath)

	sch := new(Schema)
	require.NoError(t, json.Unmarshal(doc, sch))

	require.NotPanics(t, func() {
		require.NoError(t, ExpandSchemaWithBasePath(sch, nil, opts))
	}, "calling expand schema circular refs, should not panic!")

	bbb, err := json.MarshalIndent(sch, "", " ")
	require.NoError(t, err)

	return string(bbb)
}

func expandThisOrDieTrying(t testing.TB, fixturePath string) string {
	doc, opts := docAndOpts(t, fixturePath)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(doc, spec))

	require.NotPanics(t, func() {
		require.NoError(t, ExpandSpec(spec, opts))
	}, "calling expand spec with circular refs, should not panic!")

	bbb, err := json.MarshalIndent(spec, "", " ")
	require.NoError(t, err)

	return string(bbb)
}

func assertRefInJSON(t testing.TB, jazon, prefix string) {
	// assert a match in a references
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)

	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, prefix),
			"expected $ref to match %q, got: %s", prefix, matched[0])
	}
}

func assertRefInJSONRegexp(t testing.TB, jazon, match string) {
	// assert a match in a references
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)

	refMatch, err := regexp.Compile(match)
	require.NoError(t, err)

	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, refMatch.MatchString(subMatch),
			"expected $ref to match %q, got: %s", match, matched[0])
	}
}
