package expression

import "fmt"

func printChunk(chunk Chunk) {
	for i := 0; i < len(chunk.Instructions); i++ {
		switch chunk.Instructions[i] {
		case OpGet:
			fmt.Println("OpGet", chunk.Constants[chunk.Instructions[i+1]])
			i++
		case OpConstant:
			fmt.Println("OpConstant", chunk.Constants[chunk.Instructions[i+1]])
			i++
		case OpBuiltin1:
			fmt.Println("OpBuiltin1", builtinNames[chunk.Instructions[i+1]])
			i++
		case OpAdd:
			fmt.Println("OpAdd")
		case OpCall:
			fmt.Println("OpCall")
		default:
			fmt.Println("unknown", chunk.Instructions[i])
		}
	}
}
