package pathstyle

import (
	"errors"
	"testing"
)

func TestTransformWSLConvertsDrivePath(t *testing.T) {
	got, err := Transform(`C:\Users\Alice\Pictures\a.png`, WSL)
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}
	want := "/mnt/c/Users/Alice/Pictures/a.png"
	if got != want {
		t.Fatalf("Transform() = %q, want %q", got, want)
	}
}

func TestTransformWSLConvertsForwardSlashDrivePath(t *testing.T) {
	got, err := Transform("D:/work/demo.txt", WSL)
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}
	want := "/mnt/d/work/demo.txt"
	if got != want {
		t.Fatalf("Transform() = %q, want %q", got, want)
	}
}

func TestTransformNativeLeavesPathUnchanged(t *testing.T) {
	path := `C:\Users\Alice\Pictures\a.png`
	got, err := Transform(path, Native)
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}
	if got != path {
		t.Fatalf("Transform() = %q, want %q", got, path)
	}
}

func TestTransformWindowsLeavesPathUnchanged(t *testing.T) {
	path := `C:\Users\Alice\Pictures\a.png`
	got, err := Transform(path, Windows)
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}
	if got != path {
		t.Fatalf("Transform() = %q, want %q", got, path)
	}
}

func TestTransformWSLLeavesUnknownPathUnchanged(t *testing.T) {
	path := `\\server\share\a.png`
	got, err := Transform(path, WSL)
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}
	if got != path {
		t.Fatalf("Transform() = %q, want %q", got, path)
	}
}

func TestTransformRejectsUnknownStyle(t *testing.T) {
	_, err := Transform("C:\\a.txt", "posix")
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("error = %v, want ErrUnsupported", err)
	}
}
