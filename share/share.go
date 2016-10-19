package share

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/coccyx/gogen/generator"
	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/outputter"
	"github.com/google/go-github/github"
)

// Push pushes the running config to the Gogen API and creates a GitHub gist.  Returns the owner and ID of the Gist.
func Push(name string) (string, string) {
	var sample *config.Sample
	c := config.NewConfig()
	gh := NewGitHub(true)
	gu, _, err := gh.client.Users.Get("")

	gogen := *gu.Login + "/" + name

	if err != nil {
		log.Fatalf("Error getting user in push: %s", err)
	}

	source := rand.NewSource(time.Now().UnixNano())
	randgen := rand.New(source)
	// Generate one event for our named sample
	for _, s := range c.Samples {
		if s.Name == name {
			sample = s

			if s.Description == "" {
				log.Fatalf("Description not set for sample '%s'", s.Name)
			}

			log.Debugf("Generating for Push() sample '%s'", s.Name)
			origOutputter := s.Output.Outputter
			origOutputTemplate := s.Output.OutputTemplate
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

			s.Output.Outputter = origOutputter
			s.Output.OutputTemplate = origOutputTemplate

			log.Debugf("Buffer: %s", c.Buf.String())
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

// Pull grabs a config from the Gogen API + GitHub gist and creates it on the filesystem for editing
func Pull(gogen string, dir string, deconstruct bool) {
	gogentokens := strings.Split(gogen, "/")
	var name string
	if len(gogentokens) > 1 {
		name = gogentokens[1]
	} else {
		name = gogen
	}
	gist := pull(gogen)
	for _, file := range gist.Files {
		filename := filepath.Join(dir, *file.Filename)
		client := &http.Client{}
		resp, err := client.Get(*file.RawURL)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		defer f.Close()
		if err != nil {
			log.Fatalf("Couldn't open file %s: %s", filename, err)
		}
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			log.Fatalf("Error writing to file %s: %s", filename, err)
		}
		if deconstruct {
			samplesDir := filepath.Join(dir, "samples")
			templatesDir := filepath.Join(dir, "templates")
			err := os.Mkdir(samplesDir, 0755)
			err = os.Mkdir(templatesDir, 0755)
			if err != nil && !os.IsExist(err) {
				log.Fatalf("Error creating directories %s or %s", samplesDir, templatesDir)
			}

			config.ResetConfig()
			os.Setenv("GOGEN_FULLCONFIG", filename)
			c := config.NewConfig()
			for x := 0; x < len(c.Samples); x++ {
				s := c.Samples[x]
				for y := 0; y < len(s.Tokens); y++ {
					t := c.Samples[x].Tokens[y]
					if t.SampleString != "" {
						fname := t.SampleString
						if fname[len(fname)-6:] == "sample" {
							f, err := os.OpenFile(filepath.Join(samplesDir, fname), os.O_WRONLY|os.O_CREATE, 0644)
							if err != nil {
								log.Fatalf("Unable to open file %s: %s", filepath.Join(samplesDir, fname), err)
							}
							defer f.Close()
							for _, v := range t.Choice {
								_, err := fmt.Fprintf(f, "%s\n", v)
								if err != nil {
									log.Fatalf("Error writing to file %s: %s", filepath.Join(samplesDir, fname), err)
								}
							}
							c.Samples[x].Tokens[y].Choice = []string{}
						} else if fname[len(fname)-3:] == "csv" {
							if len(s.Lines) > 0 {
								f, err := os.OpenFile(filepath.Join(samplesDir, fname), os.O_WRONLY|os.O_CREATE, 0644)
								if err != nil {
									log.Fatalf("Unable to open file %s: %s", filepath.Join(samplesDir, fname), err)
								}
								defer f.Close()
								w := csv.NewWriter(f)

								keys := make([]string, len(t.FieldChoice[0]))
								i := 0
								for k := range t.FieldChoice[0] {
									keys[i] = k
									i++
								}
								sort.Strings(keys)
								w.Write(keys)

								for _, l := range t.FieldChoice {
									values := make([]string, len(keys))
									for j, k := range keys {
										values[j] = l[k]
									}
									w.Write(values)
								}

								w.Flush()
								c.Samples[x].Tokens[y].FieldChoice = []map[string]string{}
							}
						}

						var outb []byte
						var err error
						if outb, err = json.MarshalIndent(s, "", "  "); err != nil {
							log.Fatalf("Cannot Marshal sample '%s', err: %s", s.Name, err)
						}
						err = ioutil.WriteFile(filepath.Join(samplesDir, name+".json"), outb, 0644)
						if err != nil {
							log.Fatalf("Cannot write file %s: %s", filepath.Join(samplesDir, name+".json"), err)
						}
					}
				}
			}

			for _, t := range c.Templates {
				var outb []byte
				var err error
				if outb, err = json.MarshalIndent(t, "", "  "); err != nil {
					log.Fatalf("Cannot Marshal template '%s', err: %s", t.Name, err)
				}
				err = ioutil.WriteFile(filepath.Join(templatesDir, t.Name+".json"), outb, 0644)
				if err != nil {
					log.Fatalf("Error writing file %s", filepath.Join(templatesDir, t.Name+".json"))
				}
			}

			err = os.Remove(filename)
			if err != nil {
				log.Debugf("Error removing original config file during deconstruction: %s", filename)
			}
		}
		break
	}
}

// PullFile pulls a config from the Gogen API + GitHub gist and writes it to a single file
func PullFile(gogen string, filename string) {
	gist := pull(gogen)
	for _, file := range gist.Files {
		log.Debugf("Writing config at file '%s' for gogen '%s'", filename, gogen)
		client := &http.Client{}
		resp, err := client.Get(*file.RawURL)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		defer f.Close()
		if err != nil {
			log.Fatalf("Couldn't open file %s: %s", filename, err)
		}
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			log.Fatalf("Error writing to file %s: %s", filename, err)
		}
		break
	}
}

func pull(gogen string) (gist *github.Gist) {
	g := Get(gogen)
	gh := NewGitHub(false)
	gist, _, err := gh.client.Gists.Get(g.GistID)
	if err != nil {
		log.Fatalf("Couldn't get gist: %s", err)
	}
	return gist
}
