package patches

import (
	"log"
	"os"
	"path"
)

func PatchV8ContextSnapshot(contextFile string, xpuiAppPath string) error {
	log.Printf("Extracting xpui-modules.js from: %s, to: %s\n", contextFile, xpuiAppPath)

	startMarker := []byte("var __webpack_modules__={")
	endMarker := []byte("xpui-modules.js.map")

	embeddedString, _, _, err := ReadStringFromUTF16Binary(contextFile, startMarker, endMarker)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(xpuiAppPath, "xpui-modules.js"), []byte(embeddedString), 0700)
	if err != nil {
		return err
	}

	return nil
}
