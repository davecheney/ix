package main

import (
	"encoding/xml"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/kr/fs"
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Feed struct {
	Entry []Entry `xml:"entry"`
}

type Entry struct {
	Id        string    `xml:"id"`
	Title     string    `xml:"title"`
	Published time.Time `xml:"published"`
	Content   string    `xml:"content"`
	Updates   []Update  `xml:"updates"`
	Author    struct {
		Name string `xml:"name"`
	} `xml:"author"`
	Owner      string   `xml:"owner>username"`
	Status     string   `xml:"status"`
	Label      []string `xml:"label"`
	MergedInto string   `xml:"mergedInto"`
	CC         []string `xml:"cc>username"`

	Dir      string
	Number   int
	Comments []Entry
}

type Update struct {
	Summary    string   `xml:"summary"`
	Owner      string   `xml:"ownerUpdate"`
	Label      string   `xml:"label"`
	Status     string   `xml:"status"`
	MergedInto string   `xml:"mergedInto"`
	CC         []string `xml:"cc>username"`
}

func ParseFile(path string) []Entry {
	f, err := os.Open(path)
	fatal(err)
	defer f.Close()
	var feed Feed
	fatal(xml.NewDecoder(f).Decode(&feed))
	return feed.Entry
}

type Issue struct {
	Id        int
	Title     string
	Published time.Time
	Content   string
	Status    string
	Label     []string
}

type Model struct {
	sync.Mutex
	issues []*Issue
}

func (m *Model) LoadIssues(dir string) {
	walker := fs.Walk(dir)
	for walker.Step() {
		st := walker.Stat()
		if st.IsDir() {
			continue
		}
		for _, e := range ParseFile(filepath.Join(dir, st.Name())) {
			id, err := strconv.Atoi(path.Base(e.Id))
			fatal(err)
			m.Lock()
			m.issues = append(m.issues,
				&Issue{
					Id:        id,
					Title:     e.Title,
					Published: e.Published,
					Content:   e.Content,
					Status:    e.Status,
					Label:     e.Label,
				})
			m.Unlock()
		}
	}
}

func (m *Model) FindIssueById(id int) (*Issue, bool) {
	m.Lock()
	defer m.Unlock()
	for _, i := range m.issues {
		if i.Id == id {
			return i, true
		}
	}
	return nil, false
}

func (m *Model) FindIssuesByTag(name string) []*Issue {
	m.Lock()
	defer m.Unlock()
	var issues []*Issue
	for _, i := range m.issues {
		for _, l := range i.Label {
			if l == name {
				issues = append(issues, i)
				break
			}
		}
	}
	return issues
}

func (m *Model) FindIssuesByTagAndStatus(name, status string) []*Issue {
	var issues []*Issue
	for _, i := range m.FindIssuesByTag(name) {
		if i.Status == status {
			issues = append(issues, i)
		}
	}
	return issues
}

func (m *Model) FindTags() []string {
	m.Lock()
	defer m.Unlock()
	var tags []string
	found := make(map[string]bool)
	for _, i := range m.issues {
		for _, tag := range i.Label {
			if !found[tag] {
				tags = append(tags, tag)
				found[tag] = true
			}
		}
	}
	return tags
}

type ById []*Issue

func (x ById) Len() int           { return len(x) }
func (x ById) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ById) Less(i, j int) bool { return x[i].Id < x[j].Id }
