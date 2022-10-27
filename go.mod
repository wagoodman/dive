module github.com/jauderho/dive

go 1.19

require (
	github.com/awesome-gocui/gocui v1.1.0
	github.com/awesome-gocui/keybinding v1.0.1-0.20190805183143-864552bd36b7
	github.com/cespare/xxhash v1.1.0
	github.com/docker/cli v20.10.21+incompatible
	github.com/docker/docker v20.10.21+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.13.0
	github.com/google/uuid v1.3.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/lunixbochs/vtclean v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/phayes/permbits v0.0.0-20190612203442-39d7c581d2ee
	github.com/sergi/go-diff v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/afero v1.9.2
	github.com/spf13/cobra v1.6.0
	github.com/spf13/viper v1.13.0
	github.com/wagoodman/dive v0.10.0
	golang.org/x/net v0.1.0
)

require (
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/awesome-gocui/termbox-go v0.0.0-20190427202837-c0aef3d18bcc // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	golang.org/x/mod v0.6.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.2.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

// related to an invalid pseudo version in github.com/docker/distribution@v0.0.0-20181126153310-93e082742a009850ac46962150b2f652a822c5ff
//replace github.com/docker/distribution => github.com/docker/distribution v2.7.0-rc.0.0.20181024170156-93e082742a00+incompatible

// related to an invalid pseudo version in github.com/docker/distribution@v0.0.0-20181126153310-93e082742a009850ac46962150b2f652a822c5ff
//replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190822205725-ed20165a37b4

// relates to https://github.com/golangci/golangci-lint/issues/581
//replace github.com/go-critic/go-critic => github.com/go-critic/go-critic v0.3.5-0.20190526074819-1df300866540

//replace github.com/golangci/errcheck => github.com/golangci/errcheck v0.0.0-20181223084120-ef45e06d44b6

//replace github.com/golangci/go-tools => github.com/golangci/go-tools v0.0.0-20190318060251-af6baa5dc196

//replace github.com/golangci/gofmt => github.com/golangci/gofmt v0.0.0-20181222123516-0b8337e80d98

//replace github.com/golangci/gosec => github.com/golangci/gosec v0.0.0-20190211064107-66fb7fc33547

//replace github.com/golangci/ineffassign => github.com/golangci/ineffassign v0.0.0-20190609212857-42439a7714cc

//replace github.com/golangci/lint-1 => github.com/golangci/lint-1 v0.0.0-20190420132249-ee948d087217

//replace mvdan.cc/unparam => mvdan.cc/unparam v0.0.0-20190209190245-fbb59629db34

// related to https://github.com/awesome-gocui/gocui/blob/master/migrate-to-v1.md
replace github.com/awesome-gocui/gocui => github.com/awesome-gocui/gocui v0.6.0

//replace github.com/golang/protobuf => google.golang.org/protobuf v1.5.2
