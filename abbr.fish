function _pal_get_completion
    set -l completions_file ~/.local/share/pal_helper/completions.txt
    set -l line_number $argv[1]
    sed -n {$line_number}p $completions_file
end

abbr --regex 'pal(\d+)' --function _pal_get_completion

