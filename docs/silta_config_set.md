## silta config set

Set configuration

### Synopsis

This will set configuration information. The first argument is the key and the second argument is the value.
If the key already exists, the value will be overwritten. Supports nested keys using dot notation.
Usage: silta config set <key> <value>
Example: silta config set mykey
Example: silta config set mykey myvalue
Example: silta config set mykey.subkey myvalue


```
silta config set [flags]
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta config](silta_config.md)	 - Silta configuration commands

