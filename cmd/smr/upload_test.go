package smr

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// --- Fix #1: uploadChunk must surface previously-silent failures ---

func TestUploadChunk_HTTPErrorStatusSurfacesAsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	err := uploadChunk(context.Background(), server.Client(), server.URL, "bad-key", "mod-1", "version-1", 1, []byte("chunk-data"), "test.smod")
	if err == nil {
		t.Fatal("expected an error for an HTTP 401 response, got nil")
	}
	if !strings.Contains(err.Error(), "chunk") {
		t.Errorf("expected error to mention the chunk, got: %v", err)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to mention the status code, got: %v", err)
	}
}

func TestUploadChunk_GraphQLErrorsArraySurfacesAsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"errors":[{"message":"nope"}]}`))
	}))
	defer server.Close()

	err := uploadChunk(context.Background(), server.Client(), server.URL, "key", "mod-1", "version-1", 2, []byte("chunk-data"), "test.smod")
	if err == nil {
		t.Fatal("expected an error for a GraphQL errors array, got nil")
	}
	if !strings.Contains(err.Error(), "chunk") {
		t.Errorf("expected error to mention the chunk, got: %v", err)
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("expected error to include the GraphQL error message, got: %v", err)
	}
}

func TestUploadChunk_ExplicitSuccessIsAccepted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uploadVersionPart":true}}`))
	}))
	defer server.Close()

	err := uploadChunk(context.Background(), server.Client(), server.URL, "key", "mod-1", "version-1", 3, []byte("chunk-data"), "test.smod")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestUploadChunk_MissingSuccessFieldIsRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	err := uploadChunk(context.Background(), server.Client(), server.URL, "key", "mod-1", "version-1", 5, []byte("chunk-data"), "test.smod")
	if err == nil {
		t.Fatal("expected an error when uploadVersionPart is missing (schema types it Boolean!, so fail closed), got nil")
	}
	if !strings.Contains(err.Error(), "chunk") {
		t.Errorf("expected error to mention the chunk, got: %v", err)
	}
}

// TestUploadChunk_ExplicitFailureIsRejected confirms that an explicit
// uploadVersionPart:false from the server hard-fails the chunk.
func TestUploadChunk_ExplicitFailureIsRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uploadVersionPart":false}}`))
	}))
	defer server.Close()

	err := uploadChunk(context.Background(), server.Client(), server.URL, "key", "mod-1", "version-1", 4, []byte("chunk-data"), "test.smod")
	if err == nil {
		t.Fatal("expected an error for an explicit uploadVersionPart:false, got nil")
	}
	if !strings.Contains(err.Error(), "chunk") {
		t.Errorf("expected error to mention the chunk, got: %v", err)
	}
}

// TestUploadChunk_SendsExpectedMultipartRequest is a parity check: the
// refactor into uploadChunk must not change the wire-level request (same
// operations/map/file multipart fields, same headers) that the server
// actually sees.
func TestUploadChunk_SendsExpectedMultipartRequest(t *testing.T) {
	var gotAuth, gotContentType, gotOperations, gotMap, gotFileName string
	var gotFileContent []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			t.Errorf("failed to parse multipart form: %v", err)
		}
		gotOperations = r.FormValue("operations")
		gotMap = r.FormValue("map")

		file, header, err := r.FormFile("0")
		if err != nil {
			t.Errorf("failed to read file field: %v", err)
		} else {
			defer file.Close()
			gotFileName = header.Filename
			gotFileContent, _ = io.ReadAll(file)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uploadVersionPart":true}}`))
	}))
	defer server.Close()

	chunk := []byte("hello chunk contents")
	err := uploadChunk(context.Background(), server.Client(), server.URL, "secret-key", "mod-42", "version-7", 3, chunk, "MyMod.smod")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if gotAuth != "secret-key" {
		t.Errorf("expected Authorization header %q, got %q", "secret-key", gotAuth)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("expected multipart/form-data content type, got %q", gotContentType)
	}
	if !strings.Contains(gotOperations, `"part":3`) {
		t.Errorf("expected operations field to include the part number, got %q", gotOperations)
	}
	if !strings.Contains(gotOperations, `"modId":"mod-42"`) {
		t.Errorf("expected operations field to include modId, got %q", gotOperations)
	}
	if !strings.Contains(gotMap, "variables.file") {
		t.Errorf("expected map field to reference variables.file, got %q", gotMap)
	}
	if gotFileName != "MyMod.smod" {
		t.Errorf("expected uploaded file name %q, got %q", "MyMod.smod", gotFileName)
	}
	if !bytes.Equal(gotFileContent, chunk) {
		t.Errorf("expected uploaded file contents %q, got %q", chunk, gotFileContent)
	}
}

func TestUploadChunk_RetriesTransientThenSucceeds(t *testing.T) {
	// Shrink the retry delay so the test is fast.
	orig := uploadChunkBaseDelay
	uploadChunkBaseDelay = time.Millisecond
	defer func() { uploadChunkBaseDelay = orig }()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError) // transient 5xx
			_, _ = w.Write([]byte("temporary server error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uploadVersionPart":true}}`))
	}))
	defer server.Close()

	err := uploadChunk(context.Background(), server.Client(), server.URL, "key", "mod-1", "version-1", 7, []byte("chunk-data"), "test.smod")
	if err != nil {
		t.Fatalf("expected success after transient 5xx errors were retried, got error: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Errorf("expected at least 3 attempts (2 failures + 1 success), got %d", got)
	}
}

// --- Fix #2: chunk reads must not silently short-read ---

// shortReadSeeker is a test fixture that behaves like a file for Seek/Read
// purposes but deliberately hands back at most one byte per Read call --
// the same kind of short read that a bare f.Read(chunk) is not guaranteed
// to survive.
type shortReadSeeker struct {
	data []byte
	pos  int64
}

func (s *shortReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		s.pos = offset
	case io.SeekCurrent:
		s.pos += offset
	case io.SeekEnd:
		s.pos = int64(len(s.data)) + offset
	default:
		return 0, fmt.Errorf("unsupported whence: %d", whence)
	}
	return s.pos, nil
}

func (s *shortReadSeeker) Read(p []byte) (int, error) {
	if s.pos >= int64(len(s.data)) {
		return 0, io.EOF
	}
	n := copy(p, s.data[s.pos:s.pos+1])
	s.pos += int64(n)
	return n, nil
}

// TestReadChunk_FillsBufferDespiteShortReads exercises the actual production
// read path (readChunk, the helper the upload loop calls) against a source
// that only ever hands back one byte per Read call. The old bare
// `f.Read(chunk)` call was not guaranteed to fill the buffer in one call;
// readChunk must loop (via io.ReadFull) until the buffer is completely and
// correctly filled.
func TestReadChunk_FillsBufferDespiteShortReads(t *testing.T) {
	original := []byte("this is definitely more than one byte of chunk data, repeated for length")
	src := &shortReadSeeker{data: original}

	buf := make([]byte, len(original))
	if err := readChunk(src, 0, buf); err != nil {
		t.Fatalf("expected readChunk to fill the buffer despite short reads, got error: %v", err)
	}
	if !bytes.Equal(buf, original) {
		t.Fatalf("expected buffer to round-trip byte-identically, got %q, want %q", buf, original)
	}
}

// TestReadChunk_HonorsOffset confirms readChunk seeks to the requested chunk
// offset before reading, matching the per-iteration seek the upload loop
// relies on to walk through the file one chunk at a time.
func TestReadChunk_HonorsOffset(t *testing.T) {
	original := []byte("0123456789ABCDEF")
	src := &shortReadSeeker{data: original}

	buf := make([]byte, 4)
	if err := readChunk(src, 10, buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := original[10:14]
	if !bytes.Equal(buf, want) {
		t.Fatalf("got %q, want %q", buf, want)
	}
}

// TestReadChunk_SurfacesShortReadError confirms the failure mode the old code
// could not detect: when fewer bytes are available than the chunk size
// requires, readChunk must return an error instead of silently returning a
// short, zero-padded buffer.
func TestReadChunk_SurfacesShortReadError(t *testing.T) {
	src := &shortReadSeeker{data: []byte("short")}

	buf := make([]byte, 15) // longer than the source has available
	if err := readChunk(src, 0, buf); err == nil {
		t.Fatal("expected readChunk to return an error when the source is shorter than the buffer, got nil")
	}
}

// --- Fix #3: .smod validation before upload ---

func TestValidateSmodFile_AcceptsRealZipNamedSmod(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "MyMod.smod")
	writeMinimalZip(t, path)

	if err := validateSmodFile(path); err != nil {
		t.Fatalf("expected a valid .smod zip to pass validation, got error: %v", err)
	}
}

func TestValidateSmodFile_AcceptsUppercaseExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "MyMod.SMOD")
	writeMinimalZip(t, path)

	if err := validateSmodFile(path); err != nil {
		t.Fatalf("expected case-insensitive extension match, got error: %v", err)
	}
}

func TestValidateSmodFile_AcceptsZipExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "MyMod.zip")
	writeMinimalZip(t, path)

	if err := validateSmodFile(path); err != nil {
		t.Fatalf("expected a .zip mod archive to pass validation, got error: %v", err)
	}
}

func TestValidateSmodFile_RejectsWrongExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "MyMod.txt")
	writeMinimalZip(t, path) // valid zip bytes, but the extension is wrong

	if err := validateSmodFile(path); err == nil {
		t.Fatal("expected a .txt file to be rejected regardless of content, got nil error")
	}
}

func TestValidateSmodFile_RejectsNonZipContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "MyMod.smod")
	if err := os.WriteFile(path, []byte("this is not a zip archive"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := validateSmodFile(path); err == nil {
		t.Fatal("expected a non-zip .smod file to be rejected, got nil error")
	}
}

func writeMinimalZip(t *testing.T, path string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test zip file: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	w, err := zw.Create("mod.uplugin")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := w.Write([]byte(`{"FriendlyName":"Test"}`)); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
}
