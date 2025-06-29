func TestCompare(t *testing.T) {
	// Source values
	sampleData := []struct {
		typeflag    byte
		isDir       bool
		size        int64
		mode        os.FileMode
		hash        string
		compareWith int
		expected    DiffType
	}{
		{
			typeflag:    tar.TypeReg,
			isDir:       false,
			size:        1,
			mode:        0600,
			hash:        "deadbeef",
			compareWith: 1,
			expected:    Unmodified,
		},
		{
			typeflag:    tar.TypeReg,
			isDir:       false,
			size:        2,
			mode:        0600,
			hash:        "deadbeef",
			compareWith: 0,
			expected:    Modified,
		},
		{
			typeflag:    tar.TypeReg,
			isDir:       false,
			size:        1,
			mode:        0601,
			hash:        "deadbeef",
			compareWith: 0,
			expected:    Modified,
		},
		{
			typeflag:    tar.TypeReg,
			isDir:       false,
			size:        1,
			mode:        0600,
			hash:        "deadbeee",
			compareWith: 0,
			expected:    Modified,
		},
		{
			typeflag:    tar.TypeReg,
			isDir:       false,
			size:        1,
			mode:        0600,
			hash:        "deadbeef",
			compareWith: -1,
			expected:    Modified,
		},
		{
			typeflag:    tar.TypeReg,
			isDir:       true,
			size:        1,
			mode:        0600,
			hash:        "deadbeef",
			compareWith: 0,
			expected:    Modified,
		},
		{
			typeflag:    tar.TypeReg,
			isDir:       true,
			size:        1,
			mode:        0600,
			hash:        "deadbeef",
			compareWith: 6,
			expected:    Unmodified,
		},
	}

	// Generate the file nodes!
	var files []*FileNode
	for _, tc := range sampleData {
		data := FileInfo{
			TypeFlag: tc.typeflag,
			IsDir:    tc.isDir,
			Size:     tc.size,
			Mode:     tc.mode,
			MD5sum:   tc.hash,
			XAttrs:   make(map[string][]byte),
		}
		files = append(files, NewFileNode(nil, nil, "blerg", data))
	}

	// Match!
	for idx, fc := range sampleData {
		var expected DiffType
		var actual DiffType
		expected = fc.expected
		if fc.compareWith < 0 {
			actual = files[idx].Compare(nil)
		} else {
			actual = files[idx].Compare(files[fc.compareWith])
		}

		if actual != expected {
			t.Errorf("actual: %v but expected: %v for TC #%v", actual, expected, idx)
		}
	}
}

func TestCompareXAttrs(t *testing.T) {
	// Test cases for extended attributes comparison
	testCases := []struct {
		name     string
		xattrs1  map[string][]byte
		xattrs2  map[string][]byte
		expected bool
	}{
		{
			name:     "Both nil",
			xattrs1:  nil,
			xattrs2:  nil,
			expected: true,
		},
		{
			name:     "Empty maps",
			xattrs1:  make(map[string][]byte),
			xattrs2:  make(map[string][]byte),
			expected: true,
		},
		{
			name: "Same single attribute",
			xattrs1: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
			},
			xattrs2: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
			},
			expected: true,
		},
		{
			name: "Different attribute values",
			xattrs1: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
			},
			xattrs2: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x04},
			},
			expected: false,
		},
		{
			name: "Different attribute keys",
			xattrs1: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
			},
			xattrs2: map[string][]byte{
				"user.attribute": []byte{0x01, 0x02, 0x03},
			},
			expected: false,
		},
		{
			name: "Different number of attributes",
			xattrs1: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
				"user.attribute":      []byte{0x04, 0x05, 0x06},
			},
			xattrs2: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
			},
			expected: false,
		},
		{
			name: "Multiple same attributes",
			xattrs1: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
				"user.attribute":      []byte{0x04, 0x05, 0x06},
			},
			xattrs2: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
				"user.attribute":      []byte{0x04, 0x05, 0x06},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := compareXAttrs(tc.xattrs1, tc.xattrs2)
			if result != tc.expected {
				t.Errorf("compareXAttrs(%v, %v) = %v, want %v", tc.xattrs1, tc.xattrs2, result, tc.expected)
			}
		})
	}

	// Test that file nodes with different xattrs are marked as modified
	t.Run("FileNodes with different xattrs", func(t *testing.T) {
		// Create two identical file nodes except for xattrs
		data1 := FileInfo{
			TypeFlag: tar.TypeReg,
			IsDir:    false,
			Size:     1,
			Mode:     0600,
			MD5sum:   "deadbeef",
			XAttrs: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x03},
			},
		}
		data2 := FileInfo{
			TypeFlag: tar.TypeReg,
			IsDir:    false,
			Size:     1,
			Mode:     0600,
			MD5sum:   "deadbeef",
			XAttrs: map[string][]byte{
				"security.capability": []byte{0x01, 0x02, 0x04}, // Different value
			},
		}
		node1 := NewFileNode(nil, nil, "file", data1)
		node2 := NewFileNode(nil, nil, "file", data2)

		// They should be marked as modified due to different xattrs
		if node1.Compare(node2) != Modified {
			t.Errorf("Expected files with different xattrs to be marked as Modified")
		}
	})
}
