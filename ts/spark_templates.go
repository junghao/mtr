package ts

import (
	"bytes"
	"text/template"
)

// Use human time to lable latest values

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
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(sparkLatestBaseTemplate + sparkScatterTemplate)),
	width:    100,
	height:   20,
}

const sparkLatestBaseTemplate = `<?xml version="1.0"?>
<svg viewBox="0,0,800,28" class="svg" xmlns="http://www.w3.org/2000/svg" font-family="Arial, sans-serif" font-size="14px" fill="darkslategrey">
<g transform="translate(3,4)"> 
{{if .RangeAlert}}<rect x="0" y="0" width="100" height="20" fill="mistyrose"/>{{end}}
{{if .Threshold.Show}}
<rect x="0" y="{{.Threshold.Y}}" width="100" height="{{.Threshold.H}}" fill="lightgrey" fill-opacity="0.3"/>
{{end}}
{{template "data" .}}
<circle cx="{{.LatestPt.X}}" cy="{{.LatestPt.Y}}" r="3" stroke="deepskyblue" fill="deepskyblue" />
</g>
<text font-style="italic" x="110" y="19" text-anchor="start">{{ printf "%.1f" .Latest.Value}} {{.Unit}} ({{date .Latest.DateTime}})</text>
</svg>	
`
const sparkScatterTemplate = `
{{define "data"}}
{{range .Data}}
{{range .Pts}}<circle cx="{{.X}}" cy="{{.Y}}" r="1" fill="none" stroke="deepskyblue"/>{{end}}
{{end}}
<circle cx="{{.LatestPt.X}}" cy="{{.LatestPt.Y}}" r="3" stroke="deepskyblue" fill="deepskyblue" />
{{end}}`
