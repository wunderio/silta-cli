## silta secrets decrypt

Decrypt encrypted files

```
silta secrets decrypt [flags]
```

### Options

```
      --files string            Encrypted file location. Can have multiple, comma separated paths (i.e. 'silta/secrets.enc,silta/secrets2.enc')
  -h, --help                    help for decrypt
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

