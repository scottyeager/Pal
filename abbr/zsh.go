package abbr

import (
	_ "embed"
	"fmt"
	"github.com/scottyeager/pal/inout"
)

//go:embed abbr.zsh
var ZshAbbrEmbed string

func GetZshAbbrScript(abbreviationPrefix string) string {
	path, err := inout.GetAbbrFilePath()
	if err != nil {
		// Fallback to Linux default if something goes wrong
		path = fmt.Sprintf("%s", "~/.local/share/pal_helper/expansions.txt")
	}
	return `local pal_prefix="` + abbreviationPrefix + `"` + "\n" +
		`export PAL_ABBR_FILE="` + path + `"` + "\n" +
		ZshAbbrEmbed + "\n"
}
