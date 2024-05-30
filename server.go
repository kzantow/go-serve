package serve

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
)

type Handler interface {
	http.Handler
	AddFunc(path string, obj any) Handler
	AddStruct(obj any) Handler
}

// NewHandler takes objects and exposes each method as a simple POST-based API handler as well as generating javascript
func NewHandler(path string) Handler {
	srv := &server{
		ApiPath:   path,
		JSPath:    ".js",
		JSFunc:    "goServePost",
		Endpoints: map[string]*endpoint{},
	}
	return srv
}

type endpoint struct {
	Path    string
	Params  []string
	Handler func(w http.ResponseWriter, r *http.Request) error
}

type server struct {
	ApiPath   string
	JSPath    string
	JSFunc    string
	Endpoints map[string]*endpoint
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.RequestURI[len(s.ApiPath):]
	switch path {
	case s.JSPath:
		w.Header().Add("Accept", "*/*")
		w.Header().Add("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		s.generateJS(w)
	default:
		e := s.Endpoints[path]
		if e == nil {
			http.Error(w, fmt.Sprintf("'%s' not found", path), http.StatusNotFound)
			return
		}
		s.invoke(e, w, r)
	}
}

func (s *server) AddStruct(obj any) Handler {
	val := reflect.ValueOf(obj)
	typ := val.Type()
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if !m.IsExported() {
			continue
		}
		s.addEndpoint(fmt.Sprintf("%s/%s", baseType(typ).Name(), m.Name), val.Method(i))
	}
	return s
}

func (s *server) AddFunc(path string, fn any) Handler {
	val := reflect.ValueOf(fn)
	s.addEndpoint(path, val)
	return s
}

func (s *server) addEndpoint(path string, fn reflect.Value) {
	typ := fn.Type()
	e := &endpoint{
		Path: path,
	}
	for p := 0; p < typ.NumIn(); p++ {
		in := typ.In(p)
		e.Params = append(e.Params, in.Name())
	}
	e.Handler = s.makeHandlerFunc(fn)
	s.Endpoints[path] = e
}

func (s *server) EndpointList() []*endpoint {
	return SortedMapValues(s.Endpoints)
}

func (s *server) generateJS(w io.Writer) {
	root := endpointNode{}
	for k, v := range s.Endpoints {
		parts := regexp.MustCompile("[./]").Split(k, -1)
		root.getNode(parts).Endpoint = v
	}
	root.WriteJS(s, "", func(strings ...string) {
		for _, str := range strings {
			_ = GetOrPanic(w.Write([]byte(str)))
		}
	})
}

func (s *server) makeHandlerFunc(fn reflect.Value) func(w http.ResponseWriter, r *http.Request) error {
	typ := fn.Type()
	if typ.NumOut() == 0 || IsErrorType(typ.Out(0)) || typ.NumOut() > 2 {
		panic(fmt.Errorf("must have exactly one non-error return value: %#v", fn))
	}
	if typ.NumOut() == 2 && !IsErrorType(typ.Out(1)) {
		panic(fmt.Errorf("second arg must be error type: %#v", fn))
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		var in []reflect.Value
		if typ.NumIn() > 0 {
			var decoded map[string]any
			PanicOnErr(json.NewDecoder(r.Body).Decode(&decoded))
			args, ok := decoded["Args"].([]any)
			if !ok {
				return fmt.Errorf("unable to get args; got: %#v", decoded)
			}
			if len(args) != typ.NumIn() {
				return fmt.Errorf("wrong number of args, expected: %v", typ.NumIn())
			}
			for i := 0; i < typ.NumIn(); i++ {
				argT := typ.In(i)
				inst := reflect.New(argT).Interface()
				j, _ := json.Marshal(args[i])
				PanicOnErr(json.Unmarshal(j, inst))
				in = append(in, reflect.ValueOf(inst).Elem())
			}
		}
		out := fn.Call(in)
		PanicOnErr(extractError(out))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		val := extractValue(out)
		return json.NewEncoder(w).Encode(val)
	}
}

func (s *server) invoke(e *endpoint, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			http.Error(w, fmt.Sprintf("%v", e), http.StatusInternalServerError)
		}
	}()
	PanicOnErr(e.Handler(w, r))
}

func extractError(out []reflect.Value) error {
	if len(out) > 0 {
		last := out[len(out)-1]
		if last.CanInterface() {
			val := last.Interface()
			err, _ := val.(error)
			return err
		}
	}
	return nil
}

func extractValue(out []reflect.Value) any {
	if len(out) > 0 {
		last := out[0]
		if last.CanInterface() {
			return last.Interface()
		}
	}
	return nil
}

func baseType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}
