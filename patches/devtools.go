package patches

import (
	"bytes"
	"os"
	"strings"
)

func PatchDevTools(offlineBnkFile string, enable bool) error {
	file, err := os.OpenFile(offlineBnkFile, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	fileBuffer := bytes.Buffer{}
	fileBuffer.ReadFrom(file)
	fileContent := fileBuffer.String()

	firstLocation := strings.Index(fileContent, "app-developer")
	firstPatchLocation := int64(firstLocation + 14)

	secondLocation := strings.LastIndex(fileContent, "app-developer")
	secondPatchLocation := int64(secondLocation + 15)

	if enable {
		file.WriteAt([]byte{50}, firstPatchLocation)
		file.WriteAt([]byte{50}, secondPatchLocation)
	} else {
		file.WriteAt([]byte{30}, firstPatchLocation)
		file.WriteAt([]byte{30}, secondPatchLocation)
	}

	return nil
}
