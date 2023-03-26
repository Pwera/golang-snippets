package twitter

import (
	"encoding/json"
	"github.com/pwera/app/db"
	"github.com/pwera/app/domain"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joeshaw/envdecode"
	"github.com/matryer/go-oauth/oauth"
)

type Twitter struct {
	authClient *oauth.Client
	creds      *oauth.Credentials
	Connector  *db.Connector
}
func (t *Twitter) setupTwitterAuth() {
	var ts struct {
		ConsumerKey    string `env:"SP_TWITTER_KEY,required"`
		ConsumerSecret string `env:"SP_TWITTER_SECRET,required"`
		AccessToken    string `env:"SP_TWITTER_ACCESSTOKEN,required"`
		AccessSecret   string `env:"SP_TWITTER_ACCESSSECRET,required"`
	}
	if err := envdecode.Decode(&ts); err != nil {
		log.Fatalln(err)
	}
	t.creds = &oauth.Credentials{
		Token:  ts.AccessToken,
		Secret: ts.AccessSecret,
	}
	t.authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  ts.ConsumerKey,
			Secret: ts.ConsumerSecret,
		},
	}
}

var (
	authSetupOnce sync.Once
	httpClient    *http.Client
)

func (t *Twitter) makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	authSetupOnce.Do(func() {
		t.setupTwitterAuth()
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: t.Connector.Dial,
			},
		}
	})
	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	req.Header.Set("Authorization", t.authClient.AuthorizationHeader(t.creds, "POST", req.URL, params))
	return httpClient.Do(req)
}

func (t *Twitter) readFromTwitter(votes chan<- string) {
	options, err := t.Connector.LoadOptions()
	if err != nil {
		log.Printf("Option loaded problem %v", err)
		return
	}
	u, err := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
	if err != nil {
		log.Printf("Creation filter problem %v\n", err)
		return
	}
	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	if err != nil {
		log.Printf("Filter request problem %v\n", err)
		return
	}
	resp, err := t.makeRequest(req, query)
	if err != nil {
		log.Printf("Request problem: %v\n", err)
		return
	}
	reader := resp.Body
	decoder := json.NewDecoder(reader)
	for {
		var t domain.Tweet
		if err := decoder.Decode(&t); err != nil {
			break
		}
		for _, option := range options {
			if strings.Contains(
				strings.ToLower(t.Text),
				strings.ToLower(option),
			) {
				log.Println("option:", option)
				votes <- option
			}
		}
	}
}

func (t *Twitter) StartTwitterStream(stopchan <-chan struct{}, votes chan<- string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()
		for {
			select {
			case <-stopchan:
				log.Println("Twitterem connection closing")
				return
			default:
				log.Println("Twittera searching")
				t.readFromTwitter(votes)
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return stoppedchan
}
