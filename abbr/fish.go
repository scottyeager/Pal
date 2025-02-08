package abbr

import (
	_ "embed"
)

//go:embed abbr.fish
var FishAbbrEmbed string

func GetFishAbbrScript(abbreviationPrefix string) string {
	return `set -l pal_prefix "` + abbreviationPrefix + `"` + "\n" + FishAbbrEmbed + "\n"
}
