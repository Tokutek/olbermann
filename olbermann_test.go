package olbermann

import (
	"bufio"
	"log"
	"os"
	"time"
)

type exampleValueSet struct {
	A int `type:"counter" report:"iter,total"`
	B int `type:"counter" report:"ewma1,cum,total"`
	//Tps int `type:"counter" report:"iter,cum" name:"tps"`
}

func gen(c chan<- interface{}) {
	for i := 0; i < 10; i++ {
		//c <- &exampleValueSet{A: 1, B: 1, Tps: 198273}
		c <- &exampleValueSet{A: 1, B: 1}
		time.Sleep(400 * time.Millisecond)
	}
}

// This is not a perfect example, because we can't rely on clocks to
// generate the right output.  So we use a log.Logger that doesn't print
// times, and we don't use the iter or cum report types on large values in
// this example.  Also, EWMA values take 10 samples to start reporting.
func Example() {
	c := make(chan interface{}, 10)
	r := &Reporter{C: c}
	go r.Feed()
	killer, err := r.Start(exampleValueSet{}, &DstatStyler{Period: time.Second, LinesBetweenHeaders: 0, Logger: log.New(os.Stdout, "example: ", 0)})
	if err != nil {
		return
	}
	gen(c)
	close(c)
	killer <- true
	// Output:
	// example: ----------- a ------------ ------------------ b ------------------
	// example:         iter        total |        ewma1          cum        total
	// example:         2.00            2 |         0.00         2.00            2
	// example:         1.00            4 |         0.00         2.00            4
	// example:         1.00            7 |         0.00         2.33            7
	// example:         0.50            9 |         0.00         2.25            9
}

// This output is too high precision to be an accurate test, but this is about what it would produce:
// Output:
// time,"A iter","A total","B ewma1","B cum","B total"
// "2014-04-12 00:41:06.921153316 -0400 EDT",1.999723,2.000000,1.999723,1.999723,2.000000
// "2014-04-12 00:41:07.921158351 -0400 EDT",0.999928,4.000000,1.935220,1.999856,4.000000
// "2014-04-12 00:41:08.921147586 -0400 EDT",0.999956,7.000000,1.874880,2.333230,7.000000
// "2014-04-12 00:41:09.921124252 -0400 EDT",0.499986,9.000000,1.786177,2.249938,9.000000
func ExampleCsv() {
	c := make(chan interface{}, 10)
	r := &Reporter{C: c}
	go r.Feed()
	csvKiller, err := r.Start(exampleValueSet{}, &CsvStyler{Period: time.Duration(10)*time.Millisecond, Writer: bufio.NewWriter(os.Stdout)})
	if err != nil {
		return
	}
	gen(c)
	close(c)
	csvKiller <- true
}
