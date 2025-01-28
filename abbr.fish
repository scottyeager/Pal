function _pal_get_completion
    set -l completions_file ~/.local/share/pal_helper/completions.txt
    set -l line_number (string match -r "$pal_prefix(\d+)" $argv[1] | tail -n1)
    set -l total_lines (wc -l < $completions_file)
    if test $line_number -gt $total_lines
        tail -n1 $completions_file
    else
        sed -n {$line_number}p $completions_file
    end
end

abbr --add pal_complete --regex "$pal_prefix(\d+)" --function _pal_get_completion
