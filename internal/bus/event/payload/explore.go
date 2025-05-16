package payload

import (
	"github.com/wagoodman/dive/dive/v1/image"
)

type Explore struct {
	Analysis image.Analysis
	Content  image.ContentReader
}
