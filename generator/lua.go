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

func (lg *luagen) replaceTokens(L *lua.LState) int {
	item := lg.currentItem

	luaEvent := L.ToTable(1)
	keepChoices := L.ToBool(1)
	event := make(map[string]string)
	if keepChoices {
		if item.S.LuaChoices == nil {
			item.S.LuaChoices = make(map[int]int)
		}
	} else {
		item.S.LuaChoices = nil
	}
	luaEvent.ForEach(func(k lua.LValue, v lua.LValue) {
		event[lua.LVAsString(k)] = lua.LVAsString(v)
	})
	replaceTokens(item, &event, &item.S.LuaChoices)
	retEvent := new(lua.LTable)
	for k, v := range event {
		retEvent.RawSetString(k, lua.LString(v))
	}
	L.Push(retEvent)
	return 1
}

func (lg *luagen) Gen(item *config.GenQueueItem) error {
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

	// Register functions
	L.SetGlobal("sleep", L.NewFunction(sleep))
	L.SetGlobal("debug", L.NewFunction(logdebug))
	L.SetGlobal("replaceTokens", L.NewFunction(lg.replaceTokens))

	// log.Debugf("Calling DoString for %# v", s.CustomGenerator.Script)
	if err := L.DoString(s.CustomGenerator.Script); err != nil {
		return fmt.Errorf("Error executing script for generator '%s': %s", s.CustomGenerator.Name, err)
	}
	// log.Debugf("Script returned")

	lv := L.Get(-1)
	events, err := lg.getEventsFromTable(lv)
	if err != nil {
		return err
	}
	lg.sendevents(events)
	return nil
}
