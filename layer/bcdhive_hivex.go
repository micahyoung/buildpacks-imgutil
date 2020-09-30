//+build hivex

package layer

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

var orderedEntries = []entry{
	{"Description", []hivex.HiveValue{
		{
			Type:  hivex.RegDword,
			Key:   "FirmwareModified",
			Value: toDword(0x00, 0x00, 0x00, 0x01),
		},
		{
			Type:  hivex.RegSz,
			Key:   "KeyName",
			Value: toRegSz("BCD00000000"),
		},
	}},
	{`Objects/{6a6c1f1b-59d4-11ea-9438-9402e6abd998}/Description`, []hivex.HiveValue{
		{
			Type:  hivex.RegDword,
			Key:   "Type",
			Value: toDword(0x10, 0x20, 0x00, 0x03),
		},
	}},
	{`Objects/{6a6c1f1b-59d4-11ea-9438-9402e6abd998}/Elements/12000004`, []hivex.HiveValue{
		{
			Type:  hivex.RegSz,
			Key:   "Element",
			Value: toRegSz("buildpacks.io"),
		},
	}},
	{`Objects/{9dea862c-5cdd-4e70-acc1-f32b344d4795}/Description`, []hivex.HiveValue{
		{
			Type:  hivex.RegDword,
			Key:   "Type",
			Value: toDword(0x10, 0x10, 0x00, 0x02),
		},
	}},
	{`Objects/{9dea862c-5cdd-4e70-acc1-f32b344d4795}/Elements/23000003`, []hivex.HiveValue{
		{
			Type:  hivex.RegSz,
			Key:   "Element",
			Value: toRegSz("{6a6c1f1b-59d4-11ea-9438-9402e6abd998}"),
		},
	}},
}

func toDword(values ...int) []byte {
	var result []byte
	for _, val := range values {
		result = append([]byte{byte(val)}, result...)
	}
	return result
}

type entry struct {
	path       string
	hiveValues []hivex.HiveValue
}

func BaseLayerBCD() ([]byte, error) {
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

func toRegSz(inStr string) []byte {
	utf16Encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	outStr, _ := utf16Encoder.String(inStr)
	outBytes := append([]byte(outStr), []byte("\x00\x00")...)
	return outBytes
}

func addBCDHiveEntries(h *hivex.Hivex) error {
	for _, ent := range orderedEntries {
		node, err := h.Root()
		if err != nil {
			return err
		}

		pathChildren := strings.Split(ent.path, "/")
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

		if _, err := h.NodeSetValues(node, ent.hiveValues); err != nil {
			return err
		}
	}

	if _, err := h.Commit(); err != nil {
		return err
	}
	return nil
}
