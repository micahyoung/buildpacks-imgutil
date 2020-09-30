package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"text/template"

	"github.com/buildpack/pack/tools/windows-baselayer-bcd/baselayer"
	"github.com/buildpack/pack/tools/windows-baselayer-bcd/bcdhive"
)

const tmpl = `package {{.PackageName}}

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
)

const encodedBytes="{{.EncodedBytes}}"

func {{.FuncName}}() ([]byte, error) {
	gzipBytes, err := base64.StdEncoding.DecodeString(encodedBytes)
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(bytes.NewBuffer(gzipBytes))
	if err != nil {
		return nil, err
	}

	decodedBytes, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}

	return decodedBytes, nil
}
`

type tmplData struct{ PackageName, FuncName, EncodedBytes string }

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <package name> <func name>\n", os.Args[0])
	}

	// grab last 2 args to support `go run main.go -- pkg out.go`
	nArgs := len(os.Args)
	outputPackageName := os.Args[nArgs-2]
	outputFuncName := os.Args[nArgs-1]
	if err := run(outputPackageName, outputFuncName); err != nil {
		log.Fatal(err)
	}
}

func run(outputPackageName, outputFuncName string) error {
	bcdBytes, err := bcdhive.Generate()
	if err != nil {
		return err
	}

	baseLayerBytes, err := baselayer.Generate(bcdBytes)
	if err != nil {
		return err
	}

	encodedBaseLayer, err := encodeLayer(baseLayerBytes)
	if err != nil {
		return err
	}

	out, err := generateGoSource(outputPackageName, outputFuncName, encodedBaseLayer)
	if err != nil {
		return err
	}

	if _, err := fmt.Print(out); err != nil {
		return err
	}

	return nil
}

func encodeLayer(rawBytes []byte) (string, error) {
	gzipBuffer := &bytes.Buffer{}
	gzipWriter, err := gzip.NewWriterLevel(gzipBuffer, gzip.BestCompression)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(gzipWriter, bytes.NewBuffer(rawBytes)); err != nil {
		return "", err
	}

	if err := gzipWriter.Close(); err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(gzipBuffer.Bytes())
	return encoded, nil
}

func generateGoSource(packageName, funcName, bcdData string) (string, error) {
	t := template.Must(template.New("tmpl").Parse(tmpl))

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, tmplData{packageName, funcName, bcdData}); err != nil {
		return "", err
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return "", err
	}

	return string(src), nil
}
