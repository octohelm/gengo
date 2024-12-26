package camelcase

import (
	"fmt"
	"reflect"
	"testing"
)

func ExampleCases() {
	s := "vimRPCPluginS3"
	fmt.Println(UpperCamelCase(s))
	fmt.Println(LowerCamelCase(s))
	fmt.Println(UpperKebabCase(s))
	fmt.Println(LowerKebabCase(s))
	fmt.Println(UpperSnakeCase(s))
	fmt.Println(LowerSnakeCase(s))
	// Output:
	// VimRpcPluginS3
	// vimRpcPluginS3
	// VIM-RPC-PLUGIN-S3
	// vim-rpc-plugin-s3
	// VIM_RPC_PLUGIN_S3
	// vim_rpc_plugin_s3
}

func TestSplit(t *testing.T) {
	for _, c := range [][]string{
		{""},
		{"S3", "S3"},
		{"lowercase", "lowercase"},
		{"Class", "Class"},
		{"MyClass", "My", "Class"},
		{"MyC", "My", "C"},
		{"HTML", "HTML"},
		{"ID", "ID"},
		{"PDFLoader", "PDF", "Loader"},
		{"AString", "A", "String"},
		{"SimpleXMLParser", "Simple", "XML", "Parser"},
		{"vimRPCPlugin", "vim", "RPC", "Plugin"},
		{"GL11Version", "GL11", "Version"},
		{"99Bottles", "99", "Bottles"},
		{"May5", "May5"},
		{"BFG9000", "BFG9000"},
		{"BöseÜberraschung", "Böse", "Überraschung"},
		{"Two  spaces", "Two", "  ", "spaces"},
		{"BadUTF8\xe2\xe2\xa1", "BadUTF8\xe2\xe2\xa1"},
	} {
		t.Run(c[0], func(t *testing.T) {
			ret := Split(c[0])
			expect := c[1:]

			if !reflect.DeepEqual(ret, expect) {
				t.Fatalf("expect %v, but got %v", expect, ret)
			}
		})
	}
}
