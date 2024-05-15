// include this on a cmdline build only
//go:build cmdline

package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var choice string
	var state string
	var year string
	var begin string
	var end string
	var html bool

	flag.StringVar(&choice, "choice", "", "choice (1 = year, 2 = state, 3 = state/year range)")
	flag.StringVar(&state, "state", "", "state (for choice = 2 or 3)")
	flag.StringVar(&year, "year", "", "year (for choice = 1)")
	flag.StringVar(&begin, "begin", "", "beginning of year range (for choice = 3)")
	flag.StringVar(&end, "end", "", "end of year range (for choice = 3)")
	flag.BoolVar(&html, "html", false, "output html instead of text")
	flag.Parse()

	req := lookupRequest{choice: choice, state: state, year: year, begin: begin, end: end}
	res := lookup(req)

	if res.err != nil {
		fmt.Printf("ERROR: %s\n", res.err.Error())
		os.Exit(1)
	}

	var buf string
	var err error

	if html == true {
		buf, err = res.toHTML()
	} else {
		buf, err = res.toText()
	}

	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("%s", buf)
}
