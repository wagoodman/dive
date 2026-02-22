package view

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/lunixbochs/vtclean"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/dive/dive/filetree"
)

func imageDetailsContentSourcePath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to determine test file path")
	return filepath.Join(filepath.Dir(thisFile), "image_details_content.go")
}

func imageDetailsContentAST(t *testing.T) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, imageDetailsContentSourcePath(t), nil, parser.ParseComments)
	require.NoError(t, err, "unable to parse image_details_content.go")
	return file
}

func funcDeclByName(file *ast.File, funcName string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil || fn.Name.Name != funcName || fn.Recv != nil {
			continue
		}
		return fn
	}
	return nil
}

// Regression guard: report construction must not use repeated string appends
// (`+=`) in loops, which can reintroduce O(n^2) behavior.
func TestRenderImageDetailsBodyAvoidsInLoopStringAddAssign(t *testing.T) {
	file := imageDetailsContentAST(t)
	fn := funcDeclByName(file, "renderImageDetailsBody")
	require.NotNil(t, fn, "missing renderImageDetailsBody declaration")

	found := false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok || assign.Tok != token.ADD_ASSIGN || len(assign.Lhs) != 1 {
			return true
		}

		lhs, ok := assign.Lhs[0].(*ast.Ident)
		if ok && lhs.Name == "inefficiencyReport" {
			found = true
			return false
		}
		return true
	})

	require.False(t, found, "renderImageDetailsBody reintroduced inefficiencyReport += ... in-loop concatenation")
}

func TestRenderImageDetailsBodyIncludesExpectedContent(t *testing.T) {
	body, err := renderImageDetailsBody(
		"ghcr.io/openclaw/openclaw:test",
		10*1024*1024,
		0.72,
		filetree.EfficiencySlice{
			{Path: "/a", Nodes: []*filetree.FileNode{{}, {}, {}}, CumulativeSize: 40},
			{Path: "/b", Nodes: []*filetree.FileNode{{}}, CumulativeSize: 60},
		},
	)
	require.NoError(t, err)

	clean := vtclean.Clean(body, false)
	require.Contains(t, clean, "Image name:")
	require.Contains(t, clean, "ghcr.io/openclaw/openclaw:test")
	require.Contains(t, clean, "Total Image size:")
	require.Contains(t, clean, "Potential wasted space:")
	require.Contains(t, clean, "Image efficiency score:")
	require.Contains(t, clean, "/a")
	require.Contains(t, clean, "/b")
}

// Integration-style perf check: rendering should scale near-linearly as
// inefficiency row counts grow.
func TestRenderImageDetailsBodyScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping perf-sensitive scaling check in short mode")
	}

	buildDuration := func(entries, rounds int) time.Duration {
		payload := make(filetree.EfficiencySlice, 0, entries)
		for i := 0; i < entries; i++ {
			payload = append(payload, &filetree.EfficiencyData{
				Path:           "/path/" + strconv.Itoa(i),
				Nodes:          make([]*filetree.FileNode, (i%3)+1),
				CumulativeSize: int64(64 + i%1024),
			})
		}

		start := time.Now()
		for i := 0; i < rounds; i++ {
			_, err := renderImageDetailsBody("img:test", 1234, 0.9, payload)
			require.NoError(t, err)
		}
		return time.Since(start)
	}

	small := buildDuration(1200, 10)
	large := buildDuration(2400, 10)
	ratio := float64(large) / float64(small)

	require.Less(t, ratio, 3.2, "expected near-linear scaling, got ratio %.2f (small=%s large=%s)", ratio, small, large)
}
