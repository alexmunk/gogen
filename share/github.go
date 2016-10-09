package share

// Mostly from https://jacobmartins.com/2016/02/29/getting-started-with-oauth2-in-go/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/kr/pretty"
	uuid "github.com/satori/go.uuid"
	"github.com/skratchdot/open-golang/open"

	config "github.com/coccyx/gogen/internal"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

var (
	gitHubClientID     string // Passed in during the build process
	gitHubClientSecret string // Passed in during the build process
	oauthConf          = &oauth2.Config{
		RedirectURL:  "http://localhost:46436/GitHubCallback",
		ClientID:     gitHubClientID,
		ClientSecret: gitHubClientSecret,
		Endpoint:     githuboauth.Endpoint,
		Scopes:       []string{"gist"},
	}
	// Some random string, random for each request
	oauthStateString = uuid.NewV4().String()
)

// GitHub allows posting gists to a user's GitHub
type GitHub struct {
	done   chan int
	token  string
	client *github.Client
	c      *config.Config
}

// Push will create a public gist of "name.json" from our running config
func (gh *GitHub) Push(name string) {
	gist := new(github.Gist)
	files := make(map[github.GistFilename]github.GistFile)

	file := new(github.GistFile)
	fname := name + ".json"
	file.Filename = &fname
	var outb []byte
	var outbs *string
	var err error
	if outb, err = json.MarshalIndent(gh.c, "", "  "); err != nil {
		gh.c.Log.Fatalf("Cannot Marshal c.Global, err: %s", err)
	}
	outbs = new(string)
	*outbs = string(outb)
	file.Content = outbs
	files[github.GistFilename(name)] = *file

	gist.Files = files
	gist.Description = &name
	public := true
	gist.Public = &public

	_, _, err = gh.client.Gists.Create(gist)
	if err != nil {
		gh.c.Log.Fatalf("Error creating gist %# v: %s", pretty.Formatter(gist), err)
	}
}

// Here's some code that I'm saving that decomposed a sample into files, but I didn't like the way it worked for posting to GitHub
// So now, we post the same way the old config command worked, one big file, but I'm saving this code because it'll allow me to later
// deconstruct that, optionally, when pulling, into separate files for easier editing

// for _, s := range gh.c.Samples {
// 	if len(s.Path) > 0 {
// 		file := new(github.GistFile)
// 		fname := filepath.Base(s.Path)
// 		file.Filename = &fname
// 		if fname[len(fname)-6:] == "sample" {
// 			out := ""
// 			for _, v := range s.Lines {
// 				out += fmt.Sprintf("%s\n", v["_raw"])
// 			}
// 			file.Content = &out
// 		} else if fname[len(fname)-3:] == "csv" {
// 			if len(s.Lines) > 0 {
// 				buf := new(bytes.Buffer)
// 				w := csv.NewWriter(buf)

// 				keys := make([]string, len(s.Lines[0]))
// 				i := 0
// 				for k := range s.Lines[0] {
// 					keys[i] = k
// 					i++
// 				}
// 				sort.Strings(keys)
// 				w.Write(keys)

// 				for _, l := range s.Lines {
// 					values := make([]string, len(keys))
// 					for j, k := range keys {
// 						values[j] = l[k]
// 					}
// 					w.Write(values)
// 				}

// 				w.Flush()
// 				outbs = new(string)
// 				*outbs = buf.String()
// 				file.Content = outbs
// 			}
// 		} else {
// 			var outb []byte
// 			var outbs *string
// 			var err error
// 			if outb, err = yaml.Marshal(s); err != nil {
// 				gh.c.Log.Fatalf("Cannot Marshal sample '%s', err: %s", s.Name, err)
// 			}
// 			outbs = new(string)
// 			*outbs = string(outb)
// 			file.Content = outbs
// 		}
// 		files[github.GistFilename(filepath.Base(s.Path))] = *file
// 	}
// }

// for _, t := range gh.c.Templates {
// 	if len(t.Path) > 0 {
// 		file := new(github.GistFile)
// 		fname := filepath.Base(t.Path)
// 		file.Filename = &fname
// 		var outb []byte
// 		var outbs *string
// 		var err error
// 		if outb, err = yaml.Marshal(t); err != nil {
// 			gh.c.Log.Fatalf("Cannot Marshal template '%s', err: %s", t.Name, err)
// 		}
// 		outbs = new(string)
// 		*outbs = string(outb)
// 		file.Content = outbs
// 		files[github.GistFilename(filepath.Base(t.Path))] = *file
// 	}
// }

// NewGitHub returns a GitHub object, with a set auth token
func NewGitHub() *GitHub {
	gh := new(GitHub)
	gh.done = make(chan int)

	if oauthConf.ClientID == "" {
		oauthConf.ClientID = os.Getenv("GITHUB_OAUTH_CLIENT_ID")
	}
	if oauthConf.ClientSecret == "" {
		oauthConf.ClientSecret = os.Getenv("GITHUB_OAUTH_CLIENT_SECRET")
	}

	c := config.NewConfig()
	gh.c = c
	tokenFile := os.ExpandEnv("$GOGEN_HOME/.githubtoken")
	_, err := os.Stat(tokenFile)
	if err == nil {
		buf, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			c.Log.Fatalf("Error reading from file %s: %s", tokenFile, err)
		}
		gh.token = string(buf)
		c.Log.Debugf("Getting GitHub token '%s' from file", gh.token)
	} else {
		if !os.IsNotExist(err) {
			c.Log.Fatalf("Unexpected error accessing %s: %s", tokenFile, err)
		}
		http.HandleFunc("/GitHubLogin", gh.handleGitHubLogin)
		open.Run("http://localhost:46436/GitHubLogin")
		http.HandleFunc("/GitHubCallback", gh.handleGitHubCallback)
		go http.ListenAndServe(":46436", nil)
		<-gh.done
		c.Log.Debugf("Getting GitHub token '%s' from oauth", gh.token)

		err = ioutil.WriteFile(tokenFile, []byte(gh.token), 400)
		if err != nil {
			c.Log.Fatalf("Error writing token to file %s: %s", tokenFile, err)
		}
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh.token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	gh.client = github.NewClient(tc)
	return gh
}

func (gh *GitHub) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (gh *GitHub) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("Code exchange failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	gh.token = token.AccessToken
	gh.done <- 1
}
