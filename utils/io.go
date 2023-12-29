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

	"github.com/satisfactorymodding/ficsit-cli/cli/disk"
)

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

	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			outFileLocation := filepath.Join(location, file.Name)

			if err := d.MkDir(filepath.Dir(outFileLocation)); err != nil {
				return fmt.Errorf("failed to create mod directory: %s: %w", location, err)
			}

			channelUsers := sync.WaitGroup{}

			var fileUpdates chan GenericProgress
			if updates != nil {
				fileUpdates = make(chan GenericProgress)
				channelUsers.Add(1)
				beforeProgress := totalExtracted
				go func() {
					defer channelUsers.Done()
					for fileUpdate := range fileUpdates {
						updates <- GenericProgress{
							Completed: beforeProgress + fileUpdate.Completed,
							Total:     totalSize,
						}
					}
				}()
			}

			if err := writeZipFile(outFileLocation, file, d, fileUpdates); err != nil {
				channelUsers.Wait()
				return err
			}

			channelUsers.Wait()

			totalExtracted += int64(file.UncompressedSize64)
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

	progressInWriter := &Progresser{
		Total:   int64(file.UncompressedSize64),
		Updates: updates,
	}

	if _, err := io.Copy(io.MultiWriter(outFile, progressInWriter), inFile); err != nil {
		return fmt.Errorf("failed to write to file: %s: %w", outFileLocation, err)
	}

	return nil
}
