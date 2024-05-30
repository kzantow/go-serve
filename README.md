# go-serve

Tiny API generator that matches your Go structs in the browser.

This is in a pre-alpha state.

Example:

```go
import "github.com/kzantow/go-serve"

// define structs normally
type Api struct {}

type Results struct {
	Name string
	...
}

func (a *Api) ProcessStuff(input string) []Results {
	...
}

func main() {
    // pass an instance of the struct to the handler and serve it up
    srv := serve.NewHandler("/api/").AddStruct(&api{})
    
    fmt.Println("Serving your go code at: http://localhost:9000/api/.js 🚀")
    serve.PanicOnErr(http.ListenAndServe(":9000", srv))
}
```

And get async callbacks available matching `<struct-name>.<method-name>` at global `Api.ProcessStuff`,
just use this from any browser code like:

```js
var results = await Api.ProcessStuff("some-input")
```

Also add functions directly:

```go
    // pass an instance of the struct to the handler and serve it up
    serve.NewHandler("/api/").
	    AddFunc("GlobalFunc", func(...) { ... })
	    AddFunc("With/Any/Nesting", func(...) { ... })
```

Available at:
```js
var global = await GlobalFunc()
var nested = await With.Any.Nesting()
```
