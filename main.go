package main

import (
	"os"

	"gerrit.instructure.com/muss/cmd"
	_ "gerrit.instructure.com/muss/cmd/config"
	"gerrit.instructure.com/muss/proc"
)

func main() {
	proc.EnableExec()
	os.Exit(cmd.Execute(os.Args[1:]))
}
