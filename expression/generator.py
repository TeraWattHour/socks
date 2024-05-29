def generate_binary_ops():
    types = ['int', 'int8', 'int16', 'int32', 'int64', 'uint', 'uint8', 'uint16', 'uint32', 'uint64', 'uintptr', 'float32', 'float64']
    for op in [('Addition', '+'), ('Subtraction', '-'), ('Multiplication', '*'), ('Division', '/'), ('Modulus', '%'), ('Equal', '=='), ('NotEqual', '!='), ('Less', '<'), ('LessEqual', '<='), ('Greater', '>'), ('GreaterEqual', '>=')]:
        template = f"""func operation{op[0]}(a, b any) any {{
    switch a := a.(type) {{"""
        for type in types:
            template += f"""\n    case {type}:
        switch b := b.(type) {{
        case {type}:
            return a {op[1]} b
        }}"""
        template += f"""\n    }}
    return fmt.Errorf("invalid operation: %v {op[1]} %v (mismatched types %T and %T)", a, b, a, b)
}}\n"""
        yield template

for x in generate_binary_ops():
    print(x)
