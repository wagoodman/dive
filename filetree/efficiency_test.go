package filetree

// TODO: rewrite this to be weighted by file size

// func TestEfficencyMap(t *testing.T) {
// 	trees := make([]*FileTree, 3)
// 	for ix, _ := range trees {
// 		tree := NewFileTree()
// 		tree.AddPath("/etc/nginx/nginx.conf", FileInfo{})
// 		tree.AddPath("/etc/nginx/public", FileInfo{})
// 		trees[ix] = tree
// 	}
// 	var expectedMap = map[string]int{
// 		"/etc/nginx/nginx.conf": 3,
// 		"/etc/nginx/public":     3,
// 	}
// 	actualMap := EfficiencyMap(trees)
// 	if !reflect.DeepEqual(expectedMap, actualMap) {
// 		t.Fatalf("Expected %v but go %v", expectedMap, actualMap)
// 	}
// }
//
// func TestEfficiencyScore(t *testing.T) {
// 	trees := make([]*FileTree, 3)
// 	for ix, _ := range trees {
// 		tree := NewFileTree()
// 		tree.AddPath("/etc/nginx/nginx.conf", FileInfo{})
// 		tree.AddPath("/etc/nginx/public", FileInfo{})
// 		trees[ix] = tree
// 	}
// 	expected := 2.0 / 6.0
// 	actual := CalculateEfficiency(trees)
// 	if math.Abs(expected-actual) > 0.0001 {
// 		t.Fatalf("Expected %f but got %f", expected, actual)
// 	}
//
// 	trees = make([]*FileTree, 1)
// 	for ix, _ := range trees {
// 		tree := NewFileTree()
// 		tree.AddPath("/etc/nginx/nginx.conf", FileInfo{})
// 		tree.AddPath("/etc/nginx/public", FileInfo{})
// 		trees[ix] = tree
// 	}
// 	expected = 1.0
// 	actual = CalculateEfficiency(trees)
// 	if math.Abs(expected-actual) > 0.0001 {
// 		t.Fatalf("Expected %f but got %f", expected, actual)
// 	}
// }

