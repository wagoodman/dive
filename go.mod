module github.com/wagoodman/dive

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/buildpacks/imgutil v0.0.0-20200805143852-1844b230530d // indirect
	github.com/buildpacks/lifecycle v0.7.2
	github.com/cespare/xxhash v1.1.0
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/docker/cli v20.10.0-beta1+incompatible
	github.com/docker/docker v1.4.2-0.20200221181110-62bd5a33f707
	github.com/dustin/go-humanize v1.0.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gdamore/tcell/v2 v2.1.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/mock v1.4.1 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/google/go-containerregistry v0.0.0-20200313165449-955bf358a3d8 // indirect
	github.com/google/uuid v1.1.2
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/lunixbochs/vtclean v1.0.0
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/phayes/permbits v0.0.0-20190612203442-39d7c581d2ee
	github.com/rivo/tview v0.0.0-20201204190810-5406288b8e4e
	github.com/sergi/go-diff v1.1.0
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.4.1
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	gitlab.com/tslocum/cbind v0.1.4
	golang.org/x/net v0.0.0-20201029055024-942e2f445f3c
	golang.org/x/sys v0.0.0-20201218084310-7d0127a74742 // indirect
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20201029200359-8ce4113da6f7 // indirect
	google.golang.org/grpc v1.33.1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

// related to an invalid pseudo version in github.com/docker/distribution@v0.0.0-20181126153310-93e082742a009850ac46962150b2f652a822c5ff
//replace github.com/docker/distribution => github.com/docker/distribution v2.7.0-rc.0.0.20181024170156-93e082742a00+incompatible

// related to an invalid pseudo version in github.com/docker/distribution@v0.0.0-20181126153310-93e082742a009850ac46962150b2f652a822c5ff
//replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190822205725-ed20165a37b4

// relates to https://github.com/golangci/golangci-lint/issues/581
replace github.com/go-critic/go-critic => github.com/go-critic/go-critic v0.3.5-0.20190526074819-1df300866540

replace github.com/golangci/errcheck => github.com/golangci/errcheck v0.0.0-20181223084120-ef45e06d44b6

replace github.com/golangci/go-tools => github.com/golangci/go-tools v0.0.0-20190318060251-af6baa5dc196

replace github.com/golangci/gofmt => github.com/golangci/gofmt v0.0.0-20181222123516-0b8337e80d98

replace github.com/golangci/gosec => github.com/golangci/gosec v0.0.0-20190211064107-66fb7fc33547

replace github.com/golangci/ineffassign => github.com/golangci/ineffassign v0.0.0-20190609212857-42439a7714cc

replace github.com/golangci/lint-1 => github.com/golangci/lint-1 v0.0.0-20190420132249-ee948d087217

replace mvdan.cc/unparam => mvdan.cc/unparam v0.0.0-20190209190245-fbb59629db34

replace github.com/rivo/tview => github.com/dwillist/tview v0.0.0-20210114215028-5f3be7b5a4ec
