# mweb-export

[![Build Release](https://github.com/rapirent/mweb-export/actions/workflows/Release.yml/badge.svg)](https://github.com/rapirent/mweb-export/actions/workflows/Release.yml)

## In Short

一個幫助你將mweb類型(由sqlite管理)的markdown note，轉出為以資料夾為類別分類的markdown note
由 [TBXark的mweb-export](https://github.com/TBXark/mweb-export) 修改而來，原repo的功能是生成mweb類型筆記的樹狀文檔，我做了一些小修改

## Install

### Go

```shell
go install github.com/rapirent/mweb-export@latest
```

## Usage

mweb-export [--path mweb筆記位置] [--target 要轉出位置]

by default, path and target is `current directory`

## Special Thanks

原作者TBXark
