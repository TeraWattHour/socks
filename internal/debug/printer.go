package debug

import (
	"fmt"
	"github.com/terawatthour/socks/parser"
	"reflect"
	"strings"
)

func PrintPrograms(label string, programs []parser.Program) {
	fmt.Printf("programs (%s):\n", label)
	level := 0
	for _, program := range programs {
		if program.Kind() == "end" {
			level--
		}
		fmt.Printf("%s%s\n", strings.Repeat("  ", level), program)
		if reflect.ValueOf(program).Kind() == reflect.Ptr && reflect.Indirect(reflect.ValueOf(program)).FieldByName("EndStatement").IsValid() {
			level++
		}
	}
	fmt.Println("End programs")
}
