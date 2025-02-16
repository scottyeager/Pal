package abbr

import (
	_ "embed"
)

//go:embed abbr.zsh
var ZshAbbrEmbed string

func GetZshAbbrScript(abbreviationPrefix string) string {
	return `local pal_prefix="` + abbreviationPrefix + `"` + "\n" + ZshAbbrEmbed + "\n"
}
