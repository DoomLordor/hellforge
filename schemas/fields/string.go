package fields

type StringFiled struct {
	value string
	set   bool
}

func (f *StringFiled) Parse(raw, def string) error {
	f.value = raw

	if raw == emptyValue {
		f.value = def
	}

	return nil
}

func (f *StringFiled) Set(set bool) {
	f.set = set
}

func (f *StringFiled) IsSet() bool {
	return f.set
}

func (f *StringFiled) Value() string {
	return f.value
}

func (f *StringFiled) IsMultiple() bool {
	return false
}
