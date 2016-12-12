package generator

import (
	"fmt"
	"math"
	"sync"
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
	states      map[string]*config.GeneratorState
	code        map[string]*lua.LFunction
	lstates     map[string]*sync.Pool
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

func (lg *luagen) round(L *lua.LState) int {
	var ret float64
	var lret lua.LNumber
	num := float64(L.ToNumber(1))
	prec := L.ToInt(2)

	mult := math.Pow10(prec)
	if num >= 0 {
		ret = math.Floor(num*mult+0.5) / mult
	} else {
		ret = math.Ceil(num*mult-0.5) / mult
	}
	lret = lua.LNumber(ret)
	L.Push(lret)
	return 1
}

func (lg *luagen) getLine(L *lua.LState) int {
	s := lg.currentItem.S

	lv := L.ToInt(1)
	if lv > len(s.Lines) {
		L.ArgError(1, "Index out of range")
	}
	ret := new(lua.LTable)
	for k, v := range s.Lines[lv] {
		ret.RawSetString(k, lua.LString(v))
	}
	L.Push(ret)
	return 1
}

func (lg *luagen) getLines(L *lua.LState) int {
	s := lg.currentItem.S
	ret := new(lua.LTable)
	for _, l := range s.Lines {
		ll := new(lua.LTable)
		for k, v := range l {
			ll.RawSetString(k, lua.LString(v))
		}
		ret.Append(ll)
	}
	L.Push(ret)
	return 1
}

func (lg *luagen) getChoice(L *lua.LState) int {
	s := lg.currentItem.S

	token := L.ToString(1)
	var found *config.Token
	for _, t := range s.Tokens {
		if t.Name == token {
			found = &t
			break
		}
	}
	if found == nil {
		L.ArgError(1, "Choice not found")
	}
	ret := new(lua.LTable)
	for _, l := range found.Choice {
		ret.Append(lua.LString(l))
	}
	L.Push(ret)
	return 1
}

func (lg *luagen) getFieldChoice(L *lua.LState) int {
	s := lg.currentItem.S

	token := L.ToString(1)
	field := L.ToString(2)

	var found *config.Token
	for _, t := range s.Tokens {
		if t.Name == token {
			found = &t
			break
		}
	}
	if found == nil {
		L.ArgError(1, "Field Choice not found")
	}
	ret := new(lua.LTable)
	for _, l := range found.FieldChoice {
		if fv, ok := l[field]; ok {
			ret.Append(lua.LString(fv))
		}
	}
	L.Push(ret)
	return 1
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
			// log.Debugf("Replaced %s token value with %s", tokenName, tokenValue)
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
		// log.Debugf("Added token")
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
		lg.states = make(map[string]*config.GeneratorState)
		lg.code = make(map[string]*lua.LFunction)
		lg.lstates = make(map[string]*sync.Pool)
		lg.initialized = true
	}
	s := item.S
	var gs *config.GeneratorState
	if s.CustomGenerator.SingleThreaded {
		s.LuaMutex.Lock()
		defer s.LuaMutex.Unlock()
		gs = s.GeneratorState
	} else {
		var ok bool
		if gs, ok = lg.states[s.Name]; !ok {
			lg.states[s.Name] = config.NewGeneratorState(s)
			gs = lg.states[s.Name]
		}
	}
	lg.currentItem = item

	// log.Debugf("Lua Gen called for sample '%s'", item.S.Name)
	if _, ok := lg.lstates[s.Name]; !ok {
		lg.lstates[s.Name] = &sync.Pool{
			New: func() interface{} {
				L := lua.NewState()
				// Register global variables
				L.SetGlobal("state", gs.LuaState)
				L.SetGlobal("options", luar.New(L, s.CustomGenerator.Options))
				L.SetGlobal("lines", gs.LuaLines)
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
				L.SetGlobal("round", L.NewFunction(lg.round))
				L.SetGlobal("getLine", L.NewFunction(lg.getLine))
				L.SetGlobal("getLines", L.NewFunction(lg.getLines))
				L.SetGlobal("getChoice", L.NewFunction(lg.getChoice))
				L.SetGlobal("getFieldChoice", L.NewFunction(lg.getFieldChoice))
				return L
			},
		}
	}
	L := lg.lstates[s.Name].Get().(*lua.LState)
	defer lg.lstates[s.Name].Put(L)
	L.SetGlobal("state", gs.LuaState)
	L.SetGlobal("options", luar.New(L, s.CustomGenerator.Options))
	L.SetGlobal("lines", gs.LuaLines)
	L.SetGlobal("count", luar.New(L, item.Count))
	L.SetGlobal("earliest", luar.New(L, item.Earliest))
	L.SetGlobal("latest", luar.New(L, item.Latest))
	L.SetGlobal("now", luar.New(L, item.Now))
	// L := lua.NewState()
	// defer L.Close()

	// log.Debugf("Calling DoString for %# v", s.CustomGenerator.Script)
	var f *lua.LFunction
	if _, ok := lg.code[s.Name]; !ok {
		var err error
		f, err = L.LoadString(s.CustomGenerator.Script)
		if err != nil {
			return fmt.Errorf("Error parsing script for generator '%s': %s", s.CustomGenerator.Name, err)
		}
		lg.code[s.Name] = f
	} else {
		f = lg.code[s.Name]
	}
	L.Push(f)
	err := L.PCall(0, lua.MultRet, nil)
	if err != nil {
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
