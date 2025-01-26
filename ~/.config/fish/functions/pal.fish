function pal
    if test (count $argv) -eq 1
        if not string match -q '/*' -- $argv[1]
            set -l completions_file ~/.local/share/pal_helper/completions.txt
            if test -f $completions_file
                set -l completion (cat $completions_file)
                commandline -r $completion
                return
            end
        end
    end
    
    command pal $argv
end
