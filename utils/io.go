package utils

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/satisfactorymodding/ficsit-cli/cli/disk"
)

type GenericProgress struct {
	Completed int64
	Total     int64
}

func (gp GenericProgress) Percentage() float64 {
	if gp.Total == 0 {
		return 0
	}
	return float64(gp.Completed) / float64(gp.Total)
}

type Progresser struct {
	io.Reader
	Updates chan<- GenericProgress
	Total   int64
	Running int64
}

func (pt *Progresser) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.Running += int64(n)

	if err == nil {
		if pt.Updates != nil {
			select {
			case pt.Updates <- GenericProgress{Completed: pt.Running, Total: pt.Total}:
			default:
			}
		}
	}

	if err == io.EOF {
		return n, io.EOF
	}

	if err != nil {
		return 0, fmt.Errorf("failed to read: %w", err)
	}

	return n, nil
}

func SHA256Data(f io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func ExtractMod(f io.ReaderAt, size int64, location string, hash string, updates chan<- GenericProgress, d disk.Disk) error {
	hashFile := filepath.Join(location, ".smm")

	exists, err := d.Exists(hashFile)
	if err != nil {
		return err
	}

	if exists {
		hashBytes, err := d.Read(hashFile)
		if err != nil {
			return fmt.Errorf("failed to read .smm mod hash file: %w", err)
		}

		if hash == string(hashBytes) {
			return nil
		}
	}

	exists, err = d.Exists(location)
	if err != nil {
		return err
	}

	if exists {
		if err := d.Remove(location); err != nil {
			return fmt.Errorf("failed to remove directory: %s: %w", location, err)
		}

		if err := d.MkDir(location); err != nil {
			return fmt.Errorf("failed to create mod directory: %s: %w", location, err)
		}
	}

	reader, err := zip.NewReader(f, size)
	if err != nil {
		return fmt.Errorf("failed to read file as zip: %w", err)
	}

	totalSize := int64(0)

	for _, file := range reader.File {
		totalSize += int64(file.UncompressedSize64)
	}

	totalExtracted := int64(0)
	totalExtractedPtr := &totalExtracted

	channelUsers := sync.WaitGroup{}

	if updates != nil {
		defer func() {
			channelUsers.Wait()
		}()
	}

	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			outFileLocation := filepath.Join(location, file.Name)

			if err := d.MkDir(filepath.Dir(outFileLocation)); err != nil {
				return fmt.Errorf("failed to create mod directory: %s: %w", location, err)
			}

			var fileUpdates chan GenericProgress
			if updates != nil {
				fileUpdates = make(chan GenericProgress)
				channelUsers.Add(1)
				go func() {
					defer channelUsers.Done()
					for fileUpdate := range fileUpdates {
						updates <- GenericProgress{
							Completed: atomic.LoadInt64(totalExtractedPtr) + fileUpdate.Completed,
							Total:     totalSize,
						}
					}
				}()
			}

			if err := writeZipFile(outFileLocation, file, d, fileUpdates); err != nil {
				return err
			}

			atomic.AddInt64(totalExtractedPtr, int64(file.UncompressedSize64))
		}
	}

	if err := d.Write(hashFile, []byte(hash)); err != nil {
		return fmt.Errorf("failed to write .smm mod hash file: %w", err)
	}

	if updates != nil {
		updates <- GenericProgress{Completed: totalSize, Total: totalSize}
	}

	return nil
}

func writeZipFile(outFileLocation string, file *zip.File, d disk.Disk, updates chan<- GenericProgress) error {
	if updates != nil {
		defer close(updates)
	}

	outFile, err := d.Open(outFileLocation, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return fmt.Errorf("failed to write to file: %s: %w", outFileLocation, err)
	}

	defer outFile.Close()

	inFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to process mod zip: %w", err)
	}
	defer inFile.Close()

	progressInReader := &Progresser{
		Reader:  inFile,
		Total:   int64(file.UncompressedSize64),
		Updates: updates,
	}

	if _, err := io.Copy(outFile, progressInReader); err != nil {
		return fmt.Errorf("failed to write to file: %s: %w", outFileLocation, err)
	}

	return nil
}
