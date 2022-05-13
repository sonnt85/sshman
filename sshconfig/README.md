# sshconfig

This is a Go parser for `sshconfig` files. Importantly, this parser attempts
to preserve comments in a given file, so you can manipulate a `sshconfig` file
from a program, if your heart desires.

It's designed to be used with the excellent
[x/crypto/ssh](https://golang.org/x/crypto/ssh) package, which handles SSH
negotiation but isn't very easy to configure.

The `sshconfig` `Get()` and `GetStrict()` functions will attempt to read values
from `$HOME/.ssh/config` and fall back to `/etc/ssh/sshconfig`. The first
argument is the host name to match on, and the second argument is the key you
want to retrieve.

```go
port := sshconfig.Get("myhost", "Port")
```

You can also load a config file and read values from it.

```go
var config = `
Host *.test
  Compression yes
`

cfg, err := sshconfig.Decode(strings.NewReader(config))
fmt.Println(cfg.Get("example.test", "Port"))
```

Some SSH arguments have default values - for example, the default value for
`KeyboardAuthentication` is `"yes"`. If you call Get(), and no value for the
given Host/keyword pair exists in the config, we'll return a default for the
keyword if one exists.

### Manipulating SSH config files

Here's how you can manipulate an SSH config file, and then write it back to
disk.

```go
f, _ := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))
cfg, _ := sshconfig.Decode(f)
for _, host := range cfg.Hosts {
    fmt.Println("patterns:", host.Patterns)
    for _, node := range host.Nodes {
        // Manipulate the nodes as you see fit, or use a type switch to
        // distinguish between Empty, KV, and Include nodes.
        fmt.Println(node.String())
    }
}

// Print the config to stdout:
fmt.Println(cfg.String())
```

## Spec compliance

Wherever possible we try to implement the specification as documented in
the `sshconfig` manpage. Unimplemented features should be present in the
[issues][issues] list.

Notably, the `Match` directive is currently unsupported.

[issues]: https://github.com/kevinburke/sshconfig/issues

## Errata

This is the second [comment-preserving configuration parser][blog] I've written, after
[an /etc/hosts parser][hostsfile]. Eventually, I will write one for every Linux
file format.

[blog]: https://kev.inburke.com/kevin/more-comment-preserving-configuration-parsers/
[hostsfile]: https://github.com/kevinburke/hostsfile

## Donating

Donations free up time to make improvements to the library, and respond to
bug reports. You can send donations via Paypal's "Send Money" feature to
kev@inburke.com. Donations are not tax deductible in the USA.
