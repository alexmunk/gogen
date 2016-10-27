package rater

import (
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	luar "github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"
)

// ScriptRater executes a Lua Script to rate events
type ScriptRater struct {
	c *config.RaterConfig

	L        *lua.LState
	luaState *lua.LTable
}

// GetRate implements Rater interface
func (sr *ScriptRater) GetRate(now time.Time) float64 {
	if sr.luaState == nil {
		sr.luaState = new(lua.LTable)
	}
	L := lua.NewState()
	defer L.Close()
	L.SetGlobal("state", sr.luaState)
	L.SetGlobal("options", luar.New(L, sr.c.Options))
	if err := L.DoString(sr.c.Script); err != nil {
		log.Errorf("Error executing script for rater '%s': %s", sr.c.Name, err)
	}
	return float64(lua.LVAsNumber(L.Get(-1)))
}

// EventRate takes a given sample and current count and returns the rated count
func (sr *ScriptRater) EventRate(s *config.Sample, now time.Time, count int) int {
	return EventRate(s, now, count)
}

// TokenRate takes a token and returns the rated value
func (sr *ScriptRater) TokenRate(t config.Token, now time.Time) float64 {
	return TokenRate(t, now)
}
