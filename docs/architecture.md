Pal is basically a Cobra based CLI app that sends user input to an LLM provider, as configured via some yaml, and does something™️ with the response.

All data stored by pal goes into plain text files. Performance in this regard is generally not a concern, because the majority of time in most operations is spent waiting on LLM responses.

A typical invocation of pal includes a slash command (/command) and a message from the user. Pal accepts an arbitrary amount of "positional arguments" as user input, relieving the user from needing to worry about quoting their input. That said, quotation marks, apostrophies, and other symbols can be awkward in this mode.

To achieve the flexible input mode mentioned above, pal pre-parses its arguments before presenting them to Cobra. This way, pal can support flags both before and after slash commands, while also accepting any inputs afterwards. That is to say, user messages can freely include slash commands and flag forms, as long as they don't occur at the beginning in a way that makes things ambiguious. Pal also has a default command that runs when the slash command is ommitted, slightly complicating things further.

Therefore, a bit of unusual CLI arg parsing is one thing that makes pal special. At the same time, pal also accepts input from stdin, which opens further possibilities for providing inputs.
