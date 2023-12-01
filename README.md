# Socks

Socks is a simple template engine for Go. It is made to be simple
and easy to use. It's not a code generator, so its performance is
not the best, but it's good enough for most use cases.

## Getting started

### Installation
```bash
go get github.com/terawatthour/socks
```

### Usage
```html
<!-- templates/index.html -->
<html>
    <head>
        <title>Hello from {$ server $}</title>
    </head>
    <body>
        <h1>Hello {{ name }}</h1>
        <p>The current time is {{ currentTime() }}</p>
    </body>
</html>
```

```go
package main

import (
    "fmt"
    "time"
    "github.com/terawatthour/socks"
)

func main() {
    s := socks.NewSocks()
	
    err := s.LoadTemplates("templates/*.html")
    if err != nil {
        panic(err)
    }

    if err := s.PreprocessTemplates(map[string]interface{}{
        "Server": "localhost",
    }); err != nil {
        panic(err)
    }

    s.AddGlobal("now", "2019-01-01")
	
    s.AddGlobal("currentTime", func() string {
        return time.Now().Format("2006-01-02 15:04:05")
    })
	
    result, err := s.Run("templates/index.html", map[string]interface{}{
        "name": "World",
        "currentTime": func() string {
            return time.Now().Format("01-02-2006 15:04:05")
        },
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result)
}
```

## Roadmap
- [x] Global functions
- [x] For loops
- [x] Template inheritance
- [x] Template embedding
- [x] If statements
- [x] Inline math/boolean operations
- [ ] Improve error messages
- [ ] Write proper tests
