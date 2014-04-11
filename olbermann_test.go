package olbermann

import (
	"log"
	"os"
	"time"
)

type exampleValueSet struct {
	A int `type:"counter" report:"iter,total"`
	B int `type:"counter" report:"total"`
	//Tps int `type:"counter" report:"iter,cum" name:"tps"`
}

// This is not a perfect example, because we can't rely on clocks to
// generate the right output.  So we use a log.Logger that doesn't print
// times, and we don't use the iter or cum report types on large values in
// this example.
func Example() {
	c := make(chan interface{}, 10)
	r := &Reporter{C: c}
	go r.Feed()
	killer, err := r.Start(exampleValueSet{}, &DstatStyler{Period: 1, LinesBetweenHeaders: 0, Logger: log.New(os.Stdout, "example: ", 0)})
	if err != nil {
		return
	}
	for i := 0; i < 10; i++ {
		//c <- &exampleValueSet{A: 1, B: 1, Tps: 198273}
		c <- &exampleValueSet{A: 1, B: 1}
		time.Sleep(400 * time.Millisecond)
	}
	close(c)
	killer <- true
	// Output:
	// example: ----------- a ------------ ----- b -----
	// example:         iter        total |        total
	// example:         2.00            2 |            2
	// example:         1.00            4 |            4
	// example:         1.00            7 |            7
	// example:         0.50            9 |            9
}
