package docker

import (
	"testing"
)

func Test_Analysis(t *testing.T) {

	table := map[string]struct {
		efficiency    float64
		sizeBytes     uint64
		userSizeBytes uint64
		wastedBytes   uint64
		wastedPercent float64
		path          string
	}{
		"docker-image": {0.9844212134184309, 1220598, 66237, 32025, 0.4834911001404049, "../../../.data/test-docker-image.tar"},
	}

	for name, test := range table {
		result := TestAnalysisFromArchive(t, test.path)

		if result.SizeBytes != test.sizeBytes {
			t.Errorf("%s.%s: expected sizeBytes=%v, got %v", t.Name(), name, test.sizeBytes, result.SizeBytes)
		}

		if result.UserSizeByes != test.userSizeBytes {
			t.Errorf("%s.%s: expected userSizeBytes=%v, got %v", t.Name(), name, test.userSizeBytes, result.UserSizeByes)
		}

		if result.WastedBytes != test.wastedBytes {
			t.Errorf("%s.%s: expected wasterBytes=%v, got %v", t.Name(), name, test.wastedBytes, result.WastedBytes)
		}

		if result.WastedUserPercent != test.wastedPercent {
			t.Errorf("%s.%s: expected wastedPercent=%v, got %v", t.Name(), name, test.wastedPercent, result.WastedUserPercent)
		}

		if result.Efficiency != test.efficiency {
			t.Errorf("%s.%s: expected efficiency=%v, got %v", t.Name(), name, test.efficiency, result.Efficiency)
		}
	}
}
