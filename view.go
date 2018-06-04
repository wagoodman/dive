package main

type View interface {
	keybindings() error
	cursorDown() error
	cursorUp() error
	render() error
}
