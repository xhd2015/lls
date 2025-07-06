# lls
lls helps developers to quickly navigate through working projects.

When invoked, lls will find projects from configured locations, and pass to `fzf` to filter based on your selection, then open the project with `code` (VSCode).

It streamlines the operation opening and switching projects inside VSCode and terminal.

# Installation
```sh
go install github.com/xhd2015/lls@latest
```

# Usage
```sh
lls
```

Edit your projects:
```sh
lls edit
```

Example config:
```sh
{
    "envs": [
        "X",
        "W",
        "GOPATH"
    ],
    "projects": [
        "$X/xgo",
        "$W/company-stuff
    ]
}
```

Explanation:
- `$X` points to your personal projects directory
- `$W` points to the projects at work
- `GOPATH` points to your `GOPATH` env

You can add arbitrary more.