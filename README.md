# Socks

Socks is a templating library for Go. It is made to be simple
and easy to use. It provides easy ways to extend and include
templates, it comes with a custom bytecode expression evaluator, 
static evaluation and more nice-to-haves. 

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

    if err := s.Compile(map[string]any{
        "Server": "localhost",
    }); err != nil {
        panic(err)
    }

    s.AddGlobal("now", "2019-01-01")
	
    s.AddGlobal("currentTime", func() string {
        return time.Now().Format("2006-01-02 15:04:05")
    })
	
    result, err := s.ExecuteToString("templates/index.html", map[string]any{
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

## Elements

### Expression
Value of this expression will be printed to the template.
It will be sanitized if a sanitizer function is provided.
```html
{{ Users[0].CreatedAt.Format("2006-01-02") }}
```

### Raw expression
Value of this expression will be printed to the template without any sanitization.
```html
{! client.scripts !}
```

### Preprocessor statements
- `@extend`, `@slot` and `@define` are used for extending templates.
```html
<!--base.html-->
<html>
    <head>
        <title>{! title !}</title>
    </head>
    <body>
        @slot("content")
        <h2>404 – Content not found</h2>
        @endslot
    </body>
</html>
```
```html
@extend("base.html") <!-- extends the base.html template -->

@define("content")
    <!-- defines a block named content, 
         if omitted – template will fallback to default slot value 
         -->
    <h1>Hello World</h1>
@enddefine
```
- `@template`, `@slot` and `@define` are used for including templates.
```html
<!--header.html-->
<header>
    @slot("content") Fallback content @endslot
</header>
```
```html
@template("header.html")
    @define("content")
        <!-- defines a block named content, 
             if omitted – template will fallback to default slot value 
             -->
        <h1>Hello World</h1>
    @enddefine
@endtemplate
```

### Comment
Comment tag, the content of this tag will be ignored.
```html
{# This is a comment #}
<!-- this is an HTML comment and it won't be removed -->
```

### For statement

```html
@for(user in Users)
<p>{{ user.ID }} – {{ user.Name }}</p>
@endfor

@for(user in Users with i)
<p>{{ i }}: {{ user.Name }}</p>
@endfor
```

### If statement
```html
@if(len(Users) > 0)
<p>{{ Users[0].Name }}</p>
@endif
```

## Roadmap
- [ ] Runtime error handling
- [ ] Inline functions like `map`, `filter`, `reduce`, etc.