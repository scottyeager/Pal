# In normal usage, we prepend a line to set the prefix. For testing, it's
# helpful to have a default here
if not set -q pal_prefix
    set -g pal_prefix pal
end

function _pal_get_completion
    set completions_file ~/.local/share/pal_helper/completions.txt
    set suffix (string match -r "$pal_prefix(\d+)" $argv[1] | tail -n1)
    
    # Handle prefix0 specially
    if test "$suffix" = "0"
        # Return just the first line (prefix0 content)
        set completion (head -n1 $completions_file)
        echo $completion
        return
    end
    
    # Handle regular completions
    set line_numbers (string split "" $suffix)
    mkdir -p (dirname $completions_file)
    touch $completions_file
    set -f completion ""
    for line_number in $line_numbers
        # Skip the first line (prefix0) when processing regular completions
        set adjusted_line (math $line_number + 1)
        set new_line (sed -n {$adjusted_line}p $completions_file)
        if not test $line_number = $line_numbers[-1]
            set completion {$completion}{$new_line}\n
        else
            set completion {$completion}{$new_line}
        end
    end
    echo $completion
end

abbr --add pal_complete --regex "$pal_prefix(\d+)" --function _pal_get_completion
