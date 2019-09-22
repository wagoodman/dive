package dive

import (
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/dive/image/docker"
)

func GetAnalyzer(imageID string) image.Analyzer {
	// u, _ := url.Parse(imageID)
	// fmt.Printf("\n\nurl: %+v\n", u.Scheme)
	return docker.NewImageAnalyzer(imageID)
}
