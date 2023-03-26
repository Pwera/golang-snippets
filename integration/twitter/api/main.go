package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pwera/app/db"
	"github.com/pwera/app/domain"
	"github.com/pwera/app/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type contextKey struct {
	name string
}
type server struct {
	db        *mgo.Session
	connector *db.Connector
}

func (s server) handlePoolsGet(w http.ResponseWriter, r *http.Request) {
	session := s.connector.SessionCopy()
	defer session.Close()

	c := session.DB("ballots").C("polls")
	var q *mgo.Query
	p := utils.NewPath(r.URL.Path)
	if p.HadID() {
		q = c.FindId(bson.ObjectIdHex(p.ID))
	} else {
		q = c.FindId(nil)
	}
	var result []*domain.Poll
	if err := q.All(&result); err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, &result)

}
func (s server) handlePoolsPost(w http.ResponseWriter, r *http.Request) {
	respondErr(w, r, http.StatusOK)
}
func (s server) handlePoolsDelete(w http.ResponseWriter, r *http.Request) {
	respondErr(w, r, http.StatusOK)
}

func (s server) handlePools(w http.ResponseWriter, r *http.Request) {
	if ok := handlers[r.Method]; ok != nil {
		handlers[r.Method](w, r)
		return
	}
	respondHTTPErr(w, r, http.StatusNotFound)
}

var (
	handlers      = map[string]func(w http.ResponseWriter, r *http.Request){}
	contextAPIKey = &contextKey{"key"}
)

func init() {
	//handlers["GET"]= func(writer http.ResponseWriter, request *http.Request) {
	//
	//}
}
func APIKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(contextAPIKey).(string)
	return key, ok
}
func isValidAPIKey(_ string) bool { return true }

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !isValidAPIKey(key) {
			respondErr(w, r, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), contextAPIKey, key)
		fn(w, r.WithContext(ctx))
	}
}
func withCors(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(w, r)
	}
}

func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
func encodeBody(w http.ResponseWriter, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		encodeBody(w, data)
	}
}
func respondErr(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...),
		},
	})
}

func respondHTTPErr(w http.ResponseWriter, r *http.Request, status int) {
	respondErr(w, r, status, http.StatusText(status))
}

func main() {
	s := server{connector: db.NewConnector()}
	defer s.connector.CloseDb()

	handlers["GET"] = s.handlePoolsGet
	handlers["DELETE"] = s.handlePoolsDelete
	handlers["POST"] = s.handlePoolsPost

	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCors(withAPIKey(s.handlePools)))
	http.ListenAndServe(":8080", mux)
}
