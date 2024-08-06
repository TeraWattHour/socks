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
    s := socks.New(&socks.Options{
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

### Escaped expression
Value of this expression will be printed to the template.
It will be sanitized if a sanitizer function is provided.
```html
{{ Users[0].CreatedAt.Format("2006-01-02") }}
```

### Unescaped expression
Value of this expression will be printed to the template without any sanitization.
```html
{{ raw(client.scripts) }}
```

### Preprocessor statements
```html
<!--base.html-->
<html>
    <head>
        <title>{{ title }}</title>
    </head>
    <body>
        <v-slot name="content">
            <h2>404 – Content not found</h2>
        </v-slot>
    </body>
</html>
```
```html
<v-component name="base.html">
    <div :slot="content">
        <h1>Hello World</h1>
    </div>
</v-component>
```
- `@template`, `@slot` and `@define` are used for including templates.
```html
<!--header.html-->
<header>
    <v-slot name="content">Fallback content</v-slot>
</header>
```
```html
<v-component name="header.html">
    <div :slot="content">
        <h1>Hello World</h1>
    </div>
</v-component>
```

### Loops

```html
<p :for="user in Users">{{ user.ID }} – {{ user.Name }}</p>

<p :for="user, i in Users">{{ i }}: {{ user.Name }}</p>
```

### Conditional statements
```html
<p :if="len(Users) == 1">{{ Users[0].Name }}</p>
<p :elif="len(Users) > 1">{{ Users[0].Name }} and {{ len(Users)-1 }} more...</p>
<p :else>No users</p>
```
