package patches

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func AdditionalOptions(appsFolderPath string) {
	jsModifiers := []func(path string){
		insertExpFeatures,
		insertHomeConfig,
	}
	filesToModified := map[string][]func(path string){
		filepath.Join(appsFolderPath, "xpui", "index.html"): {
			htmlMod,
		},
		filepath.Join(appsFolderPath, "xpui", "xpui.js"):         jsModifiers,
		filepath.Join(appsFolderPath, "xpui", "xpui-modules.js"): jsModifiers,
		filepath.Join(appsFolderPath, "xpui", "xpui-snapshot.js"): {
			insertCustomApp,
		},
		filepath.Join(appsFolderPath, "xpui", "home-v2.js"): {
			insertHomeConfig,
		},
		filepath.Join(appsFolderPath, "xpui", "xpui-desktop-modals.js"): {
			insertVersionInfo,
		},
	}

	filesToModified[filepath.Join(appsFolderPath, "xpui", "xpui.js")] = append(filesToModified[filepath.Join(appsFolderPath, "xpui", "xpui.js")], insertCustomApp)
	filesToModified[filepath.Join(appsFolderPath, "xpui", "xpui.js")] = append(filesToModified[filepath.Join(appsFolderPath, "xpui", "xpui.js")], insertExpFeatures)
	// filesToModified[filepath.Join(appsFolderPath, "xpui", "vendor~xpui.js")] = []func(string, Flag){insertExpFeatures}

	for file, calls := range filesToModified {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}

		for _, call := range calls {
			call(file)
		}
	}

	copyFile(
		filepath.Join("jsPatches", "homeConfig.js"),
		filepath.Join(appsFolderPath, "xpui", "helper"))

	copyFile(
		filepath.Join("jsPatches", "spicetifyWrapper.js"),
		filepath.Join(appsFolderPath, "xpui", "helper"))

	copyFile(
		filepath.Join("jsPatches", "expFeatures.js"),
		filepath.Join(appsFolderPath, "xpui", "helper"))
}

func checkExistAndCreate(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		os.MkdirAll(dir, 0700)
	}
}

func copyFile(srcPath, dest string) error {
	fSrc, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer fSrc.Close()

	checkExistAndCreate(dest)
	destPath := filepath.Join(dest, filepath.Base(srcPath))
	fDest, err := os.OpenFile(
		destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer fDest.Close()

	_, err = io.Copy(fDest, fSrc)
	if err != nil {
		return err
	}

	return nil
}

func copy(src, dest string, recursive bool, filters []string) error {
	dir, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	os.MkdirAll(dest, 0700)

	for _, file := range dir {
		fileName := file.Name()
		fSrcPath := filepath.Join(src, fileName)

		fDestPath := filepath.Join(dest, fileName)
		if file.IsDir() && recursive {
			os.MkdirAll(fDestPath, 0700)
			if err = copy(fSrcPath, fDestPath, true, filters); err != nil {
				return err
			}
		} else {
			if len(filters) > 0 {
				isMatch := false

				for _, filter := range filters {
					if strings.Contains(fileName, filter) {
						isMatch = true
						break
					}
				}

				if !isMatch {
					continue
				}
			}

			fSrc, err := os.Open(fSrcPath)
			if err != nil {
				return err
			}
			defer fSrc.Close()

			fDest, err := os.OpenFile(
				fDestPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
			if err != nil {
				return err
			}
			defer fDest.Close()

			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func UserCSS(appsFolderPath, themeFolder string, scheme map[string]string) {
	colorsDest := filepath.Join(appsFolderPath, "xpui", "colors.css")
	if err := os.WriteFile(colorsDest, []byte(getColorCSS(scheme)), 0700); err != nil {
		log.Fatalln(err)
	}
	cssDest := filepath.Join(appsFolderPath, "xpui", "user.css")
	if err := os.WriteFile(cssDest, []byte(getUserCSS(themeFolder)), 0700); err != nil {
		log.Fatalln(err)
	}
}

func UserAsset(appsFolderPath, themeFolder string) {
	var assetsPath = getAssetsPath(themeFolder)
	var xpuiPath = filepath.Join(appsFolderPath, "xpui")
	if err := copy(assetsPath, xpuiPath, true, nil); err != nil {
		log.Fatalln(err)
	}
}

func htmlMod(htmlPath string) {
	extensionsHTML := "\n"
	helperHTML := "\n"

	extensionsHTML += "<script defer src='extensions/theme.js'></script>\n"
	helperHTML += "<script defer src='helper/homeConfig.js'></script>\n"
	helperHTML += "<script defer src='helper/expFeatures.js'></script>\n"

	var extList string
	for _, ext := range []string{} {
		extList += fmt.Sprintf(`"%s",`, ext)
	}

	var customAppList string
	for _, app := range []string{} {
		customAppList += fmt.Sprintf(`"%s",`, app)
	}

	helperHTML += fmt.Sprintf(`<script>
			Spicetify.Config={};
			Spicetify.Config["version"]="%s";
			Spicetify.Config["current_theme"]="%s";
			Spicetify.Config["color_scheme"]="%s";
			Spicetify.Config["extensions"] = [%s];
			Spicetify.Config["custom_apps"] = [%s];
			Spicetify.Config["check_spicetify_update"]=%v;
		</script>
		`, "0", "", "", extList, customAppList, "1")

	for _, v := range []string{} {
		if strings.HasSuffix(v, ".mjs") {
			extensionsHTML += fmt.Sprintf("<script defer type='module' src='extensions/%s'></script>\n", v)
		} else {
			extensionsHTML += fmt.Sprintf("<script defer src='extensions/%s'></script>\n", v)
		}
	}

	// for _, v := range []string{} {
	// manifest, _, err := getAppManifest(v)
	// if err == nil {
	// for _, extensionFile := range manifest.ExtensionFiles {
	// if strings.HasSuffix(extensionFile, ".mjs") {
	// extensionsHTML += fmt.Sprintf("<script defer type='module' src='extensions/%s/%s'></script>\n", v, extensionFile)
	// } else {
	// extensionsHTML += fmt.Sprintf("<script defer src='extensions/%s/%s'></script>\n", v, extensionFile)
	// }
	// }
	// }

	modifyFile(htmlPath, func(content string) string {
		replace(
			&content,
			`<script defer="defer" src="/xpui-snapshot\.js"></script>`,
			func(submatches ...string) string {
				return `<script defer="defer" src="/xpui-modules.js"></script><script defer="defer" src="/xpui-snapshot.js"></script>`
			})
		replace(
			&content,
			`<\!-- spicetify helpers -->`,
			func(submatches ...string) string {
				return fmt.Sprintf("%s%s", submatches[0], helperHTML)
			})
		replace(
			&content,
			`</body>`,
			func(submatches ...string) string {
				return fmt.Sprintf("%s%s", extensionsHTML, submatches[0])
			})
		return content
	})
}

func getUserCSS(themeFolder string) string {
	if len(themeFolder) == 0 {
		return ""
	}

	cssFilePath := filepath.Join(themeFolder, "user.css")
	_, err := os.Stat(cssFilePath)

	if err != nil {
		return ""
	}

	content, err := os.ReadFile(cssFilePath)
	if err != nil {
		return ""
	}

	return string(content)
}

var baseColorList = map[string]string{
	"text":               "ffffff",
	"subtext":            "b3b3b3",
	"main":               "121212",
	"main-elevated":      "242424",
	"highlight":          "1a1a1a",
	"highlight-elevated": "2a2a2a",
	"sidebar":            "000000",
	"player":             "181818",
	"card":               "282828",
	"shadow":             "000000",
	"selected-row":       "ffffff",
	"button":             "1db954",
	"button-active":      "1ed760",
	"button-disabled":    "535353",
	"tab-active":         "333333",
	"notification":       "4687d6",
	"notification-error": "e22134",
	"misc":               "7f7f7f",
}

func getColorCSS(scheme map[string]string) string {
	var variableList string
	var variableRGBList string
	mergedScheme := make(map[string]string)

	for k, v := range scheme {
		mergedScheme[k] = v
	}

	for k, v := range baseColorList {
		if len(mergedScheme[k]) == 0 {
			mergedScheme[k] = v
		}
	}

	for k, v := range mergedScheme {
		parsed := parseColor(v)
		variableList += fmt.Sprintf("    --spice-%s: #%s;\n", k, parsed.Hex())
		variableRGBList += fmt.Sprintf("    --spice-rgb-%s: %s;\n", k, parsed.RGB())
	}

	return fmt.Sprintf(":root {\n%s\n%s\n}\n", variableList, variableRGBList)
}

type color struct {
	red, green, blue int64
}

type Color interface {
	Hex() string
	RGB() string
	TerminalRGB() string
}

var xrdb map[string]string

func getXRDB() error {
	db := map[string]string{}

	if len(xrdb) > 0 {
		return nil
	}

	output, err := exec.Command("xrdb", "-query").Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	re := regexp.MustCompile(`^\*\.?(\w+?):\s*?#([A-Za-z0-9]+)`)
	for scanner.Scan() {
		line := scanner.Text()
		for _, match := range re.FindAllStringSubmatch(line, -1) {
			if match != nil {
				db[match[1]] = match[2]
			}
		}
	}

	xrdb = db

	return nil
}

func fromXResources(input string) string {
	queries := strings.Split(input, ":")
	if len(queries[1]) == 0 {
		os.Exit(0)
	}

	if err := getXRDB(); err != nil {
		log.Fatalln(err)
	}

	if len(xrdb) < 1 {
		os.Exit(0)
	}

	value, ok := xrdb[queries[1]]

	if !ok || len(value) == 0 {
		if len(queries) > 2 {
			value = queries[2]
		} else {
			os.Exit(0)
		}
	}

	return value
}

func parseColor(raw string) Color {
	var red, green, blue int64

	if strings.HasPrefix(raw, "${") {
		endIndex := len(raw) - 1
		raw = raw[2:endIndex]

		if strings.HasPrefix(raw, "xrdb:") {
			raw = fromXResources(raw)

		} else if env := os.Getenv(raw); len(env) > 0 {
			raw = env
		}
	}

	if strings.Contains(raw, ",") {
		list := strings.SplitN(raw, ",", 3)
		list = append(list, "255", "255")

		red = stringToInt(list[0], 10)
		green = stringToInt(list[1], 10)
		blue = stringToInt(list[2], 10)

	} else {
		re := regexp.MustCompile("[a-fA-F0-9]+")
		hex := re.FindString(raw)

		if len(hex) == 3 {
			expanded := []byte{
				hex[0], hex[0],
				hex[1], hex[1],
				hex[2], hex[2]}

			hex = string(expanded)
		}

		hex += "ffffff"

		red = stringToInt(hex[:2], 16)
		green = stringToInt(hex[2:4], 16)
		blue = stringToInt(hex[4:6], 16)
	}

	return color{red, green, blue}
}

func (c color) Hex() string {
	return fmt.Sprintf("%02x%02x%02x", c.red, c.green, c.blue)
}

func (c color) RGB() string {
	return fmt.Sprintf("%d,%d,%d", c.red, c.green, c.blue)
}

func (c color) TerminalRGB() string {
	return fmt.Sprintf("%d;%d;%d", c.red, c.green, c.blue)
}

func stringToInt(raw string, base int) int64 {
	value, err := strconv.ParseInt(raw, base, 0)
	if err != nil {
		value = 255
	}

	if value < 0 {
		value = 0
	}

	if value > 255 {
		value = 255
	}

	return value
}

func insertCustomApp(jsPath string) {
	modifyFile(jsPath, func(content string) string {
		reactPatterns := []string{
			`([\w_\$][\w_\$\d]*(?:\(\))?)\.lazy\(\((?:\(\)=>|function\(\)\{return )(\w+)\.(\w+)\(\d+\)\.then\(\w+\.bind\(\w+,\d+\)\)\}?\)\)`,
			`([\w_\$][\w_\$\d]*)\.lazy\(async\(\)=>\{(?:[^{}]|\{[^{}]*\})*await\s+(\w+)\.(\w+)\(\d+\)\.then\(\w+\.bind\(\w+,\d+\)\)`,
			`([\w_\$][\w_\$\d]*(?:\(\))?)\.lazy\(async\(\)=>await\s+Promise\.all\(\[[^\]]+\]\)\.then\((\w+)\.bind\((\w+),\d+\)\)`,
		}

		elementPatterns := []string{
			`(\([\w$\.,]+\))\(([\w\.]+),\{path:"/settings(?:/[\w\*]+)?",?(element|children)?`,
			`([\w_\$][\w_\$\d]*(?:\(\))?\.createElement|\([\w$\.,]+\))\(([\w\.]+),\{path:"\/collection"(?:,(element|children)?[:.\w,{}()$/*"]+)?\}`,
		}

		reactSymbs, matchedReactPattern := findSymbolWithPattern(
			"Custom app React symbols",
			content,
			reactPatterns)
		eleSymbs, matchedElementPattern := findSymbolWithPattern(
			"Custom app React Element",
			content,
			elementPatterns)

		if (len(reactSymbs) < 2) || (len(eleSymbs) == 0) {
			return content
		}

		appMap := ""
		appReactMap := ""
		appEleMap := ""
		cssEnableMap := ""
		appNameArray := ""

		wildcard := ""
		if eleSymbs[2] == "" {
			eleSymbs[2] = "children"
		} else if eleSymbs[2] == "element" {
			wildcard = "*"
		}

		for index, app := range []string{} {
			appName := `spicetify-routes-` + app
			appMap += fmt.Sprintf(`"%s":"%s",`, appName, appName)
			appNameArray += fmt.Sprintf(`"%s",`, app)

			appReactMap += fmt.Sprintf(
				`,spicetifyApp%d=%s.lazy((()=>%s.%s("%s").then(%s.bind(%s,"%s"))))`,
				index, reactSymbs[0], reactSymbs[1], reactSymbs[2],
				appName, reactSymbs[1], reactSymbs[1], appName)

			appEleMap += fmt.Sprintf(
				`%s(%s,{path:"/%s/%s",pathV6:"/%s/*",%s:%s(spicetifyApp%d,{})}),`,
				eleSymbs[0], eleSymbs[1], app, wildcard, app, eleSymbs[2], eleSymbs[0], index)

			cssEnableMap += fmt.Sprintf(`,"%s":1`, appName)
		}

		replace(
			&content,
			`\{(\d+:"xpui)`,
			func(submatches ...string) string {
				return fmt.Sprintf("{%s%s", appMap, submatches[1])
			})

		matchedReactPattern = seekToCloseParen(
			content,
			matchedReactPattern,
			'(',
			')',
		)

		content = strings.Replace(
			content,
			matchedReactPattern,
			fmt.Sprintf("%s%s", matchedReactPattern, appReactMap),
			1,
		)

		replaceOnce(
			&content,
			matchedElementPattern,
			func(submatches ...string) string {
				return fmt.Sprintf("%s%s", appEleMap, submatches[0])
			})

		content = insertNavLink(content, appNameArray)

		replaceOnce(
			&content,
			`\d+:1,\d+:1,\d+:1`,
			func(submatches ...string) string {
				return fmt.Sprintf("%s%s", submatches[0], cssEnableMap)
			})

		return content
	})
}

func findSymbolWithPattern(debugInfo, content string, clues []string) ([]string, string) {
	for _, v := range clues {
		re := regexp.MustCompile(v)
		found := re.FindStringSubmatch(content)
		if found != nil {
			return found[1:], v
		}
	}

	return nil, ""
}

func insertNavLink(str string, appNameArray string) string {
	libraryXItemMatch := seekToCloseParen(
		str,
		`\("li",\{[^\{]*\{[^\{]*\{to:"\/search`,
		'(', ')')

	if libraryXItemMatch != "" {
		str = strings.Replace(
			str,
			libraryXItemMatch,
			fmt.Sprintf("%s,Spicetify._renderNavLinks([%s], false)", libraryXItemMatch, appNameArray),
			1)
	}

	replaceOnceWithPriority(&str,
		[]string{
			// Global Navbar <= 1.2.45
			`(,[a-zA-Z_\$][\w\$]*===(?:[a-zA-Z_\$][\w\$]*\.){2}HOME_NEXT_TO_NAVIGATION&&.+?)\]`,
			// Global Navbar >= 1.2.60, greedy matching with enclosing brackets
			`("global-nav-bar".*[[\w\$&|]*\(0,[a-zA-Z_\$][\w\$]*\.jsx\)\(\s*\w+,\s*\{\s*className:\w*\s*\}\s*\))\]`,
			// Global Navbar >= 1.2.46, lazy matching
			`("global-nav-bar".*?)(\(0,\s*[a-zA-Z_\$][\w\$]*\.jsx\))(\(\s*\w+,\s*\{\s*className:\w*\s*\}\s*\))`,
		},
		func(index int, submatches ...string) string {
			switch index {
			case 0, 1:
				return fmt.Sprintf("%s,Spicetify._renderNavLinks([%s], true)]", submatches[1], appNameArray)
			case 2:
				return fmt.Sprintf("%s[%s%s,Spicetify._renderNavLinks([%s], true)].flat()", submatches[1], submatches[2], submatches[3], appNameArray)
			}
			return ""
		},
	)

	return str
}

func replaceOnceWithPriority(str *string, patterns []string, repl func(index int, submatches ...string) string) {
	for i, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		firstMatch := true
		*str = re.ReplaceAllStringFunc(*str, func(match string) string {
			if firstMatch {
				firstMatch = false
				submatches := re.FindStringSubmatch(match)
				if submatches != nil {
					return repl(i, submatches...)
				}
			}
			return match
		})
		if !firstMatch {
			break
		}
	}
}

func insertHomeConfig(jsPath string) {
	modifyFile(jsPath, func(content string) string {
		replaceOnce(
			&content,
			`(createDesktopHomeFeatureActivationShelfEventFactory.*?)([\w\.]+)(\.map)`,
			func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetifyHomeConfig.arrange(%s)%s", submatches[1], submatches[2], submatches[3])
			})

		// >= 1.2.40
		replaceOnce(
			&content,
			`(&&"HomeShortsSectionData".*?[\],}])([a-zA-Z])(\}\)?\()`,
			func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetifyHomeConfig.arrange(%s)%s", submatches[1], submatches[2], submatches[3])
			})

		return content
	})
}

func getAssetsPath(themeFolder string) string {
	dir := filepath.Join(themeFolder, "assets")

	if _, err := os.Stat(dir); err != nil {
		return ""
	}

	return dir
}

func insertExpFeatures(jsPath string) {
	modifyFile(jsPath, func(content string) string {
		replaceOnce(
			&content,
			`(function \w+\((\w+)\)\{)(\w+ \w+=\w\.name;if\("internal")`,
			func(submatches ...string) string {
				return fmt.Sprintf("%s%s=Spicetify.expFeatureOverride(%s);%s", submatches[1], submatches[2], submatches[2], submatches[3])
			})

		// utils.ReplaceOnce(
		// 	&content,
		// 	`(\w+\.fromJSON)(\s*=\s*function\b[^{]*{[^}]*})`,
		// 	func(submatches ...string) string {
		// 		return fmt.Sprintf("%s=Spicetify.createInternalMap%s", submatches[1], submatches[2])
		// 	})

		replaceOnce(
			&content,
			`(([\w$.]+\.fromJSON)\(\w+\)+;)(return ?[\w{}().,]+[\w$]+\.Provider,)(\{value:\{localConfiguration)`,
			func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetify.createInternalMap=%s;%sSpicetify.RemoteConfigResolver=%s", submatches[1], submatches[2], submatches[3], submatches[4])
			})

		return content
	})
}

func insertVersionInfo(jsPath string) {
	modifyFile(jsPath, func(content string) string {
		replaceOnce(
			&content,
			`(\w+(?:\(\))?\.createElement|\([\w$\.,]+\))\([\w\."]+,[\w{}():,]+\.containerVersion\}?\),`,
			func(submatches ...string) string {
				return fmt.Sprintf(`%s%s("details",{children: [
					%s("summary",{children: "Spicetify v" + Spicetify.Config.version}),
					%s("li",{children: "Theme: " + Spicetify.Config.current_theme + (Spicetify.Config.color_scheme && " / ") + Spicetify.Config.color_scheme}),
					%s("li",{children: "Extensions: " + Spicetify.Config.extensions.join(", ")}),
					%s("li",{children: "Custom apps: " + Spicetify.Config.custom_apps.join(", ")}),
					]}),`, submatches[0], submatches[1], submatches[1], submatches[1], submatches[1], submatches[1])
			})
		return content
	})
}
