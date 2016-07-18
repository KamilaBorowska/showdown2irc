# showdown2irc

[![Build Status](https://travis-ci.org/xfix/showdown2irc.svg?branch=master)](https://travis-ci.org/xfix/showdown2irc)
[![Code Climate](https://codeclimate.com/github/xfix/showdown2irc/badges/gpa.svg)](https://codeclimate.com/github/xfix/showdown2irc)

This program lets you talk on Showdown servers (currently only supports
Showdown main) using an IRC client.

It's quite unfinished, but if you are interested in using this program,
first [set up Go prefix](https://golang.org/doc/code.html#GOPATH), and then
use the following command to download the package.

```sh
go get github.com/xfix/showdown2irc
```

Then you can use `showdown2irc` command in terminal to start a server.
Once started, to use a server, connect using IRC client to `localhost`,
port `6667`. Set your real name (not nickname, the program is using
real name because Showdown nicks can contain spaces) to your Showdown
username, and server password to your Showdown account password.

Once connected, you can join a room by using `/join` command, with room
name starting with `#` (as room names on IRC do), and without spaces.
For example, to join the room Tech & Code, type `/join #tech&code`.
