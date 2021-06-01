package gengo

import (
	"bytes"
	"github.com/onsi/gomega"
	"reflect"
	"testing"
)

func TestDumper_TypeLit(t *testing.T) {
	tt := gomega.NewWithT(t)

	d := New("", NewImportTracker())

	t.Run("TypeLit", func(t *testing.T) {
		tt.Expect("*bytes.Buffer").To(gomega.Equal(d.TypeLit(reflect.TypeOf(&bytes.Buffer{}))))
		tt.Expect("[]string").To(gomega.Equal(d.TypeLit(reflect.TypeOf([]string{}))))
		tt.Expect("map[string]string").To(gomega.Equal(d.TypeLit(reflect.TypeOf(map[string]string{}))))
	})

	t.Run("ValueLit", func(t *testing.T) {
		tt.Expect("&(bytes.Buffer{})").To(gomega.Equal(d.ValueLit(reflect.ValueOf(&(bytes.Buffer{})))))
		tt.Expect(`[]string{
"1",
"2",
}`).To(gomega.Equal(d.ValueLit(reflect.ValueOf([]string{"1", "2"}))))
	})
}
