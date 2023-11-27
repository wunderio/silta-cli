## silta completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(silta completion zsh)

To load completions for every new session, execute once:

#### Linux:

	silta completion zsh > "${fpath[1]}/_silta"

#### macOS:

	silta completion zsh > $(brew --prefix)/share/zsh/site-functions/_silta

You will need to start a new shell for this setup to take effect.


```
silta completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta completion](silta_completion.md)	 - Generate the autocompletion script for the specified shell

