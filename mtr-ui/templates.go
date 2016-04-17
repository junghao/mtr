package main

import (
	"html/template"
	"log"
)

var (
	borderTemplate       *template.Template
	tagsTemplate         *template.Template
	metricDetailTemplate *template.Template
)

func init() {
	loadTemplates()
}

func loadTemplates() {
	log.Println("Loading templates.")
	borderTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/demo.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	tagsTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/tags.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	metricDetailTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/metric_detail.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	log.Println("Done loading templates.")
}
