package config

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenReplacement(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "config", "tests", "tokens.yml"))
	loc, _ := time.LoadLocation("Local")
	source := rand.NewSource(0)
	randgen := rand.New(source)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	c := NewConfig()
	s := c.FindSampleByName("tokens")
	token := s.Tokens[0]

	choice := -1
	replacement, _ := token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "foo", replacement)

	choice = -1
	token = s.Tokens[1]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "4", replacement)

	choice = -1
	token = s.Tokens[2]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "0.514", replacement)

	choice = -1
	token = s.Tokens[3]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "NERA9rI2cv", replacement)

	choice = -1
	token = s.Tokens[4]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "56289", replacement)

	choice = -1
	token = s.Tokens[5]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "c", replacement)
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "c", replacement)

	token = s.Tokens[6]
	choices := make(map[int]int)
	for i := 0; i < 1000; i++ {
		choice = -1
		_, _ = token.GenReplacement(&choice, now(), now(), randgen)
		choices[choice] = choices[choice] + 1
	}
	if choices[0] != 312 || choices[1] != 572 || choices[2] != 116 {
		t.Fatalf("Choice distribution is off: %#v\n", choices)
	}

	token = s.Tokens[7]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "7", replacement)

	token = s.Tokens[8]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	fmt.Printf("UUID: %s\n", replacement)

	token = s.Tokens[9]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "184.226.113.189", replacement)

	token = s.Tokens[10]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "c9bb:42d4:abc1:7cea:9f7f:bbb2:caf4:a3ef", replacement)

	token = s.Tokens[11]
	replacement, _ = token.GenReplacement(&choice, now(), now(), randgen)
	assert.Equal(t, "2001-10-20 12:00:00,000", replacement)

	// choice = -1
	// token = s.Tokens[]

	// Try to call Exec first, should error
	// temp, err := Exec("test", row)
	// assert.EqualError(t, err, "Exec called for template 'test' but not found in cache")

	// // Create a new test template
	// err = New("test", "{{ ._raw }}")
	// temp, err = Exec("test", row)
	// assert.Equal(t, "foo", temp)

	// // More complicated
	// err = New("test2", "index={{ .index}} host={{ .host }} _raw={{ ._raw }}")
	// temp, err = Exec("test2", row)
	// assert.Equal(t, "index=fooindex host=barhost _raw=foo", temp)

	// // JSON
	// err = New("test3", "{{ json . | printf \"%s\" }}")
	// temp, err = Exec("test3", row)
	// assert.Equal(t, `{"_raw":"foo","host":"barhost","index":"fooindex"}`, temp)

	// // Multiple variables, one replacement
	// err = New("test4", "{{ ._raw }}{{ .foo }}")
	// temp, err = Exec("test4", row)
	// fmt.Printf("Test4: %s", temp)
}
