func Test_LoadImage(t *testing.T) {
	testDataPath := getTestdataPath("test-docker-image.tar")
	testCmd := fmt.Sprintf("--source=docker-archive:%s", testDataPath)
	
	// Capture and normalize output for snapshot comparison
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	run(testCmd)
	
	w.Close()
	os.Stdout = oldStdout
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	
	// Normalize path for snapshot comparison
	output := buf.String()
	output = strings.ReplaceAll(output, testDataPath, "/home/calelin/dive/.data/test-docker-image.tar")
	
	fmt.Print(output)
}
