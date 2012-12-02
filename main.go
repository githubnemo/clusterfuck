package main

import (
	"./parser"
	"fmt"
	"os"
	"io/ioutil"
)

func max(a,b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a,b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	if false {
	//fmt.Printf( "%#v\n", parser.Tokenize("+++") )

	pd, errd := parser.Parse("+++[>+++[>+++<-]<-]>+.")
	//pd, errd := parser.Parse(",.")

	//fmt.Printf( "%#v (%s)\n", p, err )

	if errd != nil {
		fmt.Println(errd)
		return
	}

	fmt.Println( parser.Encode(pd) )

	} else {

	data, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		fmt.Printf("Reading from stdin failed: %s\n", err)
		os.Exit(1)
	}

	p, err := parser.Parse(string(data))

	if err != nil {
		fmt.Println(err)

		if pe, ok := err.(*parser.ParseError); ok {
			s, e := pe.FaultyNode.Pos(), pe.FaultyNode.End()
			t := 10
			fmt.Printf( "Details: %s\n", string( data[max(s-t,0) : min(e+t, len(data))] ) )
		}

		os.Exit(1)
	}

	fmt.Println( parser.Encode(p) )

	}
}
