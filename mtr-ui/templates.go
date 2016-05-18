package main

import (
	"html/template"
	"log"
)

var (
	homepageTemplate     *template.Template
	fieldTemplate        *template.Template
	dataTemplate         *template.Template
	tagsTemplate         *template.Template
	metricDetailTemplate *template.Template
	mapTemplate          *template.Template
)

func init() {
	loadTemplates()
}

func loadTemplates() {
	log.Println("Loading templates.")
	homepageTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/home.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	fieldTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/field.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	dataTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/data.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	tagsTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/tags.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	metricDetailTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/metric_detail.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	mapTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/map.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	log.Println("Done loading templates.")
}
