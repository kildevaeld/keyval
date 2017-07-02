package filesystem

import (
	"os"
	"testing"
)

func TestHasParent(t *testing.T) {

	hasParent("test.png")
	hasParent("test")
	hasParent("/test")
	hasParent("test/mig.png")
	hasParent("/test/mig/nu")

}

func TestSet(t *testing.T) {

	fs, err := (&filesystem{
		path: "test",
	}).init()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove("test")

	if err := fs.SetBytes([]byte("rapper"), []byte("Hello, World")); err != nil {
		t.Fatal(err)
	}

}
