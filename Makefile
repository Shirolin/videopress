# Videopress 构建入口。逻辑在 scripts/make.ps1；Windows 可直接运行：
#   pwsh -File scripts/make.ps1 test
.PHONY: help frontend-deps frontend-dist wails-bindings frontend-check test vet build-go build ci

PS1 ?= pwsh -NoProfile -File scripts/make.ps1

help:
	$(PS1) help

frontend-deps:
	$(PS1) frontend-deps

frontend-dist:
	$(PS1) frontend-dist

wails-bindings:
	$(PS1) wails-bindings

frontend-check:
	$(PS1) frontend-check

test:
	$(PS1) test

vet:
	$(PS1) vet

build-go:
	$(PS1) build-go

build:
	$(PS1) build

ci:
	$(PS1) ci
