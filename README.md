# Extended brainfuck AKA clusterfuck

## Functions

### Define some function on register 2:

    >>{++++++++++}

### Call a function in register 2:

    >>!

### Recurse 3 times:

    +++{-[!]}!

### Endless recursion:

    {!}!

## Connect to peers

Yet to be done.

# Usage

	$ go build
	$ ./main < foo.bf | gofmt > foo.go
	$ go run foo.go
