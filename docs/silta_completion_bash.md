## silta completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(silta completion bash)

To load completions for every new session, execute once:

#### Linux:

	silta completion bash > /etc/bash_completion.d/silta

#### macOS:

	silta completion bash > $(brew --prefix)/etc/bash_completion.d/silta

You will need to start a new shell for this setup to take effect.


```
silta completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta completion](silta_completion.md)	 - Generate the autocompletion script for the specified shell

