# Introduction

`xp` is a tool created to make practising extreme programming easier.

## Reference

Full list of options supported:

```
➜  ~ xp
NAME:
   xp - extreme programming made simple

USAGE:
   xp [global options] command [command options] [arguments...]

VERSION:
   0.2.1

COMMANDS:
     show-config, sc  Print the current config
     add-info         Add xp info to the COMMIT msg file
     dev, d           Dev management
     repo, r          Repo management
     help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value  set the default configuration file (default: "~/.xp")
   --help, -h      show help
   --version, -v   print the version
```

## Features

- Manage the co-authorship of commits by automatically writing appropriate* `Co-authored-by` trailers (see [link](https://help.github.com/articles/creating-a-commit-with-multiple-authors/) for details on this standard)
- Take co-authorship information written in the first line of the commit message and convert that into appropriate `Co-authored-by` trailers (overrides all other sources)
- Ensure that the author drafting the commit is not duplicated as a `Co-authored-by` trailer
- Preserve co-authorship information when ammending commits

## Installation

The simplest way to install `xp` in your dev environment is:

```
go get -u github.com/kidoman/xp
```

`brew` will be added as an option at a later date.

## Usage

`xp` stores its global configuration at `~/.xp` (can be overriden via global flag `--config`)

Most of the `xp` functionality are exposted via various subcommands:

- `show-config` (`sc`): Print the current stored configuration
- `dev` (`d`): Add/remove developers in xp
- `repo` (`r`): Add/remove repos managed by xp

A separate command `add-info` is made available for use from within `git` hooks:
- prepare-commit-msg
- commit-msg

## Example

Suppose we have a repo at `~/work/lambda` which we want to now manage using `xp` (this assumes you have already installed `xp` using the instructions above):


Add Karan Misra &lt;kidoman@gmail.com&gt; as a tracked author in the system with shortcode "km" to allow for easy referencing in future command line invocations or the first line of commit messages. Same for "akshat":

```
$ xp dev add km "Karan Misra" kidoman@beef.com
$ xp dev add ak "akshat" akshat@beef.com
```

Switch to the directory with the `git` repo:

```
$ cd ~/work/lambda
```

Initialize the git hooks and register the repo with `xp`:

```
$ xp repo init .
```

Indicate that `akshat` is pairing with you by adding him using his shortcode:

```
$ xp repo dev ak
```

Commit as normal:

```
$ touch CHANGE
$ git add .
$ git commit -m"Added CHANGE"
```

Rejoice at a well formed commit message:

```
$ git log -1
commit c4700d32046d94070de0c160eb35b2090973b507 (HEAD -> master)
Author: Karan Misra <kidoman@beef.com>
Date:   Thu Mar 7 03:04:25 2019 +0530

    Added CHANGE

    Co-authored-by: akshat <akshat@beef.com>
```

## Bonus

If you quickly want to author a commit with someone you typically don't pair with:

```
$ xp dev add as "Anand Shankar" anand@beef.com
```

After making the required changes:

```
$ git add .
$ git commit -m"[as] Make world better"
```

The commit message becomes:

```
$ git log -1
commit d4710d32046d94070de0c160eb35b2091973b507 (HEAD -> master)
Author: Karan Misra <kidoman@beef.com>
Date:   Thu Mar 7 03:12:21 2019 +0530

    Make world better

    Co-authored-by: Anand Shankar <anand@beef.com>
```

Note: See how the `[as]` from the start of the commit message has now resulted in `Anand Shankar` being added as a co-author, overriding the repo level setting (thus `akshat` is not in the list anymore.)
