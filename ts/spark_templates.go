package ts

import (
	"bytes"
	"text/template"
)

type SVGSpark struct {
	template      *template.Template // the name for the template must be "plot"
	width, height int                // for the data on the plot, not the overall size.
}

func (s *SVGSpark) Draw(p Plot, b *bytes.Buffer) error {
	p.plt.width = s.width
	p.plt.height = s.height

	p.scaleData()

	return s.template.ExecuteTemplate(b, "plot", p.plt)
}

var SparkScatterLatest = SVGSpark{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(sparkLatestBaseTemplate + sparkThresholdTemplate + sparkScatterTemplate)),
	width:    100,
	height:   20,
}

const sparkLatestBaseTemplate = `<?xml version="1.0"?>
<svg viewBox="0,0,800,28" class="svg" xmlns="http://www.w3.org/2000/svg" font-family="Arial, sans-serif" font-size="14px" fill="darkslategrey">
<g transform="translate(3,4)"> 
{{if .RangeAlert}}<rect x="0" y="0" width="100" height="20" fill="mistyrose"/>{{end}}
{{template "threshold" .Threshold}}
{{template "data" .Data}}<circle cx="{{.LastPt.X}}" cy="{{.LastPt.Y}}" r="3" stroke="blue" fill="none" />
</g>
<text font-style="italic" fill="black" x="110" y="19" text-anchor="start"><tspan fill="blue">{{ printf "%.2f" .Last.Value}} {{.Unit}}</tspan> ({{date .Last.DateTime}})</text>
</svg>	
`

const sparkThresholdTemplate = `{{define "threshold"}}{{if .Show}}
<rect x="0" y="{{.Y}}" width="100" height="{{.H}}" fill="lime" opacity="0.2"/>
{{end}}{{end}}`

const sparkScatterTemplate = `{{define "data"}}{{range .}}
{{range .Pts}}<circle cx="{{.X}}" cy="{{.Y}}" r=".5" fill="none" stroke="{{.Colour}}"/>{{end}}{{end}}{{end}}
`
