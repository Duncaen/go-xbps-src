package runtime

import (
	"testing"
)

func TestNew(t *testing.T) {
	_, err := New("/home/duncan/void-packages")
	if err != nil {
		t.Fatal(err)
	}
}
