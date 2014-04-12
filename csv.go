package olbermann

import (
	"bufio"
	"fmt"
	"time"
)

// CsvStyler is a Styler that produces output in csv format.
type CsvStyler struct {
	Period              time.Duration // How often to print
	Writer              *bufio.Writer   // A writer to print to
}

func (s *CsvStyler) period() time.Duration {
	return s.Period
}

func (s *CsvStyler) linesBetweenHeaders() int {
	return 0
}

func (s *CsvStyler) printHeader(mst *metricSetType) {
	s.Writer.WriteString("time")
	for i := range mst.metrics {
		mt := mst.metrics[i]
		s.Writer.WriteString(",")
		for j := range mt.reports {
			rt := mt.reports[j]
			if j > 0 {
				s.Writer.WriteString(",")
			}
			s.Writer.WriteString("\"")
			s.Writer.WriteString(mt.name)
			s.Writer.WriteString(" ")
			s.Writer.WriteString(rt.name())
			s.Writer.WriteString("\"")
		}
	}
	s.Writer.WriteString("\n")
}

func (s *CsvStyler) printValues(curTime time.Time, mst *metricSetType, msv *metricSetValue) {
	s.Writer.WriteString("\"")
	s.Writer.WriteString(curTime.String())
	s.Writer.WriteString("\"")
	for i := range msv.metrics {
		mv := msv.metrics[i]
		s.Writer.WriteString(",")
		for j := range mv.reports {
			rv := mv.reports[j]
			if j > 0 {
				s.Writer.WriteString(",")
			}
			s.Writer.WriteString(fmt.Sprintf("%f", rv.value))
		}
	}
	s.Writer.WriteString("\n")
	s.Writer.Flush()
}
