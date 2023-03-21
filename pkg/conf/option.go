package conf

// GetOption ...
type (
	GetOption  func(o *GetOptions)
	GetOptions struct {
		TagName string
	}
)

var defaultGetOptions = GetOptions{
	TagName: "mapstructure",
}

func TagName(tag string) GetOption {
	return func(o *GetOptions) {
		o.TagName = tag
	}
}

func TagNameJSON() GetOption {
	return TagName("json")
}

func TagNameYAML() GetOption {
	return TagName("yaml")
}
