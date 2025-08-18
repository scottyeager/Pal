package abbr

import (
	_ "embed"
	"fmt"
	"github.com/scottyeager/pal/inout"
)

//go:embed abbr.fish
var FishAbbrEmbed string

func GetFishAbbrScript(abbreviationPrefix string) string {
	path, err := inout.GetAbbrFilePath()
	if err != nil {
		// Fallback to Linux default if something goes wrong
		path = fmt.Sprintf("%s", "~/.local/share/pal_helper/expansions.txt")
	}
	return `set -g pal_prefix "` + abbreviationPrefix + `"` + "\n" +
		`set -g pal_abbr_file "` + path + `"` + "\n" +
		FishAbbrEmbed + "\n"
}
