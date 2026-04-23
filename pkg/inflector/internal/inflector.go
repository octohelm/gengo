package internal

// Defaults 保存默认英文单复数转换器。
var Defaults = &Inflector{}

// Inflector 按规则类型保存并执行词形转换规则。
type Inflector struct {
	rules map[RuleType]*Rule
}

// MustRegister 注册规则，失败时 panic。
func (i *Inflector) MustRegister(r *Rule) {
	if err := i.Register(r); err != nil {
		panic(err)
	}
}

// Register 初始化并注册规则。
func (i *Inflector) Register(r *Rule) error {
	if i.rules == nil {
		i.rules = make(map[RuleType]*Rule)
	}

	if err := r.Init(); err != nil {
		return err
	}

	i.rules[r.Type] = r

	return nil
}

// Inflected 按规则类型返回 s 的转换结果。
func (i *Inflector) Inflected(tye RuleType, s string) string {
	if r, ok := i.rules[tye]; ok {
		return r.Inflected(s)
	}
	return s
}
