package gengo

import (
	"fmt"
	"io"
	"runtime"
	"text/template"

	"github.com/go-courier/gengo/pkg/namer"
)

type SnippetWriter struct {
	w           io.Writer
	funcMap     template.FuncMap
	left, right string
	err         error
}

func NewSnippetWriter(w io.Writer, ns namer.NameSystems) *SnippetWriter {
	sw := &SnippetWriter{
		w:       w,
		funcMap: template.FuncMap{},
	}

	for k := range ns {
		sw.funcMap[k] = ns[k].Name
	}

	return sw
}

func (s *SnippetWriter) Do(format string, args map[string]interface{}) error {
	_, file, line, _ := runtime.Caller(1)

	tmpl, err := template.
		New(fmt.Sprintf("%s:%d", file, line)).
		Funcs(s.funcMap).
		Parse(format)

	if err != nil {
		return err
	}

	return tmpl.Execute(s.w, args)
}
