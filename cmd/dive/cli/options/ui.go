package options

// UI combines all UI configuration elements
type UI struct {
	Keybinding UIKeybindings `yaml:"keybinding" mapstructure:"keybinding"`
	Diff       UIDiff        `yaml:"diff" mapstructure:"diff"`
	Filetree   UIFiletree    `yaml:"filetree" mapstructure:"filetree"`
	Layer      UILayers      `yaml:"layer" mapstructure:"layer"`
}

func DefaultUI() UI {
	return UI{
		Keybinding: DefaultUIKeybinding(),
		Diff:       DefaultUIDiff(),
		Filetree:   DefaultUIFiletree(),
		Layer:      DefaultUILayers(),
	}
}
