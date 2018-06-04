SHELL := /bin/bash
.DEFAULT_GOAL := run
.PHONY: run

run:
	 go run main.go \
	 		filechangeinfo.go  \
			filenode.go \
			filetree.go \
			tar_read.go \
			filetreeview.go \
			layerview.go
