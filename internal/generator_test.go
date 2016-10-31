package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratorConfig(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "generator.yml"))

	c := NewConfig()

	g := getGenerator(c, "custom")
	assert.Equal(t, map[string]string{"foo": "bar"}, g.Options)
	assert.Equal(t, `return options["foo"] + state["somevar"]`+"\n", g.Script)
	assert.Equal(t, false, g.SingleThreaded)
}

func TestGeneratorFileConfig(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "generatorFile.yml"))

	c := NewConfig()

	g := getGenerator(c, "custom")
	assert.Equal(t, map[string]string{"foo": "bar"}, g.Options)
	assert.Equal(t, `return options["foo"] + state["somevar"]`, g.Script)
	assert.Equal(t, false, g.SingleThreaded)
}

func getGenerator(c *Config, name string) *GeneratorConfig {
	for _, g := range c.Generators {
		if g.Name == name {
			return g
		}
	}
	return nil
}
