package olbermann

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

// user-defined
type ValueSet interface{}

// semi-deep copy (just for structs)
func clone(old ValueSet) ValueSet {
	val := reflect.Indirect(reflect.ValueOf(old))
	newVal := reflect.New(val.Type())
	for i := 0; i < val.NumField(); i++ {
		reflect.Indirect(newVal).Field(i).Set(val.Field(i))
	}
	return newVal.Interface().(ValueSet)
}

type Printer interface {
	Header(acc ValueSet) Stringser
	PrintHeader(s Stringser)
	Print(tDelta time.Duration, tTotal time.Duration, last ValueSet, cur ValueSet)
}

type LogPrinter struct {
	logger *log.Logger
}

const (
	cFlagIter = 1 << iota
	cFlagCum
	cFlagTotal
)

type cFlags int

type counter struct {
	name  string
	flags cFlags
}

type header []interface{}

type Stringser interface {
	Strings() (string, string)
}

func (h *header) Strings() (fst string, snd string) {
	colSizes := make([]int, len(*h))
	for i := range *h {
		switch elt := (*h)[i].(type) {
		case *counter:
			colSizes[i] = len(elt.name)
			numbers := 0
			if elt.flags&cFlagIter != 0 {
				numbers += 12
			}
			if elt.flags&cFlagCum != 0 {
				numbers += 12
			}
			if elt.flags&cFlagTotal != 0 {
				numbers += 12
			}
			if numbers > 12 {
				numbers += 1
			}
			if numbers > 25 {
				numbers += 1
			}
			if numbers > colSizes[i] {
				colSizes[i] = numbers
			}
		}
	}

	bufs := make([]bytes.Buffer, 2)
	for i := range *h {
		switch elt := (*h)[i].(type) {
		case *counter:
			if i > 0 {
				bufs[0].WriteString("- -")
			}
			j := 0
			for ; j < (colSizes[i]-len(elt.name))/2; j++ {
				bufs[0].WriteString("-")
			}
			if i == 0 {
				bufs[0].WriteString("-")
				j++
			}
			fmt.Fprintf(&bufs[0], " %s ", elt.name)
			j += len(elt.name) + 2
			for ; j < colSizes[i]; j++ {
				bufs[0].WriteString("-")
			}
		}
	}
	for i := range *h {
		switch elt := (*h)[i].(type) {
		case *counter:
			if i > 0 {
				bufs[1].WriteString(" | ")
			}
			if elt.flags&cFlagIter != 0 {
				bufs[1].WriteString("        iter")
			}
			if elt.flags&cFlagIter != 0 && (elt.flags&cFlagCum != 0 || elt.flags&cFlagTotal != 0) {
				bufs[1].WriteString(" ")
			}
			if elt.flags&cFlagCum != 0 {
				bufs[1].WriteString("         cum")
			}
			if elt.flags&cFlagCum != 0 && elt.flags&cFlagTotal != 0 {
				bufs[1].WriteString(" ")
			}
			if elt.flags&cFlagTotal != 0 {
				bufs[1].WriteString("       total")
			}
		}
	}
	fst = bufs[0].String()
	snd = bufs[1].String()
	return
}

func (lp *LogPrinter) Header(acc ValueSet) Stringser {
	st := reflect.TypeOf(acc)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}
	hdr := header(make([]interface{}, st.NumField(), st.NumField()))
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag
		n := tag.Get("name")
		switch tag.Get("type") {
		case "counter":
			ctr := counter{name: n}
			parts := strings.Split(tag.Get("report"), ",")
			for j := range parts {
				part := parts[j]
				switch part {
				case "iter":
					ctr.flags = ctr.flags | cFlagIter
				case "cum":
					ctr.flags = ctr.flags | cFlagCum
				case "total":
					ctr.flags = ctr.flags | cFlagTotal
				default:
					panic("invalid report tag: " + part)
				}
			}
			hdr[i] = &ctr
		default:
			panic("invalid type tag: " + tag.Get("type"))
		}
	}
	return &hdr
}

func (lp *LogPrinter) PrintHeader(s Stringser) {
	fst, snd := s.Strings()
	lp.logger.Println(fst)
	lp.logger.Println(snd)
}

func (lp *LogPrinter) Print(tDelta time.Duration, tTotal time.Duration, last ValueSet, cur ValueSet) {
	st := reflect.TypeOf(cur)
	lastVal := reflect.ValueOf(last)
	curVal := reflect.ValueOf(cur)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		lastVal = lastVal.Elem()
		curVal = curVal.Elem()
	}
	var buf bytes.Buffer
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag
		if i > 0 {
			buf.WriteString(" | ")
		}
		switch tag.Get("type") {
		case "counter":
			parts := strings.Split(tag.Get("report"), ",")
			for j := range parts {
				if j > 0 {
					buf.WriteString(" ")
				}
				part := parts[j]
				switch part {
				case "iter":
					fmt.Fprintf(&buf, "%12.2f", float64(curVal.Field(i).Int()-lastVal.Field(i).Int())/tDelta.Seconds())
				case "cum":
					fmt.Fprintf(&buf, "%12.2f", float64(curVal.Field(i).Int())/tTotal.Seconds())
				case "total":
					fmt.Fprintf(&buf, "%12d", curVal.Field(i).Int())
				default:
					panic("invalid report tag: " + part)
				}
			}
		default:
			panic("invalid type tag: " + tag.Get("type"))
		}
	}
	lp.logger.Print(buf.String())
}

type Reporter struct {
	period              time.Duration
	linesBetweenHeaders int
}

var HumanReporter = Reporter{period: time.Second, linesBetweenHeaders: 24}

func NewSingleHeaderReporter(p time.Duration) *Reporter {
	return &Reporter{period: p, linesBetweenHeaders: 0}
}

func NewNoHeaderReporter(p time.Duration) *Reporter {
	return &Reporter{period: p, linesBetweenHeaders: -1}
}

// Usage: olbermann.Feed(acc, valueChan)
func Feed(acc ValueSet, c <-chan ValueSet) (err error) {
	st := reflect.TypeOf(acc)
	val := reflect.ValueOf(acc)
	for st.Kind() == reflect.Ptr {
		st, val = st.Elem(), val.Elem()
	}
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag
		switch tag.Get("type") {
		case "counter":
			if !val.Field(i).CanSet() {
				err = errors.New("can't set field: " + tag.Get("name") + ", maybe it needs to be exported (capitalized)?")
				return
			}
		default:
			err = errors.New("invalid type tag: " + tag.Get("type"))
			return
		}
	}
	go func() {
		for v := range c {
			inc := reflect.ValueOf(v)
			for inc.Kind() == reflect.Ptr {
				inc = inc.Elem()
			}
			for i := 0; i < st.NumField(); i++ {
				tag := st.Field(i).Tag
				switch tag.Get("type") {
				case "counter":
					vf := val.Field(i)
					vf.SetInt(vf.Int() + inc.Field(i).Int())
				default:
					panic("invalid type tag: " + tag.Get("type"))
				}
			}
		}
	}()
	return
}

// Usage: killer := olbermann.Start(p, r, acc)
//        ...
//        killer <- true
func Start(p Printer, r Reporter, acc ValueSet) (kill chan bool) {
	kill = make(chan bool)
	go func() {
		ticker := time.Tick(r.period)
		header := p.Header(acc)
		if r.linesBetweenHeaders == 0 {
			p.PrintHeader(header)
		}
		linesAfterHeader := r.linesBetweenHeaders
		last := clone(acc)
		startTime := time.Now()
		lastTime := startTime
		for {
			if r.linesBetweenHeaders > 0 && linesAfterHeader == r.linesBetweenHeaders {
				p.PrintHeader(header)
				linesAfterHeader = 0
			}
			select {
			case <-kill:
				return
			case curTime := <-ticker:
				cur := clone(acc)
				p.Print(curTime.Sub(lastTime), curTime.Sub(startTime), last, cur)
				last, lastTime = cur, curTime
				linesAfterHeader++
			}
		}
	}()
	return
}

func BasicStart(name string, acc ValueSet) chan bool {
	return Start(&LogPrinter{log.New(os.Stdout, name, log.LstdFlags)}, HumanReporter, acc)
}
