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
	UseIO		bool
	UseFmt		bool
}

func (n *PreambleNode) Code() string {
	headers := `
import "os"
`

	start := `
package main
`

	body := `

const REGISTERS = 100

func main() {
	registers := make([]byte, REGISTERS)
	functions := make([]func(), REGISTERS)
	currentIndex := 0

	// Suppress unused warnings
	registers[0] = 0
	functions[0] = nil

	// Program begin
`

	start += headers

	if n.UseIO {
		start += "import \"io\"\n"
	}

	if n.UseFmt {
		start += "import \"fmt\"\n"
	}

	return start + body
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
			if err == io.EOF {
				registers[currentIndex] = 0
			} else {
				fmt.Println("Keyscan Error:", err)
				return
			}
		}
	}
`
}


type LoopNode struct{
	BaseNode
	Nodes	[]Node
}

func (n *LoopNode) Code() string {
	header := `{
		for ; registers[currentIndex] > 0; {
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


// Represent a function definition.
type FunctionNode struct{
	BaseNode
	Nodes	[]Node
}

func (n *FunctionNode) Code() string {
	header := `
	functions[currentIndex] = func() {
`
	footer := `
	}
`
	code := header

	for _, node := range n.Nodes {
		code += "\t\t\t" + node.(Encodable).Code()
	}

	return code + footer

}


type FuncOpenNode struct{
	BaseNode
}


type FuncCloseNode struct{
	BaseNode
}


// Execution of a function defined on a index
type FuncExecNode struct {
	BaseNode
}

func (n *FuncExecNode) Code() string {
	return `
	if functions[currentIndex] != nil {
		functions[currentIndex]()
	}
`
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

	t.Append(&PreambleNode{ BaseNode{0,0}, false, false })

	for i, c := range s {
		end := i + len(string(c))

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
			case '{':
				t.Append(&FuncOpenNode{ BaseNode{i, end} })
			case '}':
				t.Append(&FuncCloseNode{ BaseNode{i, end} })
			case '!':
				t.Append(&FuncExecNode{ BaseNode{i, end} })
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


type ParseError struct {
	FaultyNode	Node
	Message		string
}


func (p *ParseError) Error() string {
	return p.Message
}


func parseError(n Node, msg string) *ParseError {
	return &ParseError{
		n,
		fmt.Sprintf("Error: %s, Position: %d - %d\n", msg, n.Pos(), n.End()),
	}
}


func ParseTokens(t *TokenList, nesting int) (*ParseList, int, error) {
	p := &ParseList{}

	for i := 0; i < len(t.Nodes); i++ {
		unknownNode := t.Nodes[i]

		switch unknownNode.(type) {
			case *LoopOpenNode:
				p2, skip, err := ParseTokens(&TokenList{t.Nodes[i+1:], nil}, nesting+1)

				if err != nil {
					return nil, 0, err
				}

				p.Append(&LoopNode{
					BaseNode{
						p2.Nodes[0].Pos(),
						p2.Nodes[len(p2.Nodes)-1].End(),
					},
					p2.Nodes,
				})

				i += skip

			case *LoopCloseNode:
				if nesting == 0 {
					return nil, 0, parseError(unknownNode, "Loop closed while not open")
				}

				// +1 for the skipped LoopCloseNode
				return p, i+1, nil

			case *FuncOpenNode:
				p2, skip, err := ParseTokens(&TokenList{t.Nodes[i+1:], nil}, nesting+1)

				if err != nil {
					return nil, 0, err
				}

				p.Append(&FunctionNode{
					BaseNode{
						p2.Nodes[0].Pos(),
						p2.Nodes[len(p2.Nodes)-1].End(),
					},
					p2.Nodes,
				})

				i += skip

			case *FuncCloseNode:
				if nesting == 0 {
					return nil, 0, parseError(unknownNode, "Func closed while not open")
				}

				return p, i+1, nil

			case *InputNode:
				p.Nodes[0].(*PreambleNode).UseIO = true
				p.Append(unknownNode)

			case *OutputNode:
				p.Nodes[0].(*PreambleNode).UseFmt = true
				p.Append(unknownNode)

			case Encodable:
				p.Append(unknownNode)
		}
	}

	return p, 0, nil
}


func Parse(s string) (*ParseList, error) {
	t := Tokenize(s)

	p, _, err := ParseTokens(t, 0)

	return p, err
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


