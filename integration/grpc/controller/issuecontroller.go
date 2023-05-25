package controller

import (
	"github.com/pwera/ddd/domain"
	"net/http"
)

type IssueController struct {
	BaseController
	IssueService domain.IssueService
}

func (c IssueController) List(w http.ResponseWriter, r *http.Request) {
	issue, err := c.IssueService.Issues()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.MarshalAndWriteHeaders(issue, w)
}

func (c IssueController) Show(w http.ResponseWriter, r *http.Request)   {}
func (c IssueController) Create(w http.ResponseWriter, r *http.Request) {}
func (c IssueController) Delete(w http.ResponseWriter, r *http.Request) {}
