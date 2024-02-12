package debug

import (
	"fmt"
	"reflect"
	"strings"
)

func PrintPrograms(label string, programs any) {
	val := reflect.ValueOf(programs)
	fmt.Printf("programs (%s):\n", label)
	indents := make([]int, 0)
	for i := 0; i < val.Len(); i++ {
		program := val.Index(i).Interface()
		if len(indents) > 0 {
			fmt.Print(strings.Repeat(" ", 2*len(indents)+2*(len(indents)-1)), "└─–")
			for i := len(indents) - 1; i >= 0; i-- {
				indents[i] -= 1
				if indents[i] == 0 {
					indents = indents[:i]
				}
			}
		}
		fmt.Print(program)
		if reflect.TypeOf(program).Kind() == reflect.String {
			fmt.Println()
			continue
		}
		programsField := reflect.Indirect(reflect.ValueOf(program)).FieldByName("Programs")
		if programsField.IsValid() {
			indents = append(indents, int(programsField.Int()))
		}
		fmt.Println()
	}
	fmt.Println("End programs")
}
