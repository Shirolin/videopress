# Videopress 协作约定

## 语言与提交
- 回复、计划、报告用简体中文；变量名、库名、命令、路径保持原文。
- 提交格式 `<type>(<scope>): <subject>`，type 取 feat|fix|docs|style|refactor|perf|test|chore，subject 用中文，scope 填受影响模块（如 `gui`、`engine`、`cli`）。

## 验证
- 统一入口 `pwsh -File scripts/make.ps1 <test|vet|build|ci>`；装了 make 时 `make <target>` 等价。
- 测试用替身注入：`engine.Dependencies`、`gif.Dependencies` 的函数字段就是替身入口，测试不碰本机 FFmpeg。

## 文档
- 改压缩调度、并发、取消、进度或启动分发前，先读 `docs/ARCHITECTURE.md`。
- 改界面配色、字号、间距或组件选型前，先读 `docs/DESIGN.md`。
