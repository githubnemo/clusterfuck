package parser

import (
	"reflect"
	"fmt"
)


type Node interface{
	Pos() int		// Position of first character of node
	End() int		// Position of character after the node
}


type Summarizable interface{
	Count() int
	Add()
}


type Encodable interface{
	Code() string
}


type BaseNode struct{
	start, end		int
}

func (b *BaseNode) Pos() int {
	return b.start
}

func (b *BaseNode) End() int {
	return b.end
}


type BaseCountable struct{
	count	int
}

func (b *BaseCountable) Count() int {
	return b.count
}

func (b *BaseCountable) Add() {
	b.count++
}


type PreambleNode struct{
	BaseNode
}

func (n *PreambleNode) Code() string {
	return `
package main

import "fmt"
import "os"

const REGISTERS = 100

func main() {
	registers := make([]byte, REGISTERS)
	currentIndex := 0

	// Program begin
`
}


type PostambleNode struct{
	BaseNode
}

func (n *PostambleNode) Code() string {
	return `
	// Program end
	// Flush stdout
	os.Stdout.Sync()
}
`
}


type IncNode struct{
	BaseNode
	BaseCountable		// How many increments are to be done
}

func (n *IncNode) Code() string {
	return fmt.Sprintf("registers[currentIndex] += %d\n", n.Count())
}


type DecNode struct{
	BaseNode
	BaseCountable		// How many decrements are to be done
}

func (n *DecNode) Code() string {
	return fmt.Sprintf("registers[currentIndex] -= %d\n", n.Count())
}


type PrevNode struct{
	BaseNode
	BaseCountable		// How many prev. shifts are to be done
}

func (n *PrevNode) Code() string {
	return fmt.Sprintf(`
	if currentIndex == 0 {
		currentIndex = REGISTERS-1
	} else {
		currentIndex -= %d
	}
`, n.Count())
}


type NextNode struct{
	BaseNode
	BaseCountable
}

func (n *NextNode) Code() string {
	return fmt.Sprintf("currentIndex = (currentIndex + %d) %% REGISTERS\n", n.Count())
}


type OutputNode struct{
	BaseNode
}

func (n *OutputNode) Code() string {
	return "fmt.Print(string(registers[currentIndex]))\n"
}


type InputNode struct{
	BaseNode
}

func (n *InputNode) Code() string {
	return `
	{
		_, err := fmt.Scanf("%c", &registers[currentIndex])

		if err != nil {
			fmt.Println("Keyscan Error:", err)
			return
		}

		fmt.Println(registers[currentIndex])
	}
`
}


type LoopNode struct{
	BaseNode
	Nodes	[]Node
}

func (n *LoopNode) Code() string {
	header := `{
		loopIndex := currentIndex
		for ; registers[loopIndex] > 0; {
`
	footer := `
		}
	}
`
	code := header

	for _, node := range n.Nodes {
		code += "\t\t\t" + node.(Encodable).Code()
	}

	return code + footer
}


type LoopOpenNode struct{
	BaseNode
}


type LoopCloseNode struct{
	BaseNode
}


type TokenList struct{
	Nodes		[]Node
	lastNode	Node
}

func (t *TokenList) Append(n Node) {
	if _, ok := n.(Summarizable); ok && reflect.TypeOf(t.lastNode) == reflect.TypeOf(n) {
		t.lastNode.(Summarizable).Add()
	} else {
		t.Nodes = append(t.Nodes, n)
		t.lastNode = n
	}
}



func Tokenize(s string) *TokenList {
	t := &TokenList{}

	t.Append(&PreambleNode{ BaseNode{0,0} })

	for i, c := range s {
		end := len(string(c))

		switch c {
			case '+':
				t.Append(&IncNode{ BaseNode{i, end}, BaseCountable{1} })
			case '-':
				t.Append(&DecNode{ BaseNode{i, end}, BaseCountable{1} })
			case '<':
				t.Append(&PrevNode{ BaseNode{i, end}, BaseCountable{1} })
			case '>':
				t.Append(&NextNode{ BaseNode{i, end}, BaseCountable{1} })
			case '[':
				t.Append(&LoopOpenNode{ BaseNode{i, end} })
			case ']':
				t.Append(&LoopCloseNode{ BaseNode{i, end} })
			case '.':
				t.Append(&OutputNode{ BaseNode{i, end} })
			case ',':
				t.Append(&InputNode{ BaseNode{i, end} })
		}
	}

	t.Append(&PostambleNode{ BaseNode{0,0} })

	return t
}


type ParseList struct{
	Nodes	[]Node
}

func (p *ParseList) Append(n Node) {
	p.Nodes = append(p.Nodes, n)
}


func parseLoop(t []Node) (*LoopNode, int, error) {
	loopNode := &LoopNode{}

	for i := 0; i < len(t); i++ {
		node := t[i]

		if _, ok := node.(*LoopOpenNode); ok {
			lnode, skip, err := parseLoop(t[i+1:])

			if err != nil {
				return nil, 0, err
			}

			node = lnode
			i += skip
		}

		if _, ok := node.(*LoopCloseNode); ok {
			return loopNode, i, nil
		}

		loopNode.Nodes = append(loopNode.Nodes, node)
	}

	return nil, 0, fmt.Errorf("Syntax error: No Loop Close found. Expected at pos. %d\n", t[len(t)-1].End())
}


func ParseTokens(t *TokenList) (*ParseList, error) {
	p := &ParseList{}

	for i := 0; i < len(t.Nodes); i++ {
		unknownNode := t.Nodes[i]

		switch unknownNode.(type) {
			case *LoopOpenNode:
				loopNode, skip, err := parseLoop(t.Nodes[i+1:])

				if err != nil {
					return nil, err
				}

				p.Append(loopNode)

				i += skip

			case Encodable:
				p.Append(unknownNode)
		}
	}

	return p, nil
}


func Parse(s string) (*ParseList, error) {
	t := Tokenize(s)

	return ParseTokens(t)
}


func Encode(t *ParseList) string{
	code := ""

	for _, node := range t.Nodes {
		if codable, ok := node.(Encodable); ok {
			code += "\t" + codable.Code()
		}
	}

	return code
}


