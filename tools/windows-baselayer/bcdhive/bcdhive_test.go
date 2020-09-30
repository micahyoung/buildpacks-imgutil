package bcdhive_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/buildpack/pack/tools/windows-baselayer-bcd/bcdhive"
	"gotest.tools/assert"
)

type value struct {
	id      int64
	name    string
	vString string
	vType   int64
}

type node struct {
	name   string
	id     int64
	values []value
}

func TestGenerate(t *testing.T) {
	output, err := bcdhive.Generate()
	assert.NilError(t, err)

	hiveFile, err := ioutil.TempFile("", "")
	assert.NilError(t, err)
	defer hiveFile.Close()
	defer os.Remove(hiveFile.Name())

	_, err = io.Copy(hiveFile, bytes.NewBuffer(output))
	assert.NilError(t, err)

	cmd := exec.Command("hivexregedit", "--export", hiveFile.Name(), "Description")
	descriptionRegOutput, err := cmd.CombinedOutput()
	assert.NilError(t, err)

	assert.Equal(t, string(descriptionRegOutput),
		`Windows Registry Editor Version 5.00

[\Description]
"FirmwareModified"=dword:00000001
"KeyName"=hex(1):42,00,43,00,44,00,30,00,30,00,30,00,30,00,30,00,30,00,30,00,30,00,00,00

`)

	cmd = exec.Command("hivexregedit", "--export", hiveFile.Name(), "Objects")
	objectsRegOutput, err := cmd.CombinedOutput()
	assert.NilError(t, err)

	assert.Equal(t, string(objectsRegOutput),
		`Windows Registry Editor Version 5.00

[\Objects]

[\Objects\{6a6c1f1b-59d4-11ea-9438-9402e6abd998}]

[\Objects\{6a6c1f1b-59d4-11ea-9438-9402e6abd998}\Description]
"Type"=dword:10200003

[\Objects\{6a6c1f1b-59d4-11ea-9438-9402e6abd998}\Elements]

[\Objects\{6a6c1f1b-59d4-11ea-9438-9402e6abd998}\Elements\12000004]
"Element"=hex(1):62,00,75,00,69,00,6c,00,64,00,70,00,61,00,63,00,6b,00,73,00,2e,00,69,00,6f,00,00,00

[\Objects\{9dea862c-5cdd-4e70-acc1-f32b344d4795}]

[\Objects\{9dea862c-5cdd-4e70-acc1-f32b344d4795}\Description]
"Type"=dword:10100002

[\Objects\{9dea862c-5cdd-4e70-acc1-f32b344d4795}\Elements]

[\Objects\{9dea862c-5cdd-4e70-acc1-f32b344d4795}\Elements\23000003]
"Element"=hex(1):7b,00,36,00,61,00,36,00,63,00,31,00,66,00,31,00,62,00,2d,00,35,00,39,00,64,00,34,00,2d,00,31,00,31,00,65,00,61,00,2d,00,39,00,34,00,33,00,38,00,2d,00,39,00,34,00,30,00,32,00,65,00,36,00,61,00,62,00,64,00,39,00,39,00,38,00,7d,00,00,00

`)

}
