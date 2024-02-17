from typing import List


def insert_section(filename: str, section: str, content: str):
    with open(filename, 'r') as f:
        lines = f.readlines()
        lines = lines[:lines.index(f'// BEGIN {section}\n')+1] + [content] + lines[lines.index(f'// END {section}\n'):]

        joined = "".join(lines)

    with open(filename, 'w') as f:
        f.write(joined)


def generate_numeric_casts():
    types = ['int', 'int8', 'int16', 'int32', 'int64', 'uint', 'uint8', 'uint16',
             'uint32', 'uint64', 'uintptr', 'float32', 'float64']
    result = ''
    for to_type in types:
        result += f'func cast{to_type.capitalize()}(val any) any {{\n   switch val := val.(type) {{\n'
        for from_type in types:
            if from_type == to_type:
                result += f'   case {from_type}:\n      return val\n'
            else:
                result += f'   case {from_type}:\n      return {to_type}(val)\n'
        result += f'   }}\n   panic(fmt.Sprintf("cannot cast %s to {to_type}", reflect.TypeOf(val)))\n}}\n\n'

    insert_section('pkg/expression/builtins.go', 'CASTS', result)


def generate_numeric_binary_ops():
    types = ['string', 'int', 'int8', 'int16', 'int32', 'int64', 'uint', 'uint8', 'uint16',
             'uint32', 'uint64', 'uintptr', 'float32', 'float64']

    result = ''
    for op in [('Addition', '+'), ('Subtraction', '-'), ('Multiplication', '*'), ('Division', '/'), ('Modulo', '%')]:
        result += f'func binary{op[0]}(a, b any) any {{\n   switch a := a.(type) {{\n'
        for t in types:
            if 'float' in t and op[0] == 'Modulo' or t == 'string' and op[0] != 'Addition':
                continue
            result += f'   case {t}:\n      switch b := b.(type) {{\n'
            result += f'      case {t}:\n         return a {op[1]} b\n      }}\n'
        result += f'   }}\n   panic(fmt.Sprintf("invalid operation: %v {op[1]} %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b)))\n}}\n'

    result += 'func binaryExponentiation(a, b any) any {\n   switch a := a.(type) {\n'
    for t in types:
        if t == 'string':
            continue
        result += f'   case {t}:\n      switch b := b.(type) {{\n'
        if t == 'float64':
            result += f'      case {t}:\n         return math.Pow(a, b)\n      }}\n'
        else:
            result += f'      case {t}:\n         return {t}(math.Pow(float64(a), float64(b)))\n      }}\n'

    result += f'   }}\n   panic(fmt.Sprintf("invalid operation: %v {op[1]} %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b)))\n}}\n'

    with open('pkg/expression/ops.go', 'r') as f:
        lines = f.readlines()
        lines = lines[:lines.index('// BEGIN BINARY\n')+1] + [result] + lines[lines.index('// END BINARY\n'):]

        joined = "".join(lines)

    with open('pkg/expression/ops.go', 'w') as f:
        f.write(joined)


def generate_equality_ops():
    types = ['string', 'int', 'int8', 'int16', 'int32', 'int64', 'uint', 'uint8', 'uint16',
             'uint32', 'uint64', 'uintptr', 'float32', 'float64']

    result = ''
    for op in [('LessThan', '<'), ('LessThanEqual', '<='), ('GreaterThan', '>'), ('GreaterThanEqual', '>=')]:
        result += f'func binary{op[0]}(a, b any) any {{\n   switch a := a.(type) {{\n'
        for t in types:
            result += f'   case {t}:\n      switch b := b.(type) {{\n'
            result += f'      case {t}:\n         return a {op[1]} b\n      }}\n'
        result += f'   }}\n   return reflect.DeepEqual(a, b)\n}}\n'

    with open('pkg/expression/ops.go', 'r') as f:
        lines = f.readlines()
        lines = lines[:lines.index('// BEGIN EQUALITY\n')+1] + [result] + lines[lines.index('// END EQUALITY\n'):]

        joined = "".join(lines)

    with open('pkg/expression/ops.go', 'w') as f:
        f.write(joined)


def generate_compiler_debug():
    with open('./opcodes.go') as f:
        original: List[str] = f.readlines()
        lines = original[original.index("// BEGIN OPCODES\n")+2:original.index("// END OPCODES\n")-2]
        lines = list(map(lambda x: x.strip().replace(" = iota", "")[2:], filter(lambda x: len(x)>2, lines)))
        lookup_map = 'var opcodesLookup = map[int]string {\n'
        for (idx, opcode) in enumerate(lines):
            lookup_map += f'	Op{opcode}: "{opcode.upper()}",\n'
        lookup_map += '}\n'
        result = original[:original.index("// BEGIN LOOKUP\n")+1] + [lookup_map] + original[original.index("// END LOOKUP\n"):]
    with open("./opcodes.go", 'w') as f:
        f.write("".join(result))

generate_compiler_debug()
