function pal_abbr
    set -l completions_file ~/.local/share/pal_helper/completions.txt
    cat $completions_file
end

abbr --add pal1 --function pal_abbr

