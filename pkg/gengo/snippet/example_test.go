package snippet_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/octohelm/gengo/pkg/gengo/internal"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
	"github.com/octohelm/gengo/pkg/namer"
)

func render(s snippet.Snippet) string {
	d := internal.NewDumper(namer.NewRawNamer("github.com/octohelm/gengo/pkg/gengo/snippet", namer.NewDefaultImportTracker()))
	ctx := internal.DumperContext.Inject(context.Background(), d)

	var b strings.Builder
	for part := range snippet.Fragments(ctx, s) {
		b.WriteString(part)
	}
	return b.String()
}

func ExampleT() {
	s := snippet.T("type @Name struct { Value @Value }\n",
		snippet.IDArg("Name", "User"),
		snippet.IDArg("Value", "string"),
	)

	fmt.Print(render(s))

	// Output:
	// type User struct { Value string }
}

func ExampleSprintf() {
	s := snippet.Sprintf("return %v\n", "ok")
	fmt.Print(render(s))

	// Output:
	// return "ok"
}

func ExampleComment() {
	fmt.Print(render(snippet.Comment("first\nsecond")))

	// Output:
	// // first
	// // second
}
