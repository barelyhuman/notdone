# notdone

Tiny grouped checklists ( as a daemon )

> **Note**: Only works on \*nix systems

# TOC

- [Installation](#installation)
- [Configuration](#configuration)
- [Daemon Control](#daemon-control)
  - [Reload config](#reload-config)
  - [Stop / Terminate notdone](#stop--terminate-notdone)
- [FAQ](#faq)
- [License](/LICENSE)

## Installation

1. Download a release for your system from the [releases &rarr;](/releases) page.

2. Place the binary wherever you wish to, it'll be easier to access if you place it somewhere accessible by your system

```sh
$ mkdir -p ~/.local/bin

$ install notdone ~/.local/bin

# modifiy `PATH` to include the above path
$ echo "export PATH=$HOME/.local/bin:$PATH" >> ~/.bashrc
# or if you use zsh
$ echo "export PATH=$HOME/.local/bin:$PATH" >> ~/.zshrc

# reload your shell's environment
source ~/.bashrc
# or
. ~/.zshrc
```

3. Start the daemon by running `notdone`

```sh
$ notdone
```

3. You should now be able to access notdone in your browser at `localhost:3000`

## Configuration

> **Note**: Configuration for `notdone` is stored in `~/.notdone/config.json`

| key    | description                         | default |
| ------ | ----------------------------------- | ------- |
| `port` | The port at which `notdone` runs on | `3000`  |

## Daemon Control

### Reload config

```sh
$ notdone -s reload
```

### Stop / Terminate notdone

```sh
$ notdone -s quit
# or
$ notdone -s stop
```

## FAQ

#### Why daemon?

- Faster that a desktop apps
- Easier to cross compile (at least when not targeting Windows)

#### Where's my data!!?

It's all stored in `~/.notdone/storage.json`, you can parse and modify the `json` with any other software if you wish to. Write your own GUI for it, no biggies

## License

[MIT](/LICENSE)
