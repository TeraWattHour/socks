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
    // Create a Socks struct, which will be used to render templates
	// The first argument is the directory where the templates are located
	// The second argument is a map of static values or functions that can be used only 
	// during template preprocessing
    s, err := socks.NewSocks("templates", map[string]interface{}{
		"server": "Frankfurt",
    })
	if err != nil {
		panic(err)
    }
	
	// Add a global function that can be used in all templates
	// Global functions can be overwritten by local functions
	s.AddGlobal("currentTime", func() string {
        return time.Now().Format("2006-01-02 15:04:05")
	})

    // Render the template
    result, err := s.Run("index.html", map[string]interface{}{
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