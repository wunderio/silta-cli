## silta secrets encrypt

Encrypt secret files

```
silta secrets encrypt [flags]
```

### Options

```
      --file string             Decrypted file location. Can have multiple, comma separated paths (i.e. 'silta/secrets.enc,silta/secrets2.enc')
  -h, --help                    help for encrypt
      --output-file string      Output file location (optional, rewrites original when undefined, don't use with multiple input files)
      --secret-key string       Secret key (falls back to SECRET_KEY environment variable. Also see: --secret-key-env)
      --secret-key-env string   Environment variable holding symmetrical decryption key.
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta secrets](silta_secrets.md)	 - Manage encrypted secret files

