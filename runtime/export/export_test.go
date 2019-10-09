package export

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/wagoodman/dive/dive/image/docker"
	"testing"
)

func Test_Export(t *testing.T) {
	result := docker.TestAnalysisFromArchive(t, "../../.data/test-docker-image.tar")

	export := NewExport(result)
	payload, err := export.Marshal()
	if err != nil {
		t.Errorf("Test_Export: unable to export analysis: %v", err)
	}

	expectedResult := `{
  "layer": [
    {
      "index": 0,
      "id": "28cfe03618aa2e914e81fdd90345245c15f4478e35252c06ca52d238fd3cc694",
      "digestId": "sha256:23bc2b70b2014dec0ac22f27bb93e9babd08cdd6f1115d0c955b9ff22b382f5a",
      "sizeBytes": 1154361,
      "command": "#(nop) ADD file:ce026b62356eec3ad1214f92be2c9dc063fe205bd5e600be3492c4dfb17148bd in / "
    },
    {
      "index": 1,
      "id": "1871059774abe6914075e4a919b778fa1561f577d620ae52438a9635e6241936",
      "digestId": "sha256:a65b7d7ac139a0e4337bc3c73ce511f937d6140ef61a0108f7d4b8aab8d67274",
      "sizeBytes": 6405,
      "command": "#(nop) ADD file:139c3708fb6261126453e34483abd8bf7b26ed16d952fd976994d68e72d93be2 in /somefile.txt "
    },
    {
      "index": 2,
      "id": "49fe2a475548bfa4d493fc796fce41f30704e3d4cbff3e45dd3e06f463236d1d",
      "digestId": "sha256:93e208d471756ffbac88cf9c25feb442007f221d3bd73231e27b747a0a68927c",
      "sizeBytes": 0,
      "command": "mkdir -p /root/example/really/nested"
    },
    {
      "index": 3,
      "id": "80cd2ca1ffc89962b9349c80280c2bc551acbd11e09b16badb0669f8e2369020",
      "digestId": "sha256:4abad3abe3cb99ad7a492a9d9f6b3d66287c1646843c74128bbbec4f7be5aa9e",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile1.txt"
    },
    {
      "index": 4,
      "id": "c99e2f8d3f6282668f0d30dc1db5e67a51d7a1dcd7ff6ddfa0f90760836778ec",
      "digestId": "sha256:14c9a6ffcb6a0f32d1035f97373b19608e2d307961d8be156321c3f1c1504cbf",
      "sizeBytes": 6405,
      "command": "chmod 444 /root/example/somefile1.txt"
    },
    {
      "index": 5,
      "id": "5eca617bdc3bc06134fe957a30da4c57adb7c340a6d749c8edc4c15861c928d7",
      "digestId": "sha256:778fb5770ef466f314e79cc9dc418eba76bfc0a64491ce7b167b76aa52c736c4",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile2.txt"
    },
    {
      "index": 6,
      "id": "f07c3eb887572395408f8e11a07af945e4da5f02b3188bb06b93fad713ca0b99",
      "digestId": "sha256:f275b8a31a71deb521cc048e6021e2ff6fa52bedb25c9b7bbe129a0195ddca5f",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile3.txt"
    },
    {
      "index": 7,
      "id": "461885fc22589158dee3c5b9f01cc41c87805439f58b4399d733b51aa305cbf9",
      "digestId": "sha256:dd1effc5eb19894c3e9b57411c98dd1cf30fa1de4253c7fae53c9cea67267d83",
      "sizeBytes": 6405,
      "command": "mv /root/example/somefile3.txt /root/saved.txt"
    },
    {
      "index": 8,
      "id": "a10327f68ffed4afcba78919052809a8f774978a6b87fc117d39c53c4842f72c",
      "digestId": "sha256:8d1869a0a066cdd12e48d648222866e77b5e2814f773bb3bd8774ab4052f0f1d",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /root/.saved.txt"
    },
    {
      "index": 9,
      "id": "f2fc54e25cb7966dc9732ec671a77a1c5c104e732bd15ad44a2dc1ac42368f84",
      "digestId": "sha256:bc2e36423fa31a97223fd421f22c35466220fa160769abf697b8eb58c896b468",
      "sizeBytes": 0,
      "command": "rm -rf /root/example/"
    },
    {
      "index": 10,
      "id": "aad36d0b05e71c7e6d4dfe0ca9ed6be89e2e0d8995dafe83438299a314e91071",
      "digestId": "sha256:7f648d45ee7b6de2292162fba498b66cbaaf181da9004fcceef824c72dbae445",
      "sizeBytes": 2187,
      "command": "#(nop) ADD dir:7ec14b81316baa1a31c38c97686a8f030c98cba2035c968412749e33e0c4427e in /root/.data/ "
    },
    {
      "index": 11,
      "id": "3d4ad907517a021d86a4102d2764ad2161e4818bbd144e41d019bfc955434181",
      "digestId": "sha256:a4b8f95f266d5c063c9a9473c45f2f85ddc183e37941b5e6b6b9d3c00e8e0457",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /tmp/saved.again1.txt"
    },
    {
      "index": 12,
      "id": "81b1b002d4b4c1325a9cad9990b5277e7f29f79e0f24582344c0891178f95905",
      "digestId": "sha256:22a44d45780a541e593a8862d80f3e14cb80b6bf76aa42ce68dc207a35bf3a4a",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /root/.data/saved.again2.txt"
    },
    {
      "index": 13,
      "id": "cfb35bb5c127d848739be5ca726057e6e2c77b2849f588e7aebb642c0d3d4b7b",
      "digestId": "sha256:ba689cac6a98c92d121fa5c9716a1bab526b8bb1fd6d43625c575b79e97300c5",
      "sizeBytes": 6405,
      "command": "chmod +x /root/saved.txt"
    }
  ],
  "image": {
    "sizeBytes": 1220598,
    "inefficientBytes": 32025,
    "efficiencyScore": 0.9844212134184309,
    "fileReference": [
      {
        "count": 2,
        "sizeBytes": 12810,
        "file": "/root/saved.txt"
      },
      {
        "count": 2,
        "sizeBytes": 12810,
        "file": "/root/example/somefile1.txt"
      },
      {
        "count": 2,
        "sizeBytes": 6405,
        "file": "/root/example/somefile3.txt"
      }
    ]
  }
}`
	actualResult := string(payload)
	if expectedResult != actualResult {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expectedResult, actualResult, false)

		t.Errorf("Test_Export: unexpected export result:\n%v", dmp.DiffPrettyText(diffs))
	}
}
