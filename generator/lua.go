package generator

import (
	"fmt"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	luar "github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"
)

type luagen struct {
	initialized bool
	currentItem *config.GenQueueItem
	tokens      []config.Token
}

func sleep(L *lua.LState) int {
	lv := L.ToInt64(1)
	time.Sleep(time.Duration(lv))
	return 0
}

func logdebug(L *lua.LState) int {
	lv := L.ToString(1)
	log.Debug(lv)
	return 0
}

func (lg *luagen) send(L *lua.LState) int {
	lv := L.ToTable(1)
	events, err := lg.getEventsFromTable(lv)
	if err != nil {
		log.Errorf("Received error from generator '%s': %s", lg.currentItem.S.CustomGenerator.Name, err)
		return 0
	}
	lg.sendevents(events)
	return 0
}

func (lg *luagen) sendevents(events []map[string]string) {
	item := lg.currentItem
	// log.Debugf("events: %# v", pretty.Formatter(events))
	outitem := &config.OutQueueItem{S: item.S, Events: events}
	item.OQ <- outitem
}

func (lg *luagen) getEventsFromTable(lv lua.LValue) ([]map[string]string, error) {
	s := lg.currentItem.S
	var err error
	var events []map[string]string
	if lv, ok := lv.(*lua.LTable); ok {
		events = make([]map[string]string, 0, lv.Len())
		lv.ForEach(func(k lua.LValue, v lua.LValue) {
			event := make(map[string]string)
			if castv, ok := v.(*lua.LTable); ok {
				castv.ForEach(func(k2 lua.LValue, v2 lua.LValue) {
					event[lua.LVAsString(k2)] = lua.LVAsString(v2)
					// log.Debugf("Event created: %s->%s", lua.LVAsString(k2), lua.LVAsString(v2))
				})
				events = append(events, event)
			} else {
				err = fmt.Errorf("Value of a returned row is not a LUA Table for sample '%s' with generator '%s', instead got: %s", s.Name, s.CustomGenerator.Name, v.Type())
			}
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Returned value from generator '%s' in sample '%s' is not a Lua Table", s.CustomGenerator.Name, s.Name)
	}
	return events, nil
}

func (lg *luagen) setToken(L *lua.LState) int {
	top := L.GetTop()
	tokenName := L.ToString(1)
	tokenValue := L.ToString(2)
	var tokenField string
	if top > 2 {
		tokenField = L.ToString(3)
	} else {
		tokenField = config.DefaultField
	}

	found := false
	for i := 0; i < len(lg.tokens); i++ {
		if lg.tokens[i].Name == tokenName {
			lg.tokens[i].Replacement = tokenValue
			found = true
		}
	}
	if !found {
		var t config.Token
		t.Name = tokenName
		t.Type = "static"
		t.Format = "template"
		t.Token = "$" + tokenName + "$"
		t.Replacement = tokenValue
		t.Field = tokenField
		lg.tokens = append(lg.tokens, t)
		log.Debugf("Added token")
	}

	return 0
}

func (lg *luagen) replaceTokens(L *lua.LState) int {
	item := lg.currentItem

	// Get number of args and events map
	top := L.GetTop()

	// Get events map
	luaEvent := L.ToTable(1)
	event := make(map[string]string)
	luaEvent.ForEach(func(k lua.LValue, v lua.LValue) {
		event[lua.LVAsString(k)] = lua.LVAsString(v)
	})

	// Get choices from args, or if omitted create a new map
	var choices map[int]int
	var ok bool
	if top > 1 {
		ud := L.CheckUserData(2)
		if choices, ok = ud.Value.(map[int]int); !ok {
			L.ArgError(2, "expecting choices map[int]int")
			return 0
		}
	} else {
		choices = make(map[int]int)
	}

	// Replace configured tokens
	replaceTokens(item, &event, &choices, item.S.Tokens)

	// Replace any tokens submitted through setTokens
	if len(lg.tokens) > 0 {
		throwawayChoices := make(map[int]int)
		replaceTokens(item, &event, &throwawayChoices, lg.tokens)
	}

	// Return a table of the event created from our map and a userdata of the choices map[int]int
	retEvent := new(lua.LTable)
	for k, v := range event {
		retEvent.RawSetString(k, lua.LString(v))
	}
	L.Push(retEvent)
	L.Push(luar.New(L, choices))
	return 2
}

func (lg *luagen) Gen(item *config.GenQueueItem) error {
	if !lg.initialized {
		lg.tokens = make([]config.Token, 0)
		lg.initialized = true
	}
	// log.Debugf("Lua Gen called for sample '%s'", item.S.Name)
	L := lua.NewState()
	defer L.Close()
	s := item.S
	s.LuaMutex.Lock()
	defer s.LuaMutex.Unlock()
	lg.currentItem = item

	// Register global variables
	L.SetGlobal("state", s.LuaState)
	L.SetGlobal("options", luar.New(L, s.CustomGenerator.Options))
	L.SetGlobal("lines", s.LuaLines)
	L.SetGlobal("count", luar.New(L, item.Count))
	L.SetGlobal("earliest", luar.New(L, item.Earliest))
	L.SetGlobal("latest", luar.New(L, item.Latest))
	L.SetGlobal("now", luar.New(L, item.Now))

	// Register functions
	L.SetGlobal("sleep", L.NewFunction(sleep))
	L.SetGlobal("debug", L.NewFunction(logdebug))
	L.SetGlobal("replaceTokens", L.NewFunction(lg.replaceTokens))
	L.SetGlobal("send", L.NewFunction(lg.send))
	L.SetGlobal("setToken", L.NewFunction(lg.setToken))

	// log.Debugf("Calling DoString for %# v", s.CustomGenerator.Script)
	if err := L.DoString(s.CustomGenerator.Script); err != nil {
		return fmt.Errorf("Error executing script for generator '%s': %s", s.CustomGenerator.Name, err)
	}
	// log.Debugf("Script returned")

	// lv := L.Get(-1)
	// events, err := lg.getEventsFromTable(lv)
	// if err != nil {
	// 	return err
	// }
	// lg.sendevents(events)
	return nil
}
