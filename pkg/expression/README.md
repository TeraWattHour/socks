# Socks embedded expression language

## Literals
- `"string"` String
- `123` Integer
- `0.123` Float
- `true`, `false` Boolean
- `nil` Nil

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

#### Conditional
- `a ?: b` If _a_ is nil then _b_

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