## silta completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	silta completion fish | source

To load completions for every new session, execute once:

	silta completion fish > ~/.config/fish/completions/silta.fish

You will need to start a new shell for this setup to take effect.


```
silta completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta completion](silta_completion.md)	 - Generate the autocompletion script for the specified shell

