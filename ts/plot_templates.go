package ts

import (
	"bytes"
	"strings"
	"text/template"
	"time"
)

var funcMap = template.FuncMap{
	"date": func(t time.Time) string {
		return strings.Split(t.Format(time.RFC3339), "T")[0]
	},
}

type SVGPlot struct {
	template      *template.Template // the name for the template must be "plot"
	width, height int                // for the data on the plot, not the overall size.
}

func (s *SVGPlot) Draw(p Plot, b *bytes.Buffer) error {
	p.plt.width = s.width
	p.plt.height = s.height
	p.scaleData()
	p.setAxes()

	return s.template.ExecuteTemplate(b, "plot", p.plt)
}

var Scatter = SVGPlot{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(plotBaseTemplate + plotScatterTemplate)),
	width:    600,
	height:   170,
}

/*
templates are composed.  Any template using base must also define
'data' for plotting the template and 'keyMarker'.
*/
const plotBaseTemplate = `<?xml version="1.0"?>
<svg viewBox="0,0,800,270" class="svg" xmlns="http://www.w3.org/2000/svg" font-family="Arial, sans-serif" font-size="12px" fill="darkslategrey">
<g transform="translate(70,40)">
{{if .RangeAlert}}<rect x="0" y="0" width="600" height="170" fill="mistyrose"/>{{end}}

{{/* Grid, axes, title */}}
{{range .Axes.X}}
{{if .L}}
<polyline fill="none" stroke="paleturquoise" stroke-width="2" points="{{.X}},0 {{.X}},170"/>
<text x="{{.X}}" y="190" text-anchor="middle">{{.L}}</text>
{{else}}
<polyline fill="none" stroke="paleturquoise" stroke-width="2" points="{{.X}},0 {{.X}},170"/>
{{end}}
{{end}}

{{range .Axes.Y}}
{{if .L}}
<polyline fill="none" stroke="paleturquoise" stroke-width="1" points="0,{{.Y}} 600,{{.Y}}"/>
<polyline fill="none" stroke="darkslategrey" stroke-width="1" points="-4,{{.Y}} 4,{{.Y}}"/>
<text x="-7" y="{{.Y}}" text-anchor="end" dominant-baseline="middle">{{.L}}</text>
{{else}}
<polyline fill="none" stroke="darkslategrey" stroke-width="1" points="-2,{{.Y}} 2,{{.Y}}"/>
{{end}}
{{end}}

{{if .Axes.XAxisVis}}
<polyline fill="none" stroke="darkslategrey" stroke-width="1.0" points="-5, {{.Axes.XAxisY}}, 600, {{.Axes.XAxisY}}"/>
<g transform="translate(0,{{.Axes.XAxisY}})">
{{range .Axes.X}}
{{if .L}}
<polyline fill="none" stroke="darkslategrey" stroke-width="1.0" points="{{.X}}, -4, {{.X}}, 4"/>
{{else}}
<polyline fill="none" stroke="darkslategrey" stroke-width="1.0" points="{{.X}}, -2, {{.X}}, 2"/>
{{end}}
{{end}}
</g>

<polyline fill="none" stroke="darkslategrey" stroke-width="1.0" points="0,0 0,174"/>

{{end}}

<text x="320" y="-15" text-anchor="middle"  font-size="16px"  fill="black">{{.Axes.Title}}</text>
<text x="0" y="85" transform="rotate(90) translate(85,-25)" text-anchor="middle"  fill="black">{{.Axes.Ylabel}}</text>
<text x="320" y="208" text-anchor="middle"  font-size="14px" fill="black">Date</text>
{{/* end grid, axes, title */}}
{{if .Threshold.Show}}
<rect x="0" y="{{.Threshold.Y}}" width="600" height="{{.Threshold.H}}" fill="lime" opacity="0.2"/>
{{end}}
{{template "data" .}}
<circle cx="{{.LastPt.X}}" cy="{{.LastPt.Y}}" r="4" stroke="blue" fill="none" />
</g>
{{if not .Last.DateTime.IsZero}}
<text x="670" y="268" text-anchor="end" font-style="italic">
latest: <tspan fill="blue">{{ printf "%.1f" .Last.Value}} {{.Unit}}</tspan> ({{date .Last.DateTime}}) 
</text>
{{end}}
</svg>
`

const plotScatterTemplate = `
{{define "data"}}
{{range .Data}}
{{range .Pts}}<circle cx="{{.X}}" cy="{{.Y}}" r="2" fill="none" stroke="{{.Colour}}"/>{{end}}{{end}}
{{end}}

{{define "keyMarker"}}
<circle cx="{{.X}}" cy="{{.Y}}" r="2" fill="none" stroke="{{.L}}"/> 
{{end}}
`
