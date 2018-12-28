package runtime

import (
	"github.com/wagoodman/dive/image"
	"testing"
)

func Test_Export(t *testing.T) {

	result, err := image.TestLoadDockerImageTar("../.data/test-docker-image.tar")
	if err != nil {
		t.Fatalf("Test_Export: unable to fetch analysis: %v", err)
	}
	export := newExport(result)
	payload, err := export.marshal()
	if err != nil {
		t.Errorf("Test_Export: unable to export analysis: %v", err)
	}

	expectedResult := `{
  "layer": [
    {
      "index": 0,
      "digestId": "sha256:23bc2b70b2014dec0ac22f27bb93e9babd08cdd6f1115d0c955b9ff22b382f5a",
      "sizeBytes": 1154361,
      "command": "#(nop) ADD file:ce026b62356eec3ad1214f92be2c9dc063fe205bd5e600be3492c4dfb17148bd in / "
    },
    {
      "index": 1,
      "digestId": "sha256:a65b7d7ac139a0e4337bc3c73ce511f937d6140ef61a0108f7d4b8aab8d67274",
      "sizeBytes": 6405,
      "command": "#(nop) ADD file:139c3708fb6261126453e34483abd8bf7b26ed16d952fd976994d68e72d93be2 in /somefile.txt "
    },
    {
      "index": 2,
      "digestId": "sha256:93e208d471756ffbac88cf9c25feb442007f221d3bd73231e27b747a0a68927c",
      "sizeBytes": 0,
      "command": "mkdir -p /root/example/really/nested"
    },
    {
      "index": 3,
      "digestId": "sha256:4abad3abe3cb99ad7a492a9d9f6b3d66287c1646843c74128bbbec4f7be5aa9e",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile1.txt"
    },
    {
      "index": 4,
      "digestId": "sha256:14c9a6ffcb6a0f32d1035f97373b19608e2d307961d8be156321c3f1c1504cbf",
      "sizeBytes": 6405,
      "command": "chmod 444 /root/example/somefile1.txt"
    },
    {
      "index": 5,
      "digestId": "sha256:778fb5770ef466f314e79cc9dc418eba76bfc0a64491ce7b167b76aa52c736c4",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile2.txt"
    },
    {
      "index": 6,
      "digestId": "sha256:f275b8a31a71deb521cc048e6021e2ff6fa52bedb25c9b7bbe129a0195ddca5f",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile3.txt"
    },
    {
      "index": 7,
      "digestId": "sha256:dd1effc5eb19894c3e9b57411c98dd1cf30fa1de4253c7fae53c9cea67267d83",
      "sizeBytes": 6405,
      "command": "mv /root/example/somefile3.txt /root/saved.txt"
    },
    {
      "index": 8,
      "digestId": "sha256:8d1869a0a066cdd12e48d648222866e77b5e2814f773bb3bd8774ab4052f0f1d",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /root/.saved.txt"
    },
    {
      "index": 9,
      "digestId": "sha256:bc2e36423fa31a97223fd421f22c35466220fa160769abf697b8eb58c896b468",
      "sizeBytes": 0,
      "command": "rm -rf /root/example/"
    },
    {
      "index": 10,
      "digestId": "sha256:7f648d45ee7b6de2292162fba498b66cbaaf181da9004fcceef824c72dbae445",
      "sizeBytes": 2187,
      "command": "#(nop) ADD dir:7ec14b81316baa1a31c38c97686a8f030c98cba2035c968412749e33e0c4427e in /root/.data/ "
    },
    {
      "index": 11,
      "digestId": "sha256:a4b8f95f266d5c063c9a9473c45f2f85ddc183e37941b5e6b6b9d3c00e8e0457",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /tmp/saved.again1.txt"
    },
    {
      "index": 12,
      "digestId": "sha256:22a44d45780a541e593a8862d80f3e14cb80b6bf76aa42ce68dc207a35bf3a4a",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /root/.data/saved.again2.txt"
    },
    {
      "index": 13,
      "digestId": "sha256:ba689cac6a98c92d121fa5c9716a1bab526b8bb1fd6d43625c575b79e97300c5",
      "sizeBytes": 6405,
      "command": "chmod +x /root/saved.txt"
    }
  ],
  "image": {
    "sizeBytes": 1220598,
    "inefficientBytes": 32025,
    "efficiencyScore": 0.9844212134184309,
    "inefficientFiles": [
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
		t.Errorf("Test_Export: unexpected export result:\n%v", actualResult)
	}
}
