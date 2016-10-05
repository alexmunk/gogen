package github

// Mostly from https://jacobmartins.com/2016/02/29/getting-started-with-oauth2-in-go/

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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
	}
	// Some random string, random for each request
	oauthStateString = uuid.NewV4().String()
)

type GitHub struct {
	done   chan int
	token  string
	client *github.Client
}

// NewGitHub returns a GitHub object, with a set auth token
func NewGitHub() *GitHub {
	gh := new(GitHub)
	gh.done = make(chan int)

	c := config.NewConfig()
	tokenFile := os.ExpandEnv("$GOGEN_HOME/.githubtoken")
	_, err := os.Stat(tokenFile)
	if err == nil {
		buf, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			c.Log.Fatalf("Error reading from file %s: %s", tokenFile, err)
		}
		gh.token = string(buf)
	} else {
		if !os.IsNotExist(err) {
			c.Log.Fatalf("Unexpected error accessing %s: %s", tokenFile, err)
		}
		http.HandleFunc("/GitHubLogin", gh.handleGitHubLogin)
		open.Run("http://localhost:46436/GitHubLogin")
		http.HandleFunc("/GitHubCallback", gh.handleGitHubCallback)
		go http.ListenAndServe(":46436", nil)
		<-gh.done

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
