package main

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"

	"github.com/codegangsta/martini"
	"github.com/kr/fs"
)

var model = new(Model)
var view *template.Template

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	fatal(err)
	return i
}

func LoadTemplates() {
	walker := fs.Walk("templates")
	var files []string
	for walker.Step() {
		fi := walker.Stat()
		if fi.IsDir() {
			continue
		}
		files = append(files, filepath.Join("templates", fi.Name()))
	}
	fatal(walker.Err())
	var err error
	view, err = template.ParseFiles(files...)
	fatal(err)
}

func View(name string, arg interface{}) (int, string) {
	var w bytes.Buffer
	if err := view.ExecuteTemplate(&w, name, arg); err != nil {
		return 501, err.Error()
	}
	return 200, w.String()
}

func ShowIssue(params martini.Params) (int, string) {
	id := atoi(params["id"])
	issue, ok := model.FindIssueById(id)
	if !ok {
		return 404, fmt.Sprintf("Issue %d not found", id)
	}
	return View("showissue", issue)
}

func ShowTag(params martini.Params) (int, string) {
	name := params["name"]
	issues := model.FindIssuesByTag(name)
	if len(issues) == 0 {
		return 404, fmt.Sprintf("Tag %q not found", name)
	}
	return View("showtag", struct {
		Name   string
		Issues []*Issue
	}{name, ById(issues)})
}

func ShowStatus(params martini.Params) (int, string) {
	status := params["status"]
	issues := model.FindIssuesByStatus(status)
	if len(issues) == 0 {
		return 404, fmt.Sprintf("Status %q not found", status)
	}
	return View("showstatus", struct {
		Status  string
		Issues []*Issue
	}{status, ById(issues)})
}

func ShowTagAndStatus(params martini.Params) (int, string) {
	name, status := params["name"], params["status"]
	issues := model.FindIssuesByTagAndStatus(name, status)
	if len(issues) == 0 {
		return 404, fmt.Sprintf("Tag/Status %q/%q not found", name)
	}
	return View("showtag", struct {
		Name   string
		Issues []*Issue
	}{name, ById(issues)})
}

func ShowAllTags() (int,string) {
	tags := model.FindTags()
	return View("showtags", tags)
}

func Overview() (int, string) {
	tags := model.FindTags()
	return View("overview", struct{ Tags []string }{tags})
}

func main() {
	LoadTemplates()
	go model.LoadIssues("issues")

	m := martini.Classic()
	m.Get("/issue/:id", ShowIssue)
	m.Get("/tag/:name", ShowTag)
	m.Get("/tag", ShowAllTags)
	m.Get("/tag/:name/:status", ShowTagAndStatus)
	m.Get("/status/:status", ShowStatus)
	m.Get("/", Overview)
	m.Run()
}
