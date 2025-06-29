func extractImageLayer(data *AnalysisState, tarredFile *tar.Reader, header *tar.Header) error {
	name := header.Name
	if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
		data.content.AddLink(name, header.Linkname)
	}

	fileInfo := filetree.NewFileInfoFromTarHeader(tarredFile, header, name)
	
	// Extract extended attributes if they exist
	for key, value := range header.PAXRecords {
		// Check for extended attribute records
		if len(key) > 11 && key[:11] == "SCHILY.xattr" {
			attrKey := key[12:] // Extract attribute name
			if fileInfo.XAttrs == nil {
				fileInfo.XAttrs = make(map[string][]byte)
			}
			fileInfo.XAttrs[attrKey] = []byte(value)
		}
	}

	fileInfo.DiffType = filetree.Unmodified

	_, tree := getOrCreateTree(data.refTrees, *header.ModTime, data.layerIdx)
	tree.AddFile(fileInfo)

	return nil
}
