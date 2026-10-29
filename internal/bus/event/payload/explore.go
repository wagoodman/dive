package payload

import "github.com/wagoodman/dive/dive/image"

type Explore struct {
	Analysis image.Analysis
	Content  image.ContentReader
}
