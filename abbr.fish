function _pal_get_completion
    set -l completions_file ~/.local/share/pal_helper/completions.txt
    set -l line_number (string match -r "$pal_prefix(\d+)" $argv[1] | tail -n1)
    mkdir -p (dirname $completions_file)
    touch $completions_file
    sed -n {$line_number}p $completions_file
end

abbr --add pal_complete --regex "$pal_prefix(\d+)" --function _pal_get_completion
