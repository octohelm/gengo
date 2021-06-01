package gengo

import (
	"bytes"
	"fmt"
	gengotypes "github.com/go-courier/gengo/pkg/types"
	"io"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-courier/gengo/pkg/namer"
)

func Snippet(format string, args ...Args) func(s SnippetWriter) {
	return func(s SnippetWriter) {
		s.Do(format, args...)
	}
}

type SnippetWriter interface {
	io.Writer
	Do(format string, args ...Args)
	Dumper() *Dumper
}

type Args = map[string]interface{}

func NewSnippetWriter(w io.Writer, ns namer.NameSystems) SnippetWriter {
	sw := &snippetWriter{
		Writer: w,
		ns:     ns,
	}
	return sw
}

func createRender(ns namer.NameSystems) func(r func(s SnippetWriter)) string {
	return func(r func(s SnippetWriter)) string {
		b := bytes.NewBuffer(nil)
		r(NewSnippetWriter(b, ns))
		return b.String()
	}
}

type snippetWriter struct {
	io.Writer
	ns namer.NameSystems
}

func (s *snippetWriter) Dumper() *Dumper {
	if rawNamer, ok := s.ns["raw"]; ok {
		return NewDumper(rawNamer)
	}
	return nil
}

func (s *snippetWriter) Do(format string, args ...Args) {
	_, file, line, _ := runtime.Caller(1)

	tmpl, err := template.
		New(fmt.Sprintf("%s:%d", file, line)).
		Delims("[[", "]]").
		Funcs(createFuncMap(s.ns)).
		Parse(strings.TrimLeftFunc(format, func(r rune) bool {
			return r == '\n'
		}))

	if err != nil {
		panic(err)
	}

	finalArgs := Args{}

	for i := range args {
		a := args[i]
		for k := range a {
			finalArgs[k] = a[k]
		}
	}

	if err := tmpl.Execute(s.Writer, finalArgs); err != nil {
		panic(err)
	}
}

func createFuncMap(nameSystems namer.NameSystems) template.FuncMap {
	funcMap := template.FuncMap{}

	funcMap["id"] = createID(nameSystems)
	funcMap["render"] = createRender(nameSystems)
	funcMap["quote"] = strconv.Quote

	return funcMap
}

func createID(nameSystems namer.NameSystems) func(v interface{}) string {
	return func(v interface{}) string {
		switch x := v.(type) {
		case string:
			ref, err := gengotypes.ParseRef(x)
			if err != nil {
				return x
			}
			return nameSystems["raw"].Name(ref)
		case gengotypes.TypeName:
			return nameSystems["raw"].Name(x)
		default:
			panic("unspported")
		}
		return ""
	}
}
