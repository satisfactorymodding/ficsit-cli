package utils

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/satisfactorymodding/ficsit-cli/cli/disk"
)

type GenericUpdate struct {
	ModReference *string
	Progress     float64
}

func SHA256Data(f io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", errors.Wrap(err, "failed to compute hash")
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func ExtractMod(f io.ReaderAt, size int64, location string, hash string, updates chan GenericUpdate, d disk.Disk) error {
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

	for i, file := range reader.File {
		if !file.FileInfo().IsDir() {
			outFileLocation := filepath.Join(location, file.Name)

			if err := d.MkDir(filepath.Dir(outFileLocation)); err != nil {
				return errors.Wrap(err, "failed to create mod directory: "+location)
			}

			if err := writeZipFile(outFileLocation, file, d); err != nil {
				return err
			}
		}

		if updates != nil {
			select {
			case updates <- GenericUpdate{Progress: float64(i) / float64(len(reader.File)-1)}:
			default:
			}
		}
	}

	if err := d.Write(hashFile, []byte(hash)); err != nil {
		return errors.Wrap(err, "failed to write .smm mod hash file")
	}

	if updates != nil {
		select {
		case updates <- GenericUpdate{Progress: 1}:
		default:
		}
	}

	return nil
}

func writeZipFile(outFileLocation string, file *zip.File, d disk.Disk) error {
	outFile, err := d.Open(outFileLocation, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return errors.Wrap(err, "failed to write to file: "+outFileLocation)
	}

	defer outFile.Close()

	inFile, err := file.Open()
	if err != nil {
		return errors.Wrap(err, "failed to process mod zip")
	}

	if _, err := io.Copy(outFile, inFile); err != nil {
		return errors.Wrap(err, "failed to write to file: "+outFileLocation)
	}

	return nil
}
