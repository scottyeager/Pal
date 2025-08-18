# File containing the command lines to expand to
local PAL_ABBR_FILE=${PAL_ABBR_FILE:-~/.local/share/pal_helper/expansions.txt}
# Default prefix value if not set
local pal_prefix=${pal_prefix:-pal}

# Widget function to expand prefix+digit
pal-expand-abbr() {
    # Get the current line buffer
    local buffer=$BUFFER
    local prefix_length=${#pal_prefix}

    # Check if buffer starts with prefix and has digits
    if [[ $buffer =~ ^${pal_prefix}[0-9]+ ]]; then
        # Get the digits
        local digits=${buffer[$((prefix_length+1)),-1]}

        # Handle prefix0 specially (first line)
        if [[ $digits == "0" ]]; then
            local line=$(head -n1 $PAL_ABBR_FILE 2>/dev/null)
            if [[ -n $line ]]; then
                BUFFER=$line
                zle end-of-line
                zle self-insert
                return
            fi
        fi

        # Handle multi-digit expansion
        local -a lines
        for ((i=0; i<${#digits}; i++)); do
            # Skip first line (prefix0) for regular digits
            local line_num=$((${digits:$i:1} + 1))
            local line=$(sed -n "${line_num}p" $PAL_ABBR_FILE 2>/dev/null)
            if [[ -n $line ]]; then
                lines+=("$line")
            fi
        done

        if [[ ${#lines} -gt 0 ]]; then
            # Join lines with newlines
            BUFFER=${(F)lines}
            # For some reason we end up on the second line when expanding three
            # lines, so we need to move to the end of the buffer rather than
            # just the end of the line
            CURSOR=${#BUFFER}
        fi
        zle self-insert

    # If zsh-abbr is installed, defer to its expansion
    elif [[ -n $widgets[abbr-expand-and-insert] ]]; then
        zle abbr-expand-and-insert

    # Fallback to regular space
    else
        zle self-insert
    fi
}

# Create the widget
zle -N pal-expand-abbr

# Bind space key to our widget
bindkey ' ' pal-expand-abbr

# Ctrl-space always makes a space character
bindkey "^ " magic-space
