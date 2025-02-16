# File containing the command lines to expand to
local PAL_ABBR_FILE=~/.local/share/pal_helper/completions.txt

# Widget function to expand prefix+digit
pal-expand-abbr() {
    # Get the current line buffer
    local buffer=$BUFFER
    local prefix_length=${#pal_prefix}

    # Check if buffer starts with prefix and ends with a digit
    if [[ $buffer[1,$prefix_length] == $pal_prefix && $buffer[-1] =~ ^[0-9]$ ]]; then
        # Get the digit (0-9)
        local digit=$buffer[-1]

        # Read the corresponding line from file (1-based index)
        local line=$(sed -n "$((digit + 1))p" $PAL_ABBR_FILE 2>/dev/null)

        if [[ -n $line ]]; then
            # Replace prefix+digit with the line
            BUFFER=$line
        fi
    fi

    # Always add a space after expansion
    zle self-insert
}

# Create the widget
zle -N pal-expand-abbr

# Bind space key to our widget
bindkey ' ' pal-expand-abbr
