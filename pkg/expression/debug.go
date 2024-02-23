package expression

import "fmt"

func dumpChunk(chunk Chunk) {
	for i := 0; i < len(chunk.Instructions); i++ {
		symbol := opcodesLookup[chunk.Instructions[i]]
		fmt.Printf("%04d %14s", i, symbol)
		switch chunk.Instructions[i] {
		case OpConstant:
			fmt.Printf(" | %v\n", chunk.Constants[chunk.Instructions[i+1]])
			i++
		case OpJmp:
			fmt.Printf(" | jump: %v\n", chunk.Instructions[i+1])
			i++
		case OpTernary:
			fmt.Printf(" | jumpIfFalse: %v\n", chunk.Instructions[i+1])
			i++
		case OpChain:
			fmt.Printf(" | %v\n", chunk.Constants[chunk.Instructions[i+1]])
			i++
		case OpElvis:
			fmt.Printf(" | jump: %v\n", chunk.Instructions[i+1])
			i++
		case OpOptionalChain:
			fmt.Printf(" | %v | jump: %v\n", chunk.Constants[chunk.Instructions[i+2]], chunk.Instructions[i+1])
			i += 2
		case OpCall:
			fmt.Printf(" | args: %v\n", chunk.Instructions[i+1])
			i++
		case OpGet:
			fmt.Printf(" | %v\n", chunk.Constants[chunk.Instructions[i+1]])
			i++
		case OpBuiltin1:
			fmt.Printf(" | %s\n", builtinNames[chunk.Instructions[i+1]])
			i++
		case OpBuiltin2:
			fmt.Printf(" | %v\n", builtinNames[numBuiltinsOne+chunk.Instructions[i+1]])
			i++
		case OpBuiltin3:
			fmt.Printf(" | %v\n", builtinNames[numBuiltinsOne+numBuiltinsTwo+chunk.Instructions[i+1]])
			i++
		case OpCodeCount:
			panic("unreachable")
		default:
			// ops that don't take arguments
			fmt.Printf("\n")
		}
	}
}
