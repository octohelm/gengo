package snippet

import (
	"context"
	"errors"
	gotypes "go/types"
	"iter"
	"reflect"
	"strings"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo/internal"
	"github.com/octohelm/gengo/pkg/namer"
	gengotypes "github.com/octohelm/gengo/pkg/types"
)

type DemoItem struct {
	Name string
}

func renderSnippet(s Snippet) string {
	d := internal.NewDumper(namer.NewRawNamer("github.com/octohelm/gengo/pkg/gengo/snippet", namer.NewDefaultImportTracker()))
	ctx := internal.DumperContext.Inject(context.Background(), d)

	var b strings.Builder

	for part := range Fragments(ctx, s) {
		b.WriteString(part)
	}

	return b.String()
}

func TestSnippetAPI(t *testing.T) {
	t.Run("Fragments 与 Func", func(t *testing.T) {
		s := Func(func(ctx context.Context) iter.Seq[string] {
			return func(yield func(string) bool) {
				_ = yield("hello")
				_ = yield(" ")
				_ = yield("world")
			}
		})

		Then(t, "应按顺序展开片段",
			Expect(renderSnippet(s), Equal("hello world")),
		)

		Then(t, "nil snippet 应输出空串",
			Expect(renderSnippet(nilSnippet{}), Equal("")),
		)
	})

	t.Run("Snippets", func(t *testing.T) {
		snippets := Snippets(func(yield func(Snippet) bool) {
			_ = yield(Block("a"))
			_ = yield(Comment(""))
			_ = yield(Block("b"))
		})

		Then(t, "应跳过空片段并顺序拼接",
			Expect(renderSnippet(snippets), Equal("ab")),
		)

		Then(t, "String 当前应返回空串",
			Expect(snippets.String(), Equal("")),
		)
	})

	t.Run("Comment", func(t *testing.T) {
		Then(t, "多行注释应逐行输出",
			Expect(renderSnippet(Comment("first\nsecond")), Equal("// first\n// second")),
		)
	})

	t.Run("Block", func(t *testing.T) {
		Then(t, "应原样输出内容",
			Expect(renderSnippet(Block("x := 1")), Equal("x := 1")),
		)
	})

	t.Run("Value", func(t *testing.T) {
		Then(t, "基础值应输出字面量",
			Expect(renderSnippet(Value("demo")), Equal(`"demo"`)),
		)
	})

	t.Run("GoDirective", func(t *testing.T) {
		Then(t, "应输出 go 指令",
			Expect(renderSnippet(GoDirective("embed", "demo.txt")), Equal("//go:embed demo.txt")),
		)

		Then(t, "空 directive 应输出空串",
			Expect(renderSnippet(GoDirective("", "demo.txt")), Equal("")),
		)

		Then(t, "空参数应被跳过",
			Expect(renderSnippet(GoDirective("embed", "", "demo.txt")), Equal("//go:embed demo.txt")),
		)
	})

	t.Run("ID", func(t *testing.T) {
		Then(t, "同包类型引用应去掉包前缀",
			Expect(renderSnippet(ID(gengotypes.Ref("github.com/octohelm/gengo/pkg/gengo/snippet", "DemoItem"))), Equal("DemoItem")),
		)

		Then(t, "外部包类型引用应带本地名",
			Expect(renderSnippet(ID(gengotypes.Ref("fmt", "Stringer"))), Equal("fmt.Stringer")),
		)
	})

	t.Run("PkgExpose 系列", func(t *testing.T) {
		Then(t, "PkgExpose 应输出指定类型名",
			Expect(renderSnippet(PkgExpose("github.com/octohelm/gengo/pkg/gengo/snippet", "DemoItem")), Equal("DemoItem")),
		)

		Then(t, "PkgExposeOf 应处理指针实例",
			Expect(renderSnippet(PkgExposeOf(&DemoItem{})), Equal("DemoItem")),
		)

		Then(t, "PkgExposeFor 应支持覆盖名称",
			Expect(renderSnippet(PkgExposeFor[DemoItem]("AliasItem")), Equal("AliasItem")),
		)

		Then(t, "PkgExposeFor 无覆盖名称时应输出真实类型名",
			Expect(renderSnippet(PkgExposeFor[DemoItem]()), Equal("DemoItem")),
		)

		Then(t, "PkgExposeOf 遇到函数类型应 panic",
			ExpectMust(func() error {
				panicked := false
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()

				_ = renderSnippet(PkgExposeOf(func() {}))

				if !panicked {
					return errors.New("expected panic")
				}
				return nil
			}),
		)
	})

	t.Run("Sprintf", func(t *testing.T) {
		Then(t, "应同时支持 %v 与 %T",
			Expect(renderSnippet(Sprintf("%T = %v", reflect.TypeFor[DemoItem](), 1)), Equal("DemoItem = 1")),
		)

		Then(t, "Snippet 参数应按自身渲染",
			Expect(renderSnippet(Sprintf("%T = %v", Block("DemoValue"), Block("42"))), Equal("DemoValue = 42")),
		)

		t.Run("缺少参数时应 panic", func(t *testing.T) {
			ExpectMust(func() error {
				panicked := false
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()

				_ = renderSnippet(Sprintf("%v %v", 1))
				if !panicked {
					return errors.New("expected panic")
				}
				return nil
			})
		})

		t.Run("不支持的格式符应 panic", func(t *testing.T) {
			ExpectMust(func() error {
				panicked := false
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()

				_ = renderSnippet(Sprintf("%q", "x"))
				if !panicked {
					return errors.New("expected panic")
				}
				return nil
			})
		})
	})

	t.Run("模板参数", func(t *testing.T) {
		s := T("var @name = @value", IDArg("name", gengotypes.Ref("github.com/octohelm/gengo/pkg/gengo/snippet", "DemoItem")), ValueArg("value", "demo"))

		Then(t, "命名参数应被替换",
			Expect(renderSnippet(s), Equal(`var DemoItem = "demo"`)),
		)

		Then(t, "Args 也应可作为模板参数集合",
			Expect(renderSnippet(T("@left + @right", Args{
				"left":  Value(1),
				"right": Value(2),
			})), Equal("1 + 2")),
		)

		Then(t, "绑定为 nil 的参数应被跳过输出",
			Expect(renderSnippet(T("@left @right", Arg("left", Block("x")), Arg("right", Value(nil)))), Equal("x ")),
		)

		t.Run("缺失命名参数时应 panic", func(t *testing.T) {
			ExpectMust(func() error {
				panicked := false
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()

				_ = renderSnippet(T("@left @right", Arg("left", Block("x"))))
				if !panicked {
					return errors.New("expected panic")
				}
				return nil
			})
		})
	})

	t.Run("IsNil", func(t *testing.T) {
		Then(t, "空 Block 应为 nil",
			Expect(Block("").IsNil(), Be(cmp.True())),
		)

		Then(t, "Value(nil) 应为 nil",
			Expect(Value(nil).IsNil(), Be(cmp.True())),
		)

		Then(t, "Func 创建的 snippet 不应为 nil",
			Expect(Func(func(ctx context.Context) iter.Seq[string] {
				return func(yield func(string) bool) {}
			}).IsNil(), Be(cmp.False())),
		)

		Then(t, "空模板应为 nil",
			Expect(T("").IsNil(), Be(cmp.True())),
		)
	})

	t.Run("ID 其他分支", func(t *testing.T) {
		Then(t, "非引用字符串应原样输出",
			Expect(renderSnippet(ID("plain_text")), Equal("plain_text")),
		)

		Then(t, "引用字符串应按类型名输出",
			Expect(renderSnippet(ID("fmt.Stringer")), Equal("fmt.Stringer")),
		)

		Then(t, "reflect.Type 应输出类型字面量",
			Expect(renderSnippet(ID(reflect.TypeFor[*DemoItem]())), Equal("*DemoItem")),
		)

		Then(t, "types.Type 应输出类型字面量",
			Expect(renderSnippet(ID(gotypes.NewPointer(gotypes.Typ[gotypes.String]))), Equal("*string")),
		)

		t.Run("别名类型应按别名分支输出", func(t *testing.T) {
			u := MustValue(t, func() (*gengotypes.Universe, error) {
				return gengotypes.Load([]string{"github.com/octohelm/gengo/pkg/gengo/snippet/testdata/alias"})
			})

			p := u.Package("github.com/octohelm/gengo/pkg/gengo/snippet/testdata/alias")
			timeAlias := p.Type("TimeAlias")

			Then(t, "alias 类型应存在",
				Expect(timeAlias, Be(cmp.NotNil[*gotypes.TypeName]())),
			)

			Then(t, "应输出别名标识符",
				Expect(renderSnippet(ID(timeAlias.Type())), Equal("alias.TimeAlias")),
			)
		})

		t.Run("不支持的类型应 panic", func(t *testing.T) {
			ExpectMust(func() error {
				panicked := false
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()

				_ = renderSnippet(ID(1))
				if !panicked {
					return errors.New("expected panic")
				}
				return nil
			})
		})
	})

	t.Run("迭代中断", func(t *testing.T) {
		ctx := internal.DumperContext.Inject(context.Background(), internal.NewDumper(namer.NewRawNamer("github.com/octohelm/gengo/pkg/gengo/snippet", namer.NewDefaultImportTracker())))

		t.Run("Fragments 遇到 false 时应停止", func(t *testing.T) {
			count := 0
			for part := range Fragments(ctx, Block("abc")) {
				_ = part
				count++
				break
			}

			Then(t, "至少应迭代一次",
				Expect(count, Equal(1)),
			)
		})

		t.Run("Block.Frag 可直接被提前中断", func(t *testing.T) {
			count := 0
			Block("abc").Frag(ctx)(func(s string) bool {
				_ = s
				count++
				return false
			})

			Then(t, "应只收到一次回调",
				Expect(count, Equal(1)),
			)
		})

		t.Run("Func.Frag 可直接被提前中断", func(t *testing.T) {
			count := 0
			Func(func(ctx context.Context) iter.Seq[string] {
				return func(yield func(string) bool) {
					if !yield("a") {
						return
					}
					_ = yield("b")
				}
			}).Frag(ctx)(func(s string) bool {
				_ = s
				count++
				return false
			})

			Then(t, "应只收到第一次回调",
				Expect(count, Equal(1)),
			)
		})
	})
}

type nilSnippet struct{}

func (nilSnippet) IsNil() bool {
	return true
}

func (nilSnippet) Frag(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {}
}
