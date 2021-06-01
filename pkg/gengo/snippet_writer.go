package gengo

import (
	"bytes"
	"fmt"
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

	funcMap := template.FuncMap{}

	for k := range s.ns {
		funcMap[k] = s.ns[k].Name
	}

	funcMap["render"] = createRender(s.ns)
	funcMap["quote"] = strconv.Quote

	tmpl, err := template.
		New(fmt.Sprintf("%s:%d", file, line)).
		Delims("[[", "]]").
		Funcs(funcMap).
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
