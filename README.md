# go-findql

A "find" command on steroids. Created in an *1h 12m*!

## Why?

I don't even know. Sometimes I like to challenge myself to doing things like this. I love "find", but sometimes I find it difficult to remember all the options and parameters. So I wanted to create something simpler ("find" is still better...)

I'm not currently working on Go, so wanted to try out modules and all that stuff. And wanted to do it quick. I no longer have the time to code for fun.

So, I decided to create this library, and wanted a maximum time spent of an hour and a half. So I'm writing this README, trying to fill the remainder of my time.

## Is it production ready?

No. And it's slow, for sure.

## Usage

```bash
$ go-findql -path=vendor -depth=3 -filter="directory=false and size > 600 and name like 'err%' and modified_at > '2019-01-01 00:00:00'"
```

## How did you do it?

Using `sqlite` (in memory), and that's about it.

## Does it work on Windows?

I thought of creating build tags for different platforms... But no. Who uses Windows? 😄.

Enjoy!
