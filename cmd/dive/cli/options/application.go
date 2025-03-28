package options

type Application struct {
	Analysis Analysis `yaml:",inline" mapstructure:",squash"`
	CI       CI       `yaml:",inline" mapstructure:",squash"`
	Export   Export   `yaml:",inline" mapstructure:",squash"`
	UI       UI       `yaml:",inline" mapstructure:",squash"`
}

func DefaultApplication() Application {
	return Application{
		Analysis: DefaultAnalysis(),
		CI:       DefaultCI(),
		Export:   DefaultExport(),
		UI:       DefaultUI(),
	}
}
