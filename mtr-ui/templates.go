package main

import (
	"github.com/GeoNet/mtr/mtrpb"
	"html/template"
	"log"
	"time"
)

var (
	homepageTemplate     *template.Template
	fieldTemplate        *template.Template
	dataTemplate         *template.Template
	appTemplate          *template.Template
	tagSearchTemplate    *template.Template
	metricDetailTemplate *template.Template
	mapTemplate          *template.Template
	tagPageTemplate      *template.Template
	appPlotTemplate      *template.Template
)

var funcMap = template.FuncMap{
	"rfc3339str": func(sec int64) string {
		return time.Unix(sec, 0).Format(time.RFC3339)
	},
	"latencyColour": func(r *mtrpb.DataLatency, lower, upper int32) string {
		if upper == 0 && lower == 0 {
			return "red"
		}
		if r.Mean < float32(lower) || r.Mean > float32(upper) {
			return "red"
		}
		if r.Fifty != 0 && (r.Fifty < lower || r.Fifty > upper) {
			return "red"
		}
		if r.Ninety != 0 && (r.Ninety < lower || r.Ninety > upper) {
			return "red"
		}
		return "black"
	},
	"fieldColour": func(r *mtrpb.FieldMetric, lower, upper int32) string {
		if upper == 0 && lower == 0 {
			return "red"
		}
		if r.Value < float32(lower) || r.Value > float32(upper) {
			return "red"
		}
		return "black"
	},
}

func init() {
	loadTemplates()
}

func loadTemplates() {
	log.Println("Loading templates.")
	homepageTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/home.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	fieldTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/field.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	dataTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/data.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	appTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/app.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	appPlotTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/app_plot.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	tagSearchTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/tag_search.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	metricDetailTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/metric_detail.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	mapTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/map.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	tagPageTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/tmpl/tag_page.html", "assets/tmpl/components.html", "assets/tmpl/tag_list.html", "assets/tmpl/border.html"))
	log.Println("Done loading templates.")
}
