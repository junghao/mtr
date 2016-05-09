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
)

func init() {
	loadTemplates()
}

func loadTemplates() {
	log.Println("Loading templates.")
	homepageTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/home.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html", "assets/tmpl/field_components.html", "assets/tmpl/data_components.html"))
	fieldTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/field.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html", "assets/tmpl/field_components.html"))
	dataTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/data.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html", "assets/tmpl/data_components.html"))
	tagsTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/tags.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	metricDetailTemplate = template.Must(template.New("t").ParseFiles("assets/tmpl/metric_detail.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	log.Println("Done loading templates.")
}
