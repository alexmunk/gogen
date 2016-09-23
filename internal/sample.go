package config

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	strftime "github.com/hhkbp2/go-strftime"
	logging "github.com/op/go-logging"
	"github.com/satori/go.uuid"
)

const randStringLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const randHexLetters = "ABCDEF0123456789"

type Sample struct {
	Name            string              `json:"name"`
	Disabled        bool                `json:"disabled"`
	Generator       string              `json:"generator"`
	Outputter       string              `json:"outputter"`
	OutputTemplate  string              `json:"outputTemplate"`
	Rater           string              `json:"rater"`
	Interval        int                 `json:"interval"`
	Delay           int                 `json:"delay"`
	Count           int                 `json:"count"`
	Earliest        string              `json:"earliest"`
	Latest          string              `json:"latest"`
	Begin           string              `json:"begin"`
	End             string              `json:"end"`
	RandomizeCount  float32             `json:"randomizeCount"`
	RandomizeEvents bool                `json:"randomizeEvents"`
	Tokens          []Token             `json:"tokens"`
	Lines           []map[string]string `json:"lines"`
	Field           string              `json:"field"`

	// Internal use variables
	Log            *logging.Logger `json:"-"`
	Gen            Generator       `json:"-"`
	Out            Outputter       `json:"-"`
	EarliestParsed time.Duration   `json:"-"`
	LatestParsed   time.Duration   `json:"-"`
	BeginParsed    time.Time       `json:"-"`
	EndParsed      time.Time       `json:"-"`
	Current        time.Time       `json:"-"` // If we are backfilling or generating for a specified time window, what time is it?
	Realtime       bool            `json:"-"` // Are we done doing batch backfill or specified time window?
}

// Clock allows for implementers to keep track of their own view
// of current time.  In Gogen, this is used for being able to generate
// events between certain time windows, or backfill from a certain time
// while continuing to run in real time.
type Clock interface {
	Now() time.Time
}

// Now returns the current time for the sample, and handles
// whether we are currently generating a backfill or
// specified time window or whether we should be generating
// events in realtime
func (s *Sample) Now() time.Time {
	if !s.Realtime {
		return s.Current
	} else {
		return time.Now()
	}
}

// Token describes a replacement task to run against a sample
type Token struct {
	Name           string              `json:"name"`
	Format         string              `json:"format"`
	Token          string              `json:"token"`
	Type           string              `json:"type"`
	Replacement    string              `json:"replacement"`
	Group          int                 `json:"group"`
	Sample         *Sample             `json:"-"`
	SampleString   string              `json:"sample"`
	Field          string              `json:"field"`
	SrcField       string              `json:"srcField"`
	Precision      int                 `json:"precision"`
	Lower          int                 `json:"lower"`
	Upper          int                 `json:"upper"`
	Length         int                 `json:"length"`
	WeightedChoice []WeightedChoice    `json:"weightedChoice"`
	FieldChoice    []map[string]string `json:"fieldChoice"`
	Choice         []string            `json:"choice"`
}

type WeightedChoice struct {
	Weight int    `json:"weight"`
	Choice string `json:"choice"`
}

func (t Token) Replace(event *string, choice *int, et time.Time, lt time.Time) error {
	// s := t.Sample
	e := *event

	switch t.Format {
	// TODO Replacing template tokens one by one is inefficient, but test to see how inefficient before optimizing
	case "template":
		if pos := strings.Index(t.Token, e); pos >= 0 {
			replacement, err := t.GenReplacement(choice, et, lt)
			if err != nil {
				return err
			}
			part1 := e[:pos]
			part2 := e[pos+len(t.Token):]
			temps := part1 + replacement + part2
			event = &temps
		} else {
			return fmt.Errorf("Token '%s' not found in field '%s' of event '%s'", t.Token, t.Field, *event)
		}
	}
	return nil
}

// GenReplacement generates a replacement value for the token.  choice allows the user to specify
// a specific value to choose in the array.  This is useful for saving picks amongst tokens.
func (t Token) GenReplacement(choice *int, et time.Time, lt time.Time) (string, error) {
	c := *choice
	switch t.Type {
	case "timestamp":
		td := lt.Sub(et)

		var tdr int
		if int(td) > 0 {
			tdr = rand.Intn(int(td))
		}
		rd := time.Duration(tdr)
		replacementTime := lt.Add(rd * -1)
		return strftime.Format(t.Replacement, replacementTime), nil
	case "static":
		return t.Replacement, nil
	case "random":
		switch t.Replacement {
		case "int":
			offset := 0 - t.Lower
			return strconv.Itoa(rand.Intn(t.Upper-offset) + offset), nil
		case "float":
			lower := t.Lower * int(math.Pow10(t.Precision))
			offset := 0 - lower
			upper := t.Upper * int(math.Pow10(t.Precision))
			f := float64(rand.Intn(upper-offset)+offset) / math.Pow10(t.Precision)
			return strconv.FormatFloat(f, 'f', t.Precision, 64), nil
		case "string", "hex":
			b := make([]byte, t.Length)
			var l string
			if t.Replacement == "string" {
				l = randStringLetters
			} else {
				l = randHexLetters
			}
			for i := range b {
				b[i] = l[rand.Intn(len(l))]
			}
			return string(b), nil
		case "guid":
			u := uuid.NewV4()
			return u.String(), nil
		case "ipv4":
			var ret string
			for i := 0; i < 4; i++ {
				ret = ret + strconv.Itoa(rand.Intn(255)) + "."
			}
			ret = strings.TrimRight(ret, ".")
			return ret, nil
		case "ipv6":
			var ret string
			for i := 0; i < 8; i++ {
				ret = ret + fmt.Sprintf("%x", rand.Intn(65535)) + ":"
			}
			ret = strings.TrimRight(ret, ":")
			return ret, nil
		}
	case "choice":
		if c == -1 {
			c = rand.Intn(len(t.Choice))
			*choice = c
		}
		return t.Choice[c], nil
	case "weightedChoice":
		// From http://eli.thegreenplace.net/2010/01/22/weighted-random-generation-in-python/
		var totals []int
		runningTotal := 0

		for _, w := range t.WeightedChoice {
			runningTotal += w.Weight
			totals = append(totals, runningTotal)
		}

		r := rand.Float64() * float64(runningTotal)
		for j, total := range totals {
			if r < float64(total) {
				*choice = j
				break
			}
		}
		if *choice > 0 {
			return t.WeightedChoice[*choice].Choice, nil
		}
	case "fieldChoice":
		if c == -1 {
			c = rand.Intn(len(t.FieldChoice))
			*choice = c
		}
		return t.FieldChoice[c][t.SrcField], nil
	}
	return "", fmt.Errorf("getReplacement called with invalid type")
}
