package smr

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

const uploadVersionPartGQL = `mutation UploadVersionPart($modId: ModID!, $versionId: VersionID!, $part: Int!, $file: Upload!) {
  uploadVersionPart(modId: $modId, versionId: $versionId, part: $part, file: $file)
}`

// maxErrorBodySnippet caps how much of an HTTP/GraphQL response body is
// embedded in an error message, so an unexpected large response doesn't
// bloat logs.
const maxErrorBodySnippet = 500

func init() {
	Cmd.AddCommand(uploadCmd)
}

var uploadCmd = &cobra.Command{
	Use:   "upload [flags] <mod-id> <file> <changelog...>",
	Short: "Upload a new mod version",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		chunkSize := viper.GetInt64("chunk-size")
		if chunkSize < 1000000 {
			return errors.New("chunk size cannot be smaller than 1MB")
		}

		var versionStability ficsit.VersionStabilities
		switch viper.GetString("stability") {
		case "alpha":
			versionStability = ficsit.VersionStabilitiesAlpha
		case "beta":
			versionStability = ficsit.VersionStabilitiesBeta
		case "release":
			versionStability = ficsit.VersionStabilitiesRelease
		default:
			return errors.New("invalid version stability: " + viper.GetString("stability"))
		}

		modID := args[0]
		filePath := args[1]
		changelog := strings.Join(args[2:], " ")

		stat, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("failed to stat file: %w", err)
		}

		global, err := cli.InitCLI(true)
		if err != nil {
			return err
		}

		if stat.IsDir() {
			return errors.New("file cannot be a directory")
		}

		if err := validateSmodFile(filePath); err != nil {
			return err
		}

		logBase := slog.With(slog.String("mod-id", modID), slog.String("path", filePath))
		logBase.Info("creating a new mod version")

		createdVersion, err := ficsit.CreateVersion(cmd.Context(), global.APIClient, modID)
		if err != nil {
			return err
		}

		logBase = logBase.With(slog.String("version-id", createdVersion.GetVersionID()))
		logBase.Info("received version id")

		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		httpClient := &http.Client{}
		uploadURL := viper.GetString("api-base") + viper.GetString("graphql-api")
		apiKey := viper.GetString("api-key")

		// TODO Parallelize chunk uploading
		chunkCount := int(math.Ceil(float64(stat.Size()) / float64(chunkSize)))
		for i := 0; i < chunkCount; i++ {
			// part is 1-indexed (matches the GraphQL "part" variable sent to the
			// server) and is reused for both logging and the upload call so the
			// two can never drift apart.
			part := i + 1
			chunkLog := logBase.With(slog.Int("chunk", part))
			chunkLog.Info("uploading chunk")

			offset := int64(i) * chunkSize

			bufferSize := chunkSize
			if offset+chunkSize > stat.Size() {
				bufferSize = stat.Size() - offset
			}

			chunk := make([]byte, bufferSize)
			if err := readChunk(f, offset, chunk); err != nil {
				return err
			}

			if err := uploadChunk(cmd.Context(), httpClient, uploadURL, apiKey, modID, createdVersion.GetVersionID(), part, chunk, filepath.Base(filePath)); err != nil {
				chunkLog.Error("failed to upload chunk", slog.Any("err", err))
				return err
			}
		}

		logBase.Info("finalizing uploaded version")

		finalizeSuccess, err := ficsit.FinalizeCreateVersion(cmd.Context(), global.APIClient, modID, createdVersion.GetVersionID(), ficsit.NewVersion{
			Changelog: changelog,
			Stability: versionStability,
		})
		if err != nil {
			return err
		}

		if !finalizeSuccess.GetSuccess() {
			logBase.Error("failed to finalize version upload")
		}

		time.Sleep(time.Second * 1)

		for {
			logBase.Info("checking version upload state")
			state, err := ficsit.CheckVersionUploadState(cmd.Context(), global.APIClient, modID, createdVersion.GetVersionID())
			if err != nil {
				logBase.Error("failed to check version upload state", slog.Any("err", err))
				return fmt.Errorf("failed to check version upload state after finalizing upload: %w", err)
			}

			if state == nil || state.GetState().Version.Id == "" {
				time.Sleep(time.Second * 10)
				continue
			}

			if state.GetState().Auto_approved {
				logBase.Info("version successfully uploaded and auto-approved")
				break
			}

			logBase.Info("version successfully uploaded, but has to be scanned for viruses, which may take up to 15 minutes")
			break
		}

		return nil
	},
}

// readChunk seeks to offset and reads exactly len(buf) bytes into buf,
// failing loudly instead of silently returning a short, zero-padded buffer
// (the bug in the original bare f.Read(chunk) call, which was not
// guaranteed to fill the buffer in one call).
func readChunk(r io.ReadSeeker, offset int64, buf []byte) error {
	if _, err := r.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to chunk offset: %w", err)
	}

	if _, err := io.ReadFull(r, buf); err != nil {
		return fmt.Errorf("failed to read from chunk offset: %w", err)
	}

	return nil
}

// uploadVersionPartResponse is the (loosely known) shape of the GraphQL
// response body for the UploadVersionPart mutation.
type uploadVersionPartResponse struct {
	Data *struct {
		// UploadVersionPart is a pointer so an explicit `false` (failure) can be
		// told apart from a missing/null value. The GraphQL schema types this as
		// `Boolean!` (non-null), so on a 2xx response with no GraphQL errors it
		// must be present and concrete; uploadChunk therefore fails closed if it
		// is missing/null rather than silently assuming success.
		UploadVersionPart *bool `json:"uploadVersionPart"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// uploadChunkMaxAttempts and uploadChunkBaseDelay control the per-chunk retry
// policy. They are package variables (not consts) so tests can shrink the delay.
var (
	uploadChunkMaxAttempts uint = 4
	uploadChunkBaseDelay        = 2 * time.Second
)

// uploadChunk sends a single chunk of a mod file to the SMR upload endpoint as a
// GraphQL multipart request (the UploadVersionPart mutation) and validates the
// response. It never silently swallows a failed chunk. Transient failures (a
// transport error such as a dropped/corrupted TLS record, or a 5xx response) are
// retried up to uploadChunkMaxAttempts times with exponential backoff; permanent
// failures (a 4xx status, a GraphQL `errors` array, or an explicit
// `uploadVersionPart: false`) are NOT retried, since they will not succeed on a
// second attempt.
func uploadChunk(ctx context.Context, client *http.Client, url, apiKey, modID, versionID string, part int, chunk []byte, fileName string) error {
	return retry.Do(
		func() error {
			return sendChunkOnce(ctx, client, url, apiKey, modID, versionID, part, chunk, fileName)
		},
		retry.Attempts(uploadChunkMaxAttempts),
		retry.Delay(uploadChunkBaseDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn("retrying chunk upload after a transient error",
				slog.Int("chunk", part), slog.Uint64("failed_attempt", uint64(n+1)), slog.Any("err", err))
		}),
	)
}

// sendChunkOnce performs a single upload attempt for one chunk. Retryable errors
// (transport failures, 5xx) are returned as plain errors; permanent errors are
// wrapped with retry.Unrecoverable so retry.Do stops immediately.
func sendChunkOnce(ctx context.Context, client *http.Client, url, apiKey, modID, versionID string, part int, chunk []byte, fileName string) error {
	operationBody, err := json.Marshal(map[string]interface{}{
		"query": uploadVersionPartGQL,
		"variables": map[string]interface{}{
			"modId":     modID,
			"versionId": versionID,
			"part":      part,
			"file":      nil,
		},
	})
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to serialize operation body: %w", part, err))
	}

	mapBody, err := json.Marshal(map[string]interface{}{
		"0": []string{"variables.file"},
	})
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to serialize map body: %w", part, err))
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	operations, err := writer.CreateFormField("operations")
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to create operations field: %w", part, err))
	}

	if _, err := operations.Write(operationBody); err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to write to operation field: %w", part, err))
	}

	mapField, err := writer.CreateFormField("map")
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to create map field: %w", part, err))
	}

	if _, err := mapField.Write(mapBody); err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to write to map field: %w", part, err))
	}

	filePart, err := writer.CreateFormFile("0", fileName)
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to create file field: %w", part, err))
	}

	if _, err := io.Copy(filePart, bytes.NewReader(chunk)); err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to write to file field: %w", part, err))
	}

	if err := writer.Close(); err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to close body writer: %w", part, err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to create request: %w", part, err))
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		// Transport-level failure (e.g. "tls: bad record MAC", connection reset) --
		// return plain (retryable) so retry.Do can try again.
		return fmt.Errorf("chunk %d: request failed: %w", part, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// Reading the response failed mid-stream -- treat as a retryable transport issue.
		return fmt.Errorf("chunk %d: failed to read response body: %w", part, err)
	}

	if resp.StatusCode >= 500 {
		// Server-side transient error -- retryable.
		return fmt.Errorf("chunk %d: upload failed with server status %d: %s", part, resp.StatusCode, errorBodySnippet(respBody))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 4xx (bad request / auth) -- permanent; a retry will not help.
		return retry.Unrecoverable(fmt.Errorf("chunk %d: upload failed with status %d: %s", part, resp.StatusCode, errorBodySnippet(respBody)))
	}

	var parsed uploadVersionPartResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: failed to parse response body: %w (body: %s)", part, err, errorBodySnippet(respBody)))
	}

	if len(parsed.Errors) > 0 {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: server reported an error: %s", part, parsed.Errors[0].Message))
	}

	if parsed.Data == nil || parsed.Data.UploadVersionPart == nil {
		// The schema types uploadVersionPart as Boolean! (non-null), so a 2xx
		// response with no GraphQL errors must carry a concrete value. Treat a
		// missing/null result as a failure rather than silently succeeding.
		return retry.Unrecoverable(fmt.Errorf("chunk %d: upload response contained no uploadVersionPart result (status %d, body: %s)", part, resp.StatusCode, errorBodySnippet(respBody)))
	}

	if !*parsed.Data.UploadVersionPart {
		return retry.Unrecoverable(fmt.Errorf("chunk %d: server reported uploadVersionPart failure", part))
	}

	return nil
}

// errorBodySnippet truncates a response body for inclusion in an error
// message so an unexpected large response doesn't bloat logs.
func errorBodySnippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > maxErrorBodySnippet {
		return s[:maxErrorBodySnippet] + "...(truncated)"
	}
	return s
}

// validateSmodFile performs a lightweight validation of a mod upload file before
// it is uploaded: it must have a .smod or .zip extension and must open as a
// valid zip archive. SMR's distributable format is .smod (itself a zip), while
// Alpakit's packaged upload artifact is a .zip -- both are accepted. This
// intentionally does NOT enforce any deeper internal structure (e.g. a specific
// .uplugin entry), since that could false-reject otherwise valid files.
func validateSmodFile(filePath string) error {
	switch ext := strings.ToLower(filepath.Ext(filePath)); ext {
	case ".smod", ".zip":
		// accepted
	default:
		return fmt.Errorf("%q is not a valid mod upload file: unexpected extension %q, expected \".smod\" or \".zip\"", filePath, ext)
	}

	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("%q is not a valid mod upload file: %w", filePath, err)
	}
	defer zr.Close()

	slog.Debug("validated mod upload archive", slog.String("path", filePath), slog.Int("entries", len(zr.File)))

	return nil
}

func init() {
	uploadCmd.PersistentFlags().Int64("chunk-size", 10000000, "Size of chunks to split uploaded mod in bytes")
	uploadCmd.PersistentFlags().String("stability", "release", "Stability of the uploaded mod (alpha, beta, release)")

	_ = viper.BindPFlag("chunk-size", uploadCmd.PersistentFlags().Lookup("chunk-size"))
	_ = viper.BindPFlag("stability", uploadCmd.PersistentFlags().Lookup("stability"))
}
