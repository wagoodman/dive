package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

func main() {
	f, err := os.Open("image/cache.tar")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	// gzf, err := gzip.NewReader(f)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	tarReader := tar.NewReader(f)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			fmt.Println("File: ", name)
			// show the contents
			// io.Copy(os.Stdout, tarReader)
		default:
			fmt.Printf("%s : %c %s %s\n",
				"hmmm?",
				header.Typeflag,
				"in file",
				name,
			)
		}
	}
}
