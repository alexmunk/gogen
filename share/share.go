package share

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/coccyx/gogen/generator"
	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
)

// Push pushes the running config to the Gogen API and creates a GitHub gist.  Returns the owner and ID of the Gist.
func Push(name string) (string, string) {
	var sample *config.Sample
	c := config.NewConfig()
	gh := NewGitHub()
	gu, _, err := gh.client.Users.Get("")

	gogen := *gu.Login + "/" + name

	if err != nil {
		c.Log.Fatalf("Error getting user in push: %s", err)
	}

	source := rand.NewSource(time.Now().UnixNano())
	randgen := rand.New(source)
	// Generate one event for our named sample
	for _, s := range c.Samples {
		if s.Name == name {
			sample = s

			if s.Description == "" {
				c.Log.Fatalf("Description not set for sample '%s'", s.Name)
			}

			c.Log.Debugf("Generating for Push() sample '%s'", s.Name)
			s.Output.Outputter = "buf"
			s.Output.OutputTemplate = "json"
			gq := make(chan *config.GenQueueItem)
			gqs := make(chan int)
			oq := make(chan *config.OutQueueItem)
			oqs := make(chan int)

			go generator.Start(gq, gqs)
			go outputter.Start(oq, oqs, 1)

			gqi := &config.GenQueueItem{Count: 1, Earliest: time.Now(), Latest: time.Now(), S: s, OQ: oq, Rand: randgen}
			gq <- gqi

			time.Sleep(time.Second)

			close(gq)
			close(oq)

			c.Log.Debugf("Buffer: %s", c.Buf.String())
			break
		}
	}
	if sample == nil {
		fmt.Printf("Sample '%s' not found\n", name)
		os.Exit(1)
	}

	gi := gh.Push(name)

	g := GogenInfo{Gogen: gogen, Name: name, Description: sample.Description, Notes: sample.Notes, Owner: *gu.Login, SampleEvent: c.Buf.String(), GistID: *gi.ID}
	Upsert(g)

	return *gu.Login, *gi.ID
}
