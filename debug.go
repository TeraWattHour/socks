package socks

import (
	"fmt"
	"reflect"
	"strings"
)

func dumpStatements(label string, statements []Statement) {
	fmt.Printf("statements (%s):\n", label)
	level := 0
	for _, program := range statements {
		if program.Kind() == "end" {
			level--
		}
		fmt.Printf("%s%s\n", strings.Repeat("  ", level), program)
		if reflect.ValueOf(program).Kind() == reflect.Ptr && reflect.Indirect(reflect.ValueOf(program)).FieldByName("EndStatement").IsValid() {
			level++
		}
	}
	fmt.Println("End statements")
}
