package main

import "testing"

func TestAddPath(t *testing.T) {
	tree := NewTree()
	tree.AddPath("/etc/nginx/nginx.conf", 1)
	tree.AddPath("/etc/nginx/public", 2)
	tree.AddPath("/var/run/systemd", 3)
	tree.AddPath("/var/run/bashful", 4)
	tree.AddPath("/tmp", 5)
	tree.AddPath("/tmp/nonsense", 6)

	expected := `.
├── etc
│   └── nginx
│       ├── nginx.conf
│       └── public
├── tmp
│   └── nonsense
└── var
    └── run
        ├── bashful
        └── systemd
`
	actual := tree.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}

func TestRemovePath(t *testing.T) {
	tree := NewTree()
	tree.AddPath("/etc/nginx/nginx.conf", 1)
	tree.AddPath("/etc/nginx/public", 2)
	tree.AddPath("/var/run/systemd", 3)
	tree.AddPath("/var/run/bashful", 4)
	tree.AddPath("/tmp", 5)
	tree.AddPath("/tmp/nonsense", 6)

	tree.RemovePath("/var/run/bashful")
	tree.RemovePath("/tmp")

	expected := `.
├── etc
│   └── nginx
│       ├── nginx.conf
│       └── public
└── var
    └── run
        └── systemd
`
	actual := tree.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}
