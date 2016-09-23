package template

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	row := map[string]string{"_raw": "foo", "index": "fooindex", "host": "barhost"}

	// Try to call Exec first, should error
	temp, err := Exec("test", row)
	assert.EqualError(t, err, "Exec called for template 'test' but not found in cache")

	// Create a new test template
	err = New("test", "{{ ._raw }}")
	temp, err = Exec("test", row)
	assert.Equal(t, "foo", temp)

	// More complicated
	err = New("test2", "index={{ .index}} host={{ .host }} _raw={{ ._raw }}")
	temp, err = Exec("test2", row)
	assert.Equal(t, "index=fooindex host=barhost _raw=foo", temp)

	// JSON
	err = New("test3", "{{ json . | printf \"%s\" }}")
	temp, err = Exec("test3", row)
	assert.Equal(t, `{"_raw":"foo","host":"barhost","index":"fooindex"}`, temp)

	// Multiple variables, one replacement
	err = New("test4", "{{ ._raw }}{{ .foo }}")
	temp, err = Exec("test4", row)
	fmt.Printf("Test4: %s", temp)
}
