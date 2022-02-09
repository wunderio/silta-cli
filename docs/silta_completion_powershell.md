## silta completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	silta completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
silta completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta completion](silta_completion.md)	 - Generate the autocompletion script for the specified shell

