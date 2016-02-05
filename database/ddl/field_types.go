package internal

// metricTypePK maps the db table mtr.field.metricType to avoid lookups for things that change very slowly.
// additions here must be made to the table as well.
var metricTypePK = map[string]int32{
	"voltage":    1,
	"clock":      2,
	"satellites": 3,
}
