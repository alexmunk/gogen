package main

import (
	// "encoding/json"
	"github.com/coccyx/gogen/config"
	"github.com/kr/pretty"
)

func main() {

	// filename := os.Args[1]
	c := config.NewConfig()
	c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	// j, _ := json.MarshalIndent(c, "", "  ")
	// c.Log.Debugf("JSON Config: %s\n", j)
}
