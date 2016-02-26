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

func (s *SVGPlot) DrawBars(p Plot, b *bytes.Buffer) error {
	p.plt.width = s.width
	p.plt.height = s.height
	p.scaleData()
	p.setAxes()

	if err := p.setBars(); err != nil {
		return err
	}

	return s.template.ExecuteTemplate(b, "plot", p.plt)
}

func (p *Plot) setBars() error {
	if len(p.plt.Data) != 2 {
		return fmt.Errorf("drawing bars requires 2 data series.")
	}

	for i, _ := range p.plt.Data[0].Pts {
		// If the min max line is zero length it won't draw so add a little length.
		if p.plt.Data[0].Pts[i].Y == p.plt.Data[1].Pts[i].Y {
			p.plt.Lines = append(p.plt.Lines, line{
				X:      p.plt.Data[0].Pts[i].X,
				Y:      p.plt.Data[0].Pts[i].Y - 1,
				XX:     p.plt.Data[1].Pts[i].X,
				YY:     p.plt.Data[1].Pts[i].Y + 1,
				Colour: p.plt.Data[0].Pts[i].Colour,
			})
		} else {
			p.plt.Lines = append(p.plt.Lines, line{
				X:      p.plt.Data[0].Pts[i].X,
				Y:      p.plt.Data[0].Pts[i].Y,
				XX:     p.plt.Data[1].Pts[i].X,
				YY:     p.plt.Data[1].Pts[i].Y,
				Colour: p.plt.Data[0].Pts[i].Colour,
			})
		}
	}
	return nil
}

var Scatter = SVGPlot{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(plotBaseTemplate + plotScatterTemplate)),
	width:    600,
	height:   170,
}

var Bars = SVGPlot{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(plotBaseTemplate + plotBarsTemplate)),
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
{{if .Threshold.Show}}
<rect x="0" y="{{.Threshold.Y}}" width="600" height="{{.Threshold.H}}" fill="chartreuse" fill-opacity="0.2"/>
{{end}}
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

{{template "data" .}}
<circle cx="{{.LatestPt.X}}" cy="{{.LatestPt.Y}}" r="3" stroke="{{.LatestPt.Colour}}" fill="{{.LatestPt.Colour}}" />
</g>
<circle cx="500" cy="264" r="3" stroke="{{.LatestPt.Colour}}" fill="{{.LatestPt.Colour}}" />
<text x="510" y="268" text-anchor="start" font-style="italic">
latest: {{ printf "%.1f" .Latest.Value}} {{.Unit}} ({{date .Latest.DateTime}}) 
</text>
</svg>
`

const plotScatterTemplate = `
{{define "data"}}
{{range .Data}}
{{range .Pts}}<circle cx="{{.X}}" cy="{{.Y}}" r="2" fill="{{.Colour}}" stroke="{{.Colour}}"/>{{end}}{{end}}
{{end}}

{{define "keyMarker"}}
<circle cx="{{.X}}" cy="{{.Y}}" r="2" fill="none" stroke="{{.L}}"/> 
{{end}}
`
const plotBarsTemplate = `
{{define "data"}}
{{range .Lines}}<polyline fill="{{.Colour}}" stroke="{{.Colour}}" stroke-width="2" points="{{.X}},{{.Y}} {{.XX}},{{.YY}}"/>{{end}}{{end}}

{{define "keyMarker"}}
<circle cx="{{.X}}" cy="{{.Y}}" r="2" fill="none" stroke="{{.L}}"/> 
{{end}}
`
