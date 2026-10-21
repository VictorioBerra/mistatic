package api

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestZipDirAndUnzip(t *testing.T) {
	// Create a temporary source directory with some files and subdirectories
	srcDir, err := os.MkdirTemp("", "mistatic_test_src_*")
	if err != nil {
		t.Fatalf("failed to create src temp dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	testFiles := map[string]string{
		"index.html":          "<h1>Hello World</h1>",
		"css/style.css":       "body { background: #000; }",
		"js/deep/app.js":      "console.log('test');",
	}

	for relPath, content := range testFiles {
		fullPath := filepath.Join(srcDir, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", relPath, err)
		}
	}

	// 1. Zip the directory
	zipPath := filepath.Join(os.TempDir(), "mistatic_test_artifact.zip")
	defer os.Remove(zipPath)

	if err := zipDir(srcDir, zipPath); err != nil {
		t.Fatalf("zipDir failed: %v", err)
	}

	// Verify zip exists and has content
	info, err := os.Stat(zipPath)
	if err != nil || info.Size() == 0 {
		t.Fatalf("expected non-empty zip file, got err=%v, size=%d", err, info.Size())
	}

	// 2. Unzip into a destination directory
	destDir, err := os.MkdirTemp("", "mistatic_test_dest_*")
	if err != nil {
		t.Fatalf("failed to create dest temp dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	if err := unzip(zipPath, destDir); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	// 3. Verify all extracted files match original content
	for relPath, expectedContent := range testFiles {
		extractedPath := filepath.Join(destDir, filepath.FromSlash(relPath))
		data, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Fatalf("expected extracted file at %s: %v", relPath, err)
		}
		if string(data) != expectedContent {
			t.Errorf("content mismatch for %s: got %q, want %q", relPath, string(data), expectedContent)
		}
	}
}

func TestUnzipZipSlipDefense(t *testing.T) {
	// Create a zip with a malicious relative path trying to write outside dest
	zipPath := filepath.Join(os.TempDir(), "mistatic_zipslip.zip")
	defer os.Remove(zipPath)

	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	zw := zip.NewWriter(zf)

	// Add malicious header
	header := &zip.FileHeader{
		Name:   "../escaped.txt",
		Method: zip.Deflate,
	}
	writer, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}
	_, _ = writer.Write([]byte("malicious content"))
	_ = zw.Close()
	_ = zf.Close()

	destDir, err := os.MkdirTemp("", "mistatic_test_dest_*")
	if err != nil {
		t.Fatalf("failed to create dest temp dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	// Run unzip
	if err := unzip(zipPath, destDir); err != nil {
		t.Fatalf("unzip should succeed by skipping invalid paths: %v", err)
	}

	// Ensure escaped.txt was NOT created outside destDir
	escapedPath := filepath.Join(destDir, "..", "escaped.txt")
	if _, err := os.Stat(escapedPath); !os.IsNotExist(err) {
		_ = os.Remove(escapedPath)
		t.Fatalf("Zip slip defense failed! File was extracted outside destination directory: %s", escapedPath)
	}
}
