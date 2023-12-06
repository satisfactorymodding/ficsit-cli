package utils

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

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

	return n, errors.Wrap(err, "failed to read")
}

func SHA256Data(f io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", errors.Wrap(err, "failed to compute hash")
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func ExtractMod(f io.ReaderAt, size int64, location string, hash string, updates chan<- GenericProgress, d disk.Disk) error {
	hashFile := filepath.Join(location, ".smm")
	hashBytes, err := d.Read(hashFile)
	if err != nil {
		if !d.IsNotExist(err) {
			return errors.Wrap(err, "failed to read .smm mod hash file")
		}
	} else {
		if hash == string(hashBytes) {
			return nil
		}
	}

	if err := d.MkDir(location); err != nil {
		if !d.IsExist(err) {
			return errors.Wrap(err, "failed to create mod directory: "+location)
		}

		if err := d.Remove(location); err != nil {
			return errors.Wrap(err, "failed to remove directory: "+location)
		}

		if err := d.MkDir(location); err != nil {
			return errors.Wrap(err, "failed to create mod directory: "+location)
		}
	}

	reader, err := zip.NewReader(f, size)
	if err != nil {
		return errors.Wrap(err, "failed to read file as zip")
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
				return errors.Wrap(err, "failed to create mod directory: "+location)
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
		return errors.Wrap(err, "failed to write .smm mod hash file")
	}

	if updates != nil {
		select {
		case updates <- GenericProgress{Completed: totalSize, Total: totalSize}:
		default:
		}
	}

	return nil
}

func writeZipFile(outFileLocation string, file *zip.File, d disk.Disk, updates chan<- GenericProgress) error {
	if updates != nil {
		defer close(updates)
	}

	outFile, err := d.Open(outFileLocation, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return errors.Wrap(err, "failed to write to file: "+outFileLocation)
	}

	defer outFile.Close()

	inFile, err := file.Open()
	if err != nil {
		return errors.Wrap(err, "failed to process mod zip")
	}
	defer inFile.Close()

	progressInReader := &Progresser{
		Reader:  inFile,
		Total:   int64(file.UncompressedSize64),
		Updates: updates,
	}

	if _, err := io.Copy(outFile, progressInReader); err != nil {
		return errors.Wrap(err, "failed to write to file: "+outFileLocation)
	}

	return nil
}
