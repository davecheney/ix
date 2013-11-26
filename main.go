package main

import (
	"log"
	"sort"
	"strconv"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
)

var model = &Model{issues: make(map[int]*Issue)}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	fatal(err)
	return i
}

func ShowIssue(r render.Render, params martini.Params) {
	id := atoi(params["id"])
	issue, ok := model.FindIssueById(id)
	if !ok {
		r.HTML(404, "error/404", nil)
		return
	}
	r.HTML(200, "showissue", issue)
}

func ShowTag(r render.Render, params martini.Params) {
	name := params["name"]
	issues := model.FindIssuesByTag(name)
	if len(issues) == 0 {
		r.HTML(404, "error/404", name)
		return
	}
	sort.Sort(ById(issues))
	r.HTML(200, "showtag", struct {
		Name   string
		Issues []*Issue
	}{name, issues})
}

func ShowStatus(r render.Render, params martini.Params) {
	status := params["status"]
	issues := model.FindIssuesByStatus(status)
	if len(issues) == 0 {
		r.HTML(404, "error/404", status)
		return
	}
	sort.Sort(ById(issues))
	r.HTML(200, "showstatus", struct {
		Status string
		Issues []*Issue
	}{status, issues})
}

func ShowTagAndStatus(r render.Render, params martini.Params) {
	name, status := params["name"], params["status"]
	issues := model.FindIssuesByTagAndStatus(name, status)
	if len(issues) == 0 {
		r.HTML(404, "error/404", name)
		return
	}
	sort.Sort(ById(issues))
	r.HTML(200, "showtag", struct {
		Name   string
		Issues []*Issue
	}{name, issues})
}

func ShowAllTags(r render.Render) {
	tags := model.FindTags()
	r.HTML(200, "showtags", struct{ Tags []string }{tags})
}

func ShowAllStatuses(r render.Render) {
	statuses := model.FindStatuses()
	r.HTML(200, "showstatuses", struct{ Statuses []string }{statuses})
}

func ShowComments(r render.Render, params martini.Params) {
	comments := model.FindComments(params["name"])
	r.HTML(200, "showcomments", struct{ Name string; Comments []*Entry}{Name: params["name"], Comments: comments})
}

func Overview(r render.Render) {
	tags := model.FindTags()
	r.HTML(200, "overview", struct{ Tags []string }{tags})
}

func main() {
	go func() {
		model.LoadIssues("issues")
		log.Println("issues loaded")
		model.LoadComments("comments")
		log.Println("comments loaded")
	}()

	m := martini.Classic()
	m.Use(render.Renderer("templates"))

	m.Get("/issue/:id", ShowIssue)
	m.Get("/tag/:name", ShowTag)
	m.Get("/tag", ShowAllTags)
	m.Get("/tag/:name/:status", ShowTagAndStatus)
	m.Get("/status/:status", ShowStatus)
	m.Get("/status", ShowAllStatuses)
	m.Get("/comments/:name", ShowComments)
	m.Get("/", Overview)
	m.Run()
}
