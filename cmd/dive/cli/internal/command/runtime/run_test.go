package runtime

//type defaultResolver struct {
//	testing.TB
//}
//
//func (r *defaultResolver) Name() string {
//	return "default-resolver"
//}
//
//func (r *defaultResolver) Extract(id string, l string, p string) error {
//	return nil
//}
//
//func (r *defaultResolver) Fetch(id string) (*image.Image, error) {
//	archive, err := docker.TestLoadArchive(r, "../.data/test-docker-image.tar")
//	if err != nil {
//		return nil, err
//	}
//	return archive.ToImage(id)
//}
//
//func (r *defaultResolver) Build(args []string) (*image.Image, error) {
//	return r.Fetch("")
//}
//
//type failedBuildResolver struct {
//	testing.TB
//}
//
//func (r *failedBuildResolver) Name() string {
//	return "failed-build-resolver"
//}
//
//func (r *failedBuildResolver) Extract(id string, l string, p string) error {
//	return fmt.Errorf("some extract failure")
//}
//
//func (r *failedBuildResolver) Fetch(id string) (*image.Image, error) {
//	archive, err := docker.TestLoadArchive(r, "../.data/test-docker-image.tar")
//	if err != nil {
//		return nil, err
//	}
//	return archive.ToImage(id)
//}
//
//func (r *failedBuildResolver) Build(args []string) (*image.Image, error) {
//	return nil, fmt.Errorf("some build failure")
//}
//
//type failedFetchResolver struct {
//	testing.TB
//}
//
//func (r *failedFetchResolver) Name() string {
//	return "failed-fetch-resolver"
//}
//
//func (r *failedFetchResolver) Extract(id string, l string, p string) error {
//	return fmt.Errorf("some extract failure")
//}
//
//func (r *failedFetchResolver) Fetch(id string) (*image.Image, error) {
//	return nil, fmt.Errorf("some fetch failure")
//}
//
//func (r *failedFetchResolver) Build(args []string) (*image.Image, error) {
//	return nil, fmt.Errorf("some build failure")
//}
//
//type testEvent struct {
//	Stdout      string
//	Stderr      string
//	ErrMessage  string
//	ErrorOnExit bool
//}
//
//func newTestEvent(e event) testEvent {
//	var errMsg string
//	if e.err != nil {
//		errMsg = e.err.Error()
//	}
//	return testEvent{
//		Stdout:      vtclean.Clean(e.stdout, false),
//		Stderr:      vtclean.Clean(e.stderr, false),
//		ErrMessage:  vtclean.Clean(errMsg, false),
//		ErrorOnExit: e.errorOnExit,
//	}
//}
//
//func configureCi(t testing.TB) []ci.Rule {
//	rules, err := ci.Rules("0.9", "1000", "0.1")
//	require.NoError(t, err)
//	return rules
//}
//
//func TestRun(t *testing.T) {
//	testCases := []struct {
//		name     string
//		resolver image.Resolver
//		options  Config
//		events   []testEvent
//	}{
//		{
//			name:     "fetch-case",
//			resolver: &defaultResolver{t},
//			options: Config{
//				Ci:         false,
//				Image:      "dive-example",
//				Source:     dive.SourceDockerEngine,
//				ExportFile: "",
//				CiRules:    nil,
//				BuildArgs:  nil,
//			},
//			events: []testEvent{
//				{Stdout: "Image Source: docker://dive-example"},
//				{Stdout: "Extracting image from default-resolver... (this can take a while for large images)"},
//				{Stdout: "Analyzing image..."},
//			},
//		},
//		{
//			name:     "fetch-with-no-build-options-case",
//			resolver: &defaultResolver{t},
//			options: Config{
//				Ci:         false,
//				Image:      "dive-example",
//				Source:     dive.SourceDockerEngine,
//				ExportFile: "",
//				CiRules:    nil,
//				// note: empty slice is passed
//				BuildArgs: []string{},
//			},
//			events: []testEvent{
//				{Stdout: "Image Source: docker://dive-example"},
//				{Stdout: "Extracting image from default-resolver... (this can take a while for large images)"},
//				{Stdout: "Analyzing image..."},
//			},
//		},
//		{
//			name:     "build-case",
//			resolver: &defaultResolver{t},
//			options: Config{
//				Ci:         false,
//				Image:      "dive-example",
//				Source:     dive.SourceDockerEngine,
//				ExportFile: "",
//				CiRules:    nil,
//				BuildArgs:  []string{"an-option"},
//			},
//			events: []testEvent{
//				{Stdout: "Building image..."},
//				{Stdout: "Analyzing image..."},
//			},
//		},
//		{
//			name:     "failed-fetch",
//			resolver: &failedFetchResolver{t},
//			options: Config{
//				Ci:         false,
//				Image:      "dive-example",
//				Source:     dive.SourceDockerEngine,
//				ExportFile: "",
//				CiRules:    nil,
//				BuildArgs:  nil,
//			},
//			events: []testEvent{
//				{Stdout: "Image Source: docker://dive-example"},
//				{Stdout: "Extracting image from failed-fetch-resolver... (this can take a while for large images)"},
//				{Stdout: "", Stderr: "cannot fetch image", ErrorOnExit: true, ErrMessage: "some fetch failure"},
//			},
//		},
//		{
//			name:     "failed-build",
//			resolver: &failedBuildResolver{t},
//			options: Config{
//				Ci:         false,
//				Image:      "doesn't-matter",
//				Source:     dive.SourceDockerEngine,
//				ExportFile: "",
//				CiRules:    nil,
//				BuildArgs:  []string{"an-option"},
//			},
//			events: []testEvent{
//				{Stdout: "Building image..."},
//				{Stdout: "", Stderr: "cannot build image", ErrorOnExit: true, ErrMessage: "some build failure"},
//			},
//		},
//		{
//			name:     "ci-go-case",
//			resolver: &defaultResolver{t},
//			options: Config{
//				Ci:         true,
//				Image:      "doesn't-matter",
//				Source:     dive.SourceDockerEngine,
//				ExportFile: "",
//				CiRules:    configureCi(t),
//				BuildArgs:  []string{"an-option"},
//			},
//			events: []testEvent{
//				{Stdout: "Building image..."},
//				{Stdout: "Analyzing image..."},
//				{Stdout: "  efficiency: 98.4421 %"},
//				{Stdout: "  wastedBytes: 32025 bytes (32 kB)"},
//				{Stdout: "  userWastedPercent: 48.3491 %"},
//				{Stdout: "Inefficient Files:\nCount  Wasted Space  File Path\n    2         13 kB  /root/saved.txt\n    2         13 kB  /root/example/somefile1.txt\n    2        6.4 kB  /root/example/somefile3.txt\nResults:\n  FAIL: highestUserWastedPercent: too many bytes wasted, relative to the user bytes added (%-user-wasted-bytes=0.4834911001404049 > threshold=0.1)\n  FAIL: highestWastedBytes: too many bytes wasted (wasted-bytes=32025 > threshold=1000)\n  PASS: lowestEfficiency\nResult:FAIL [Total:3] [Passed:1] [Failed:2] [Warn:0] [Skipped:0]\n"},
//				{Stdout: "", Stderr: "", ErrorOnExit: true, ErrMessage: ""},
//			},
//		},
//		{
//			name:     "no-ci-config-uses-default-values",
//			resolver: &defaultResolver{t},
//			options: Config{
//				Ci:         true,
//				Image:      "doesn't-matter",
//				Source:     dive.SourceDockerEngine,
//				ExportFile: "",
//				CiRules:    nil,
//				BuildArgs:  []string{"an-option"},
//			},
//			events: []testEvent{
//
//				{Stdout: "Building image..."},
//				{Stdout: "Analyzing image..."},
//				{Stdout: "  efficiency: 98.4421 %"},
//				{Stdout: "  wastedBytes: 32025 bytes (32 kB)"},
//				{Stdout: "  userWastedPercent: 48.3491 %"},
//				{Stdout: "Inefficient Files:\nCount  Wasted Space  File Path\n    2         13 kB  /root/saved.txt\n    2         13 kB  /root/example/somefile1.txt\n    2        6.4 kB  /root/example/somefile3.txt\nResults:\nResult:PASS [Total:0] [Passed:0] [Failed:0] [Warn:0] [Skipped:0]\n"},
//			},
//		},
//		{
//			name:     "export-go-case",
//			resolver: &defaultResolver{t},
//			options: Config{
//				Ci:         true,
//				Image:      "doesn't-matter",
//				Source:     dive.SourceDockerEngine,
//				ExportFile: "some-file.json",
//				CiRules:    configureCi(t),
//				BuildArgs:  []string{"an-option"},
//			},
//			events: []testEvent{
//				{Stdout: "Building image..."},
//				{Stdout: "Analyzing image..."},
//				{Stdout: "Exporting image to 'some-file.json'..."},
//			},
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.name, func(t *testing.T) {
//			var ec = make(eventChannel)
//			var events = make([]testEvent, 0)
//			var filesystem = afero.NewMemMapFs()
//
//			go run(context.TODO(), false, testCase.options, testCase.resolver, ec, filesystem)
//
//			for e := range ec {
//				te := newTestEvent(e)
//				events = append(events, te)
//			}
//
//			// showEvents(events)
//
//			if d := cmp.Diff(testCase.events, events); d != "" {
//				t.Errorf("events mismatch (-want +got):\n%s", d)
//			}
//
//		})
//	}
//}

// func showEvents(events []testEvent) {
// 	for _, e := range events {
// 		fmt.Printf("{stdout:\"%s\", stderr:\"%s\", errorOnExit: %v, errMessage: \"%s\"},\n",
// 			strings.Replace(vtclean.Clean(e.stdout, false), "\n", "\\n", -1),
// 			strings.Replace(vtclean.Clean(e.stderr, false), "\n", "\\n", -1),
// 			e.errorOnExit,
// 			e.errMessage)
// 	}
// }
