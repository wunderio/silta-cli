## silta hub login

Log in to a Silta hub

### Synopsis

Authenticate the Silta CLI against a Silta hub.

By default a browser window is opened to approve the login. On headless
machines use --device to complete the login using a short code instead.

The Silta hub URL can be provided with --hub-url or stored in the
configuration under 'hub.url'.

```
silta hub login [flags]
```

### Options

```
      --device           Use device code flow instead of opening a browser
  -h, --help             help for login
      --hub-url string   Silta hub URL (overrides config 'hub.url'; the value is saved to config)
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta hub](silta_hub.md)	 - Interact with a Silta hub

