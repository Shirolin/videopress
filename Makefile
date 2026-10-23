# Videopress 构建入口。逻辑在 scripts/make.ps1；Windows 可直接运行：
#   pwsh -File scripts/make.ps1 test
.PHONY: help test vet build-go dev build ci

PS1 ?= pwsh -NoProfile -File scripts/make.ps1

help:
	$(PS1) help

test:
	$(PS1) test

vet:
	$(PS1) vet

build-go:
	$(PS1) build-go

dev:
	$(PS1) dev

build:
	$(PS1) build

ci:
	$(PS1) ci
