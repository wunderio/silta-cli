## silta hub mcp

Connect AI clients (Claude Code, opencode, Cursor, VS Code) to the Silta MCP server

### Synopsis

Show how to connect AI clients to the Silta dashboard MCP server, which gives
them read-only access to the releases, pods, logs, metrics and traces you can
see in the dashboard.

The AI client logs in by itself: on first use it opens a browser where you sign
in to the dashboard and approve the client. No 'silta hub login' is needed.

Without --client, setup instructions for every supported client are printed.
With --client claude, the server is registered in Claude Code directly when the
'claude' command is installed.

Use --url to print only the MCP endpoint URL.

```
silta hub mcp [flags]
```

### Options

```
      --client string    Only set up this client: claude, opencode, cursor or vscode (claude registers the server directly)
  -h, --help             help for mcp
      --hub-url string   Silta hub URL (defaults to the configured hub.url)
      --name string      Server name to use in the client configuration (default "silta")
      --scope string     Claude Code configuration scope: local, user or project (default "user")
      --url              Print only the MCP endpoint URL
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta hub](silta_hub.md)	 - Interact with a Silta hub

