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
        <title>Hello from {{ server }}</title>
        {! scripts !}
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
    s := socks.NewSocks(&socks.Options{
        Sanitizer: func(s string) string {
            // Sanitize your output here, everything that
            // goes through the {{ ... }} tag is going to be sanitized 
            return s
        },
    })
	
    if err := s.LoadTemplates("templates/*.html"); err != nil {
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

## Tags

> [!CAUTION]
> Nothing is escaped by default, so make sure to escape everything that comes from the user.
> This may change in the future.

### Expression tag
Expression tag, the result of the expression will be printed to the template.
It will be sanitized if a sanitizer function is provided.
```html
{{ Users[0].CreatedAt.Format("2006-01-02") }}
```

### Raw tag
Raw tag, the content of this tag will be printed to the template without any modification.
```html
{! client.scripts !}
```

### Preprocessor tag
Preprocessor tags, the result of the expression will be printed to the template.
```html
@extend("base.html") <!-- extends the base.html template -->
```
```html
@template("header.html") <!-- includes the header.html template -->
    @define("content")
        <!-- defines a block named content, 
             if omitted – template will fallback to default slot value 
             -->
        <h1>Hello World</h1>
    @enddefine
@endtemplate
```

### Comment tag
Comment tag, the content of this tag will be ignored.
```html
{# This is a comment #}
<!-- this is an HTML comment, but it won't be removed -->
```

### Execution tag
Execution tag, used for executing statements like `for` or `if` at runtime.
```html
@for(user, i in Users)
    <p>{{ i }}: {{ user.Name }}</p>
@endfor
```
```html
@if(len(Users) > 0)
    <p>{{ Users[0].Name }}</p>
@endif
```

## Roadmap
- [ ] Runtime error handling
- [ ] Constant folding 
- [ ] Add more tests
