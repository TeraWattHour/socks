# Socks embedded expression language

## Operators

### Infix

#### Arithmetic
- `+` Addition
- `-` Subtraction
- `*` Multiplication
- `/` Division
- `%` Modulo
- `**` Exponentiation
- `//` Floor division

#### Comparison
- `==` Equal
- `!=` Not equal
- `>` Greater than
- `>=` Greater than or equal
- `<` Less than
- `<=` Less than or equal

#### Logical
- `and` Logical and
- `or` Logical or
- `not` Negation

#### Sets
- `[...]` Array (always constant)
- `a in A` Is element _a_ present in _A_? (a ∈ A)
- `a not in A` Is element _a_ not present in _A_? (a ∉ A)

### Prefix
- `-` Negation
    - `-Numeric` → `Numeric * -1`
    - `-"Welcome"` → `"emocleW"`, reverses the string
    - `-[1, 2, 3]` → `[3, 2, 1]`, reverses the array
- `not`, `!` Negation
    - `not true` → `false`
    - `!false` → `true
    - `not 0.01` → `false`
    - `!0` → `true`