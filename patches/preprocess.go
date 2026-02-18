package patches

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Patch struct {
	Name        string
	Regex       string
	Replacement func(submatches ...string) string
	Once        bool
}

func additionalPatches(input string) string {
	graphQLPatches := []Patch{
		{
			Name:  "GraphQL definitions (<=1.2.30)",
			Regex: `((?:\w+ ?)?[\w$]+=)(\{kind:"Document",definitions:\[\{(?:\w+:[\w"]+,)+name:\{(?:\w+:[\w"]+,?)+value:("\w+"))`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetify.GraphQL.Definitions[%s]=%s", submatches[1], submatches[3], submatches[2])
			},
		},
		{
			Name:  "GraphQL definitions (>=1.2.31)",
			Regex: `(=new [\w_\$][\w_\$\d]*\.[\w_\$][\w_\$\d]*\("(\w+)","(query|mutation)","[\w\d]{64}",null\))`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf(`=Spicetify.GraphQL.Definitions["%s"]%s`, submatches[2], submatches[1])
			},
		},
	}

	return applyPatches(input, graphQLPatches)
}

func findMatch(input string, regexpTerm string) [][]string {
	re := regexp.MustCompile(regexpTerm)
	matches := re.FindAllStringSubmatch(input, -1)
	return matches
}

func findFirstMatch(input string, regexpTerm string) []string {
	matches := findMatch(input, regexpTerm)
	if len(matches) > 0 {
		return matches[0]
	}
	return nil
}

func findLastMatch(input string, regexpTerm string) []string {
	matches := findMatch(input, regexpTerm)
	if len(matches) > 0 {
		return matches[len(matches)-1]
	}
	return nil
}

func exposeAPIs_main(input string) string {
	inputContextMenu := findFirstMatch(input, `.*(?:value:"contextmenu"|"[^"]*":"context-menu")`)
	if len(inputContextMenu) > 0 {
		croppedInput := inputContextMenu[0]
		react := findLastMatch(croppedInput, `([a-zA-Z_\$][\w\$]*)\.useRef`)[1]
		candicates := findLastMatch(croppedInput, `\(\{[^}]*menu:([a-zA-Z_\$][\w\$]*),[^}]*trigger:([a-zA-Z_\$][\w\$]*),[^}]*triggerRef:([a-zA-Z_\$][\w\$]*)`)
		oldCandicates := findLastMatch(croppedInput, `([a-zA-Z_\$][\w\$]*)=[\w_$]+\.menu[^}]*,([a-zA-Z_\$][\w\$]*)=[\w_$]+\.trigger[^}]*,([a-zA-Z_\$][\w\$]*)=[\w_$]+\.triggerRef`)
		var menu, trigger, target string
		if len(oldCandicates) != 0 {
			menu = oldCandicates[1]
			trigger = oldCandicates[2]
			target = oldCandicates[3]
		} else if len(candicates) != 0 {
			menu = candicates[1]
			trigger = candicates[2]
			target = candicates[3]
		} else {
			menu = "e.menu"
			trigger = "e.trigger"
			target = "e.triggerRef"
		}

		replace(&input, `\(0,([\w_$]+)\.jsx\)\((?:[\w_$]+\.[\w_$]+,\{value:"contextmenu"[^}]+\}\)\}\)|"[\w-]+",\{[^}]+:"context-menu"[^}]+\}\))`, func(submatches ...string) string {
			return fmt.Sprintf("(0,%s.jsx)((Spicetify.ContextMenuV2._context||(Spicetify.ContextMenuV2._context=%s.createContext(null))).Provider,{value:{props:%s?.props,trigger:%s,target:%s},children:%s})", submatches[1], react, menu, trigger, target, submatches[0])
		})
	}

	xpuiPatches := []Patch{
		{
			Name:  "showNotification",
			Regex: `(?:\w+ |,)([\w$]+)=(\([\w$]+=[\w$]+\.dispatch)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf(`;globalThis.Spicetify.showNotification=(message,isError=false,msTimeout)=>%s({message,feedbackType:isError?"ERROR":"NOTICE",msTimeout});const %s=%s`, submatches[1], submatches[1], submatches[2])
			},
		},
		{
			Name:  "Remove list of exclusive shows",
			Regex: `\["spotify:show.+?\]`,
			Replacement: func(submatches ...string) string {
				return "[]"
			},
		},
		{
			Name:  "Remove Star Wars easter eggs",
			Regex: `\w+\(\)\.createElement\(\w+,\{onChange:this\.handleSaberStateChange\}\),`,
			Replacement: func(submatches ...string) string {
				return ""
			},
		},
		{
			Name:  "Remove data-testid",
			Regex: `"data-testid":`,
			Replacement: func(submatches ...string) string {
				return `"":`
			},
		},
		{
			Name:  "Expose PlatformAPI",
			Regex: `((?:setTitlebarHeight|registerFactory)[\w(){}<>:.,&$!=;""?!#% ]+)(\{version:[a-zA-Z_\$][\w\$]*,)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetify._platform=%s", submatches[1], submatches[2])
			},
		},
		{
			Name:  "Redux store",
			Regex: `(,[\w$]+=)(([$\w,.:=;(){}]+\(\{session:[\w$]+,features:[\w$]+,seoExperiment:[\w$]+\}))`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetify.Platform.ReduxStore=%s", submatches[1], submatches[2])
			},
		},
		{
			Name:  "React Component: Platform Provider",
			Regex: `(,[$\w]+=)((function\([\w$]{1}\)\{var [\w$]+=[\w$]+\.platform,[\w$]+=[\w$]+\.children,)|(\(\{platform:[\w$]+,children:[\w$]+\}\)=>\{))`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetify.ReactComponent.PlatformProvider=%s", submatches[1], submatches[2])
			},
		},
		{
			Name:  "Prevent breaking popupLyrics",
			Regex: `document.pictureInPictureElement&&\(\w+.current=[!\w]+,document\.exitPictureInPicture\(\)\),\w+\.current=null`,
			Replacement: func(submatches ...string) string {
				return ""
			},
		},
		{
			Name:  "Spotify Custom Snackbar Interfaces (<=1.2.37)",
			Regex: `\b\w\s*\(\)\s*[^;,]*enqueueCustomSnackbar:\s*(\w)\s*[^;]*;`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetify.Snackbar.enqueueCustomSnackbar=%s;", submatches[0], submatches[1])
			},
		},
		{
			Name:  "Spotify Custom Snackbar Interfaces (>=1.2.38)",
			Regex: `(=)[^=]*\(\)\.enqueueCustomSnackbar;`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("=Spicetify.Snackbar.enqueueCustomSnackbar%s;", submatches[0])
			},
		},
		{
			Name:  "Spotify Image Snackbar Interface",
			Regex: `(=)(\(\({[^}]*,\s*imageSrc)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetify.Snackbar.enqueueImageSnackbar=%s", submatches[1], submatches[2])
			},
		},
		{
			Name:  "React Component: Navigation for navLinks",
			Regex: `(;const [\w\d]+=)((?:\(0,[\w\d]+\.memo\))[\(\d,\w\.\){:}=]+\=[\d\w]+\.[\d\w]+\.getLocaleForURLPath\(\))`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%sSpicetify.ReactComponent.Navigation=%s", submatches[1], submatches[2])
			},
			Once: true,
		},
		{
			Name:  "Context Menu V2",
			Regex: `("Menu".+?children:)([\w$][\w$\d]*)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%s[Spicetify.ContextMenuV2.renderItems(),%s].flat()", submatches[1], submatches[2])
			},
		},
	}

	return applyPatches(input, xpuiPatches)
}

func seekToCloseParen(content string, regexpTerm string, leftChar, rightChar byte) string {
	loc := regexp.MustCompile(regexpTerm).FindStringIndex(content)
	if len(loc) > 0 {
		start := loc[0]
		end := start
		count := 0
		init := false

		for {
			switch content[end] {
			case leftChar:
				count += 1
				init = true
			case rightChar:
				count -= 1
			}
			end += 1
			if count == 0 && init {
				break
			}
		}
		return content[start:end]
	}
	return ""
}

func exposeAPIs_vendor(input string) string {
	// URI
	replace(
		&input,
		`,(\w+)\.prototype\.toAppType`,
		func(submatches ...string) string {
			return fmt.Sprintf(`,(globalThis.Spicetify.URI=%s)%s`, submatches[1], submatches[0])
		})
	vendorPatches := []Patch{
		{
			Name:  "Spicetify.URI",
			Regex: `,(\w+)\.prototype\.toAppType`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf(`,(globalThis.Spicetify.URI=%s)%s`, submatches[1], submatches[0])
			},
		},
		{
			Name:  "Map styled-components classes",
			Regex: `(\w+ [\w$_]+)=[\w$_]+\([\w$_]+>>>0\)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%s=Spicetify._getStyledClassName(arguments,this)", submatches[1])
			},
		},
		{
			Name:  "Tippy.js",
			Regex: `([\w\$_]+)\.setDefaultProps=`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("Spicetify.Tippy=%s;%s", submatches[1], submatches[0])
			},
		},
		{
			Name:  "Flipper components",
			Regex: `([\w$]+)=((?:function|\()([\w$.,{}()= ]+(?:springConfig|overshootClamping)){2})`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("%s=Spicetify.ReactFlipToolkit.spring=%s", submatches[1], submatches[2])
			},
		},
		{
			// https://github.com/iamhosseindhv/notistack
			Name:  "Snackbar",
			Regex: `\w+\s*=\s*\w\.call\(this,[^)]+\)\s*\|\|\s*this\)\.enqueueSnackbar`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("Spicetify.Snackbar=%s", submatches[0])
			},
		},
	}

	// URI after 1.2.4
	if !strings.Contains(input, "Spicetify.URI") {
		URIObj := regexp.MustCompile(`(?:class ([\w$_]+)\{constructor|([\w$_]+)=function\(\)\{function ?[\w$_]+)\([\w$.,={}]+\)\{[\w !?:=.,>&(){}[\];]*this\.hasBase62Id`).FindStringSubmatch(input)

		if len(URIObj) != 0 {
			URI := seekToCloseParen(
				input,
				`\{(?:constructor|function ?[\w$_]+)\([\w$.,={}]+\)\{[\w !?:=.,>&(){}[\];]*this\.hasBase62Id`,
				'{', '}')

			if URIObj[1] == "" {
				URIObj[1] = URIObj[2]
				// Class is a self-invoking function
				URI = fmt.Sprintf("%s()", URI)
			}

			input = strings.Replace(
				input,
				URI,
				fmt.Sprintf("%s;Spicetify.URI=%s;", URI, URIObj[1]),
				1)
		}
	}

	return applyPatches(input, vendorPatches)
}

func applyPatches(input string, patches []Patch) string {
	for _, patch := range patches {
		if patch.Once {
			replaceOnce(&input, patch.Regex, patch.Replacement)
		} else {
			replace(&input, patch.Regex, patch.Replacement)
		}
	}
	return input
}

func readRemoteCssMap(tag string, cssTranslationMap *map[string]string) error {
	var cssMapURL string = "https://raw.githubusercontent.com/spicetify/cli/" + tag + "/css-map.json"
	cssMapResp, err := http.Get(cssMapURL)
	if err != nil {
		return err
	} else {
		err := json.NewDecoder(cssMapResp.Body).Decode(cssTranslationMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func StartMinimal(extractedAppsPath string) {
	appPath := filepath.Join(extractedAppsPath, "xpui")

	var cssTranslationMap = make(map[string]string)
	readRemoteCssMap("latest", &cssTranslationMap)

	var filesToPatch []string
	filepath.Walk(appPath, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())
		if ext == ".js" || ext == ".css" || ext == ".html" {
			filesToPatch = append(filesToPatch, p)
		}
		return nil
	})

	for _, p := range filesToPatch {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		fileName := info.Name()
		extension := filepath.Ext(fileName)

		switch extension {
		case ".js":
			modifyFile(p, func(content string) string {
				switch fileName {
				case "xpui-modules.js", "xpui-snapshot.js":
					content = exposeAPIs_main(content)
					content = exposeAPIs_vendor(content)
				case "xpui.js":
					content = exposeAPIs_main(content)
					content = exposeAPIs_vendor(content)
				case "vendor~xpui.js":
					content = exposeAPIs_vendor(content)
				}

				replaceOnce(&content, `(typeName\])`, func(submatches ...string) string {
					return fmt.Sprintf(`%s || []`, submatches[1])
				})

				// if spotifyMajor >= 1 && spotifyMinor >= 2 && spotifyPatch < 78 {
				// utils.ReplaceOnce(&content, `\(\({[^}]*,\s*imageSrc`, func(submatches ...string) string {
				// return fmt.Sprintf("Spicetify.Snackbar.enqueueImageSnackbar=%s", submatches[0])
				// })
				// }

				content = additionalPatches(content)
				if fileName == "dwp-top-bar.js" || fileName == "dwp-now-playing-bar.js" || fileName == "dwp-home-chips-row.js" {
					replaceOnce(&content, `e\.state\.cinemaState`, func(submatches ...string) string {
						return "e.state?.cinemaState"
					})
				}
				for k, v := range cssTranslationMap {
					replace(&content, k, func(submatches ...string) string { return v })
				}
				content = colorVariableReplaceForJS(content)
				return content
			})

		case ".css":
			modifyFile(p, func(content string) string {
				for k, v := range cssTranslationMap {
					replace(&content, k, func(submatches ...string) string { return v })
				}
				if fileName == "xpui.css" || fileName == "xpui-snapshot.css" {
					content += `
.main-gridContainer-fixedWidth{grid-template-columns:repeat(auto-fill,var(--column-width));}.main-cardImage-imageWrapper{background-color:var(--card-color,#333);border-radius:6px;box-shadow:0 8px 24px rgba(0,0,0,.5);padding-bottom:100%;position:relative;width:100%;}.main-cardImage-image,.main-card-imagePlaceholder{height:100%;left:0;position:absolute;top:0;width:100%};.main-content-view{height:100%;}
`
				}
				return content
			})

		case ".html":
			modifyFile(p, func(content string) string {
				tags := "<link rel='stylesheet' class='userCSS' href='colors.css'>\n"
				tags += "<link rel='stylesheet' class='userCSS' href='user.css'>\n"
				tags += "<script src='helper/spicetifyWrapper.js'></script>\n"
				tags += "<!-- spicetify helpers -->\n"
				replace(&content, `<body(\sclass="[^"]*")?>`, func(submatches ...string) string {
					return fmt.Sprintf("%s\n%s", submatches[0], tags)
				})
				return content
			})
		}
	}
}

func colorVariableReplaceForJS(content string) string {
	colorVariablePatches := []Patch{
		{
			Name:  "CSS (JS): --spice-button",
			Regex: `"#1db954"`,
			Replacement: func(submatches ...string) string {
				return ` getComputedStyle(document.body).getPropertyValue("--spice-button").trim()`
			},
		},
		{
			Name:  "CSS (JS): --spice-subtext",
			Regex: `"#b3b3b3"`,
			Replacement: func(submatches ...string) string {
				return ` getComputedStyle(document.body).getPropertyValue("--spice-subtext").trim()`
			},
		},
		{
			Name:  "CSS (JS): --spice-text",
			Regex: `"#ffffff"`,
			Replacement: func(submatches ...string) string {
				return ` getComputedStyle(document.body).getPropertyValue("--spice-text").trim()`
			},
		},
		{
			Name:  "CSS (JS): --spice-text white",
			Regex: `color:"white"`,
			Replacement: func(submatches ...string) string {
				return `color:"var(--spice-text)"`
			},
		},
	}

	return applyPatches(content, colorVariablePatches)
}

func colorVariableReplace(content string) string {
	colorPatches := []Patch{
		{
			Name:  "CSS: --spice-player",
			Regex: `#(181818|212121)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-player)"
			},
		},
		{
			Name:  "CSS: --spice-card",
			Regex: `#282828\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-card)"
			},
		},
		{
			Name:  "CSS: --spice-main-elevated",
			Regex: `#(242424|1f1f1f)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-main-elevated)"
			},
		},
		{
			Name:  "CSS: --spice-main",
			Regex: `#121212\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-main)"
			},
		},
		{
			Name:  "CSS: --spice-card-elevated",
			Regex: `#(242424|1f1f1f)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-card-elevated)"
			},
		},
		{
			Name:  "CSS: --spice-highlight",
			Regex: `#1a1a1a\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-highlight)"
			},
		},
		{
			Name:  "CSS: --spice-highlight-elevated",
			Regex: `#2a2a2a\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-highlight-elevated)"
			},
		},
		{
			Name:  "CSS: --spice-sidebar",
			Regex: `#(000|000000)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-sidebar)"
			},
		},
		{
			Name:  "CSS: --spice-text",
			Regex: `(white;|#fff|#ffffff|#f8f8f8)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-text)"
			},
		},
		{
			Name:  "CSS: --spice-subtext",
			Regex: `#(b3b3b3|a7a7a7)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-subtext)"
			},
		},
		{
			Name:  "CSS: --spice-button",
			Regex: `#(1db954|1877f2)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-button)"
			},
		},
		{
			Name:  "CSS: --spice-button-active",
			Regex: `#(1ed760|1fdf64|169c46)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-button-active)"
			},
		},
		{
			Name:  "CSS: --spice-button-disabled",
			Regex: `#535353\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-button-disabled)"
			},
		},
		{
			Name:  "CSS: --spice-tab-active",
			Regex: `#(333|333333)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-tab-active)"
			},
		},
		{
			Name:  "CSS: --spice-misc",
			Regex: `#7f7f7f\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-misc)"
			},
		},
		{
			Name:  "CSS: --spice-notification",
			Regex: `#(4687d6|2e77d0)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-notification)"
			},
		},
		{
			Name:  "CSS: --spice-notification-error",
			Regex: `#(e22134|cd1a2b)\b`,
			Replacement: func(submatches ...string) string {
				return "var(--spice-notification-error)"
			},
		},
		{
			Name:  "CSS (rgba): --spice-main",
			Regex: `rgba\(18,18,18,([\d\.]+)\)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("rgba(var(--spice-main),%s)", submatches[1])
			},
		},
		{
			Name:  "CSS (rgba): --spice-card",
			Regex: `rgba\(40,40,40,([\d\.]+)\)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("rgba(var(--spice-card),%s)", submatches[1])
			},
		},
		{
			Name:  "CSS (rgba): --spice-rgb-shadow",
			Regex: `rgba\(0,0,0,([\d\.]+)\)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("rgba(var(--spice-rgb-shadow),%s)", submatches[1])
			},
		},
		{
			Name:  "CSS (hsla): --spice-rgb-text",
			Regex: `hsla\(0,0%,100%,\.9\)`,
			Replacement: func(submatches ...string) string {
				return "rgba(var(--spice-rgb-text),.9)"
			},
		},
		{
			Name:  "CSS (hsla): --spice-rgb-selected-row",
			Regex: `hsla\(0,0%,100%,([\d\.]+)\)`,
			Replacement: func(submatches ...string) string {
				return fmt.Sprintf("rgba(var(--spice-rgb-selected-row),%s)", submatches[1])
			},
		},
	}

	return applyPatches(content, colorPatches)
}

func parseInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func modifyFile(path string, repl func(string) string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return
	}

	content := repl(string(raw))

	os.WriteFile(path, []byte(content), 0700)
}

func replace(str *string, pattern string, repl func(submatches ...string) string) {
	re := regexp.MustCompile(pattern)
	*str = re.ReplaceAllStringFunc(*str, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		return repl(submatches...)
	})
}

func replaceOnce(str *string, pattern string, repl func(submatches ...string) string) {
	re := regexp.MustCompile(pattern)
	firstMatch := true
	*str = re.ReplaceAllStringFunc(*str, func(match string) string {
		if firstMatch {
			firstMatch = false
			submatches := re.FindStringSubmatch(match)
			if submatches != nil {
				return repl(submatches...)
			}
		}
		return match
	})
}
