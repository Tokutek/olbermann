package olbermann

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

// DstatStyler is a Styler that produces output similar to dstat.
type DstatStyler struct {
	Period              time.Duration // How often to print
	LinesBetweenHeaders int           // After how many lines to print header info
	Logger              *log.Logger   // A logger to print to
}

// A Styler that produces good default output similar to dstat, to
// standard out, with timestamps, once per second, with headers every 24
// lines.
var BasicDstatStyler = DstatStyler{Period: time.Second, LinesBetweenHeaders: 24, Logger: log.New(os.Stdout, "", log.LstdFlags)}

func (s *DstatStyler) period() time.Duration {
	return s.Period
}

func (s *DstatStyler) linesBetweenHeaders() int {
	return s.LinesBetweenHeaders
}

func (s *DstatStyler) printHeader(mst *metricSetType) {
	var buf bytes.Buffer
	for i := range mst.metrics {
		mt := mst.metrics[i]
		if i > 0 {
			buf.WriteString("- -")
		}
		numSubHeaders := len(mt.reports)
		colWidth := 12*numSubHeaders + numSubHeaders - 1
		buf.WriteString(strings.Repeat("-", int(math.Floor(float64(colWidth-len(mt.name)-2)/2))))
		fmt.Fprintf(&buf, " %s ", strings.ToLower(mt.name))
		buf.WriteString(strings.Repeat("-", int(math.Ceil(float64(colWidth-len(mt.name)-2)/2))))
	}
	s.Logger.Print(buf.String())
	buf.Reset()
	for i := range mst.metrics {
		mt := mst.metrics[i]
		if i > 0 {
			buf.WriteString(" | ")
		}
		for j := range mt.reports {
			rt := mt.reports[j]
			if j > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(strings.Repeat(" ", 12-len(rt.name())))
			buf.WriteString(rt.name())
		}
	}
	s.Logger.Print(buf.String())
}

func (s *DstatStyler) printValues(mst *metricSetType, msv *metricSetValue) {
	var buf bytes.Buffer
	for i := range mst.metrics {
		mt := mst.metrics[i]
		mv := msv.metrics[i]
		if i > 0 {
			buf.WriteString(" | ")
		}
		for j := range mt.reports {
			rt := mt.reports[j]
			rv := mv.reports[j]
			if j > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(rt.string(rv.value))
		}
	}
	s.Logger.Print(buf.String())
}
