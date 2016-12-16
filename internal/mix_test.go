package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMix1(t *testing.T) {
	// Setup environment
	ResetConfig()
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "mix", "mix1.yml"))
	os.Setenv("GOGEN_EXPORT", "")

	c := NewConfig()
	s1 := c.FindSampleByName("sample1")
	assert.Equal(t, 2, s1.Count)

	s2 := c.FindSampleByName("sample2")
	assert.Equal(t, 1, s2.Count)

	s3 := c.FindSampleByName("sample3")
	assert.Equal(t, "sample3", s3.Generator)
}

func TestMix2(t *testing.T) {
	// Setup environment
	ResetConfig()
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "mix", "mix2.yml"))

	c := NewConfig()
	s1 := c.FindSampleByName("sample1")
	assert.Equal(t, 1, s1.Count)

	s2 := c.FindSampleByName("sample2")
	assert.Equal(t, 2, s2.Interval)

	s3 := c.FindSampleByName("sample3")
	assert.Equal(t, 3, s3.EndIntervals)
}
