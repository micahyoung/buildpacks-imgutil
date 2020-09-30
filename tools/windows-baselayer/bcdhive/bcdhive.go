package bcdhive

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-samfira/go-hivex"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/tools/go/packages"
)

func Generate() ([]byte, error) {
	pkgs, err := packages.Load(&packages.Config{}, "github.com/gabriel-samfira/go-hivex")
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 || len(pkgs[0].GoFiles) != 1 {
		return nil, errors.New("hivex module root not found")
	}
	hivexRootPath := filepath.Dir(pkgs[0].GoFiles[0])
	minimalHivePath := filepath.Join(hivexRootPath, "testdata", "minimal")
	minimalHiveFile, err := os.Open(minimalHivePath)
	if err != nil {
		return nil, err
	}

	hiveFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	defer hiveFile.Close()
	defer os.Remove(hiveFile.Name())

	if _, err := io.Copy(hiveFile, minimalHiveFile); err != nil {
		return nil, err
	}

	if _, err := hiveFile.Seek(0, 0); err != nil {
		return nil, err
	}

	h, err := hivex.NewHivex(hiveFile.Name(), hivex.WRITE)
	if err != nil {
		return nil, errors.Wrap(err, "opening hive file")
	}
	defer h.Close()

	if err := addBCDHiveEntries(h); err != nil {
		return nil, err
	}

	if _, err := hiveFile.Seek(0, 0); err != nil {
		return nil, err
	}

	hiveBytes, err := ioutil.ReadAll(hiveFile)
	if err != nil {
		return nil, err
	}

	return hiveBytes, nil
}

func toUtf16LE(inStr string) []byte {
	utf16Encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	outStr, _ := utf16Encoder.String(inStr)
	outBytes := append([]byte(outStr), []byte("\x00\x00")...)
	return outBytes
}

func addBCDHiveEntries(h *hivex.Hivex) error {
	entries := map[string][]hivex.HiveValue{
		"Description": {
			{
				Type:  hivex.RegDword,
				Key:   "FirmwareModified",
				Value: []byte("\x01\x00\x00\x00"),
			},
			{
				Type:  hivex.RegSz,
				Key:   "KeyName",
				Value: toUtf16LE("BCD00000000"),
			},
		},
		`Objects/{6a6c1f1b-59d4-11ea-9438-9402e6abd998}/Description`: {
			{
				Type:  hivex.RegDword,
				Key:   "Type",
				Value: []byte("\x03\x00\x20\x10"),
			},
		},
		`Objects/{6a6c1f1b-59d4-11ea-9438-9402e6abd998}/Elements/12000004`: {
			{
				Type:  hivex.RegSz,
				Key:   "Element",
				Value: toUtf16LE("buildpacks.io"),
			},
		},
		`Objects/{9dea862c-5cdd-4e70-acc1-f32b344d4795}/Description`: {
			{
				Type:  hivex.RegDword,
				Key:   "Type",
				Value: []byte("\x02\x00\x10\x10"),
			},
		},
		`Objects/{9dea862c-5cdd-4e70-acc1-f32b344d4795}/Elements/23000003`: {
			{
				Type:  hivex.RegSz,
				Key:   "Element",
				Value: toUtf16LE("{6a6c1f1b-59d4-11ea-9438-9402e6abd998}"),
			},
		},
	}

	for p, vals := range entries {
		node, err := h.Root()
		if err != nil {
			return err
		}

		pathChildren := strings.Split(p, "/")
		for _, childPath := range pathChildren {
			existingRoot, err := h.NodeGetChild(node, childPath)
			if err != nil {
				return err
			}

			if existingRoot != 0 {
				node = existingRoot
				continue
			}

			node, err = h.NodeAddChild(node, childPath)
			if err != nil {
				return err
			}
		}

		if _, err := h.NodeSetValues(node, vals); err != nil {
			return err
		}
	}

	if _, err := h.Commit(); err != nil {
		return err
	}
	return nil
}
