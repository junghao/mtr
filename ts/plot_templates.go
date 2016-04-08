package ts

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

var funcMap = template.FuncMap{
	"date": func(t time.Time) string {
		dur := time.Now().UTC().Sub(t)

		d := int(dur.Hours() / 24)

		switch {
		case d > 730:
			return "years ago"
		case d <= 730 && d > 365:
			return "a year ago"
		case d > 14:
			return "weeks ago"
		case d <= 14 && d > 7:
			return "a week ago"
		case d > 1:
			return fmt.Sprintf("%d days ago", d)
		case d == 1:
			return "1 day ago"
		}

		h := int(dur.Minutes() / 60)

		switch {
		case h > 1:
			return fmt.Sprintf("%d hours ago", h)
		case h == 1:
			return "1 hour ago"
		}

		m := int(dur.Minutes())

		switch {
		case m > 1:
			return fmt.Sprintf("%d mins ago", m)
		case m == 1:
			return "1 min ago"
		}

		return "just now"
	},
	"sparse": func(p pts) bool {
		if len(p) < 100 {
			return true
		} else {
			return false
		}
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
	width:    780,
	height:   210,
}

var Line = SVGPlot{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(plotBaseTemplate + plotLineTemplate)),
	width:    780,
	height:   210,
}

var MixedAppMetrics = SVGPlot{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(plotAppMetricsTemplate + plotAppMixedTemplate)),
	width:    640,
	height:   210,
}

var LineAppMetrics = SVGPlot{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(plotAppMetricsTemplate + plotLineTemplate)),
	width:    640,
	height:   210,
}

/*
templates are composed.  Any template using base must also define
'data' for plotting the template and 'keyMarker'.
*/
const plotBaseTemplate = `<?xml version="1.0"?>
<svg viewBox="0,0,800,270" class="svg" xmlns="http://www.w3.org/2000/svg" font-family="Arial, sans-serif" font-size="12px" fill="lightgray">
<g transform="translate(10,10)">
<text x="0" y="0" text-anchor="start" dominant-baseline="hanging" font-size="14px" fill="darkslategray">{{.Axes.Title}}</text>
<text x="0" y="18" text-anchor="start" dominant-baseline="hanging" font-size="12px" fill="darkslategray">{{.Axes.SubTitle}}</text>
{{if .ShowLatest}}
<text x="780" y="0" text-anchor="end" dominant-baseline="hanging" fill="darkslategray">
{{ printf "%.1f" .Latest.Value}} {{.Unit}} ({{date .Latest.DateTime}})
</text>
{{end}}
{{if .Lables}}
<text x="780" y="18" text-anchor="end" dominant-baseline="hanging" font-size="8px" fill="darkslategray">{{range .Lables}}<tspan fill="{{.Colour}}">{{.Lable}}</tspan> {{end}}</text>
{{end}}
</g>

<g transform="translate(10,60)">

{{if .Threshold.Show}}
<rect x="0" y="{{.Threshold.Y}}" width="780" height="{{.Threshold.H}}" fill="lightgrey" fill-opacity="0.3"/>
{{end}}

<text x="{{400}}" y="220" text-anchor="middle" dominant-baseline="hanging">{{.Axes.Xlabel}}</text>

{{range .Axes.Y}}
{{if .L}}
<polyline fill="none" stroke="lightgray" stroke-width="1" points="0,{{.Y}} 780,{{.Y}}"/>
<text x="0" y="{{.Y}}" text-anchor="start" font-size="10px" dominant-baseline="ideographic">{{.L}}</text>
{{end}}
{{end}}

{{template "data" .}}
{{if .ShowLatest}}
<g style="stroke: {{.LatestColour}}; fill: none">
<circle cx="{{.LatestPt.X}}" cy="{{.LatestPt.Y}}" r="3" />
</g>
{{end}}
</g>

</svg>
`

const plotAppMetricsTemplate = `<?xml version="1.0"?>
<svg viewBox="0,0,800,270" class="svg" xmlns="http://www.w3.org/2000/svg" font-family="Arial, sans-serif" font-size="12px" fill="lightgray">
<g transform="translate(10,10)">
<text x="0" y="0" text-anchor="start" dominant-baseline="hanging" font-size="14px" fill="darkslategray">{{.Axes.Title}}</text>
<text x="0" y="18" text-anchor="start" dominant-baseline="hanging" font-size="12px" fill="darkslategray">{{.Axes.SubTitle}}</text>
{{if .ShowLatest}}
<text x="780" y="0" text-anchor="end" dominant-baseline="hanging" fill="darkslategray">
{{ printf "%.1f" .Latest.Value}} {{.Unit}} ({{date .Latest.DateTime}})
</text>{{end}}
{{if .Lables}}
<text x="780" y="18" text-anchor="end" dominant-baseline="hanging" font-size="8px" fill="darkslategray">{{range .Lables}}<tspan fill="{{.Colour}}" dy="10px" x="780">{{.Lable}}</tspan>{{end}}</text>
{{end}}
</g>

<g transform="translate(10,60)">

{{if .Threshold.Show}}
<rect x="0" y="{{.Threshold.Y}}" width="780" height="{{.Threshold.H}}" fill="lightgrey" fill-opacity="0.3"/>
{{end}}

<text x="{{400}}" y="220" text-anchor="middle" dominant-baseline="hanging">{{.Axes.Xlabel}}</text>

{{range .Axes.Y}}
{{if .L}}
<polyline fill="none" stroke="lightgray" stroke-width="1" points="0,{{.Y}} 640,{{.Y}}"/>
<text x="0" y="{{.Y}}" text-anchor="start" font-size="10px" dominant-baseline="ideographic">{{.L}}</text>
{{end}}
{{end}}

{{template "data" .}}
</g>

</svg>
`

const plotScatterTemplate = `
{{define "data"}}
{{range .Data}}
<g style="stroke: {{.Series.Colour}}; fill: none">
{{range .Pts}}<circle cx="{{.X}}" cy="{{.Y}}" r="2" />{{end}}
</g>
{{end}}
{{end}}`

const plotLineTemplate = `
{{define "data"}}
{{range .Data}}
<polyline style="stroke: {{.Series.Colour}}; fill: none; stroke-width: 2px; stroke-linecap: round; stroke-linejoin: round" points="{{range .Pts}}{{.X}},{{.Y}} {{end}}" />
{{end}}
{{end}}`

const plotAppMixedTemplate = `
{{define "data"}}
{{range .Data}}
{{if sparse .Pts}}
<g style="stroke: {{.Series.Colour}}; fill: {{.Series.Colour}}">
{{range .Pts}}<circle cx="{{.X}}" cy="{{.Y}}" r="2" />{{end}}
</g>
{{else}}
<polyline style="stroke: {{.Series.Colour}}; fill: none; stroke-width: 2px; stroke-linecap: round; stroke-linejoin: round" points="{{range .Pts}}{{.X}},{{.Y}} {{end}}" />
{{end}}{{end}}{{end}}`
