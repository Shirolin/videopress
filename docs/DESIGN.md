# Videopress 视觉契约

## 权威来源

配色、字号、间距、圆角、动效时长全部来自 MujicaUI：

- `core.Tokens(c)` —— 背景、表面、文字、边框、强调色、功能色（成功/警告/危险）
- `core.Density(c)` —— `Unit()` 与 `ControlHeight()`
- `theme` —— `SerifFont`、`H1Size`/`H2Size`/`H3Size`/`BodySize`/`CaptionSize`、`CardRadius`、`ControlRadius`、`BorderWidth`、`StateDuration`

本仓库不定义颜色常量、不自建色板、不写内联 hex。需要颜色时从 `core.Tokens` 取，需要间距时乘 `core.Density(c).Unit()`。

跟随系统明暗（`core.Settings{Mode: core.System}`），默认风为酒红/象牙/衬线。

## 组件

用 MujicaUI 的现成组件，不要自绘替代品：

| 位置 | 组件 |
|------|------|
| 标题栏 | `layout.TitleBar` |
| 页面滚动 | `layout.ScrollArea` |
| 顶栏导航、预设、档位、并发数 | `navigation.SegmentedControl` |
| 按钮 | `input.Button` / `input.IconButton` |
| 开关、勾选、数值、滑块、下拉 | `input.Switch` / `Checkbox` / `NumberInput` / `Slider` / `Select` |
| 卡片、折叠、列表 | `data.Card` / `data.Collapsible` |
| 进度、状态、空态、toast | `feedback.Progress` / `ShowToast` / `Toaster` |
| 配置页骨架 | `account.SettingsLayout` / `SettingsSection` / `SettingsRow` |
| 图标 | `icons` 包，缺失名字用 `ic()` 兜底 |

MujicaUI 的组件只随包发布约 90 个 Lucide 图标，`icons.Must` 对未收录的名字会 panic；取图标一律走 [`internal/gui/view.go`](../internal/gui/view.go) 的 `ic()`。

## 文案

- 界面文案陈述功能，不推销。
- 中英文都从 [`internal/gui/i18n.go`](../internal/gui/i18n.go) 的 `copyTable` 取，不要写死在视图里。
- 短标签可直接用库名/格式名（`H.265 (HEVC)`、`GIF`），不翻译。

## 禁止（装饰性 slop）

- 自建色板或内联 hex
- 毛玻璃、光晕、渐变边框、渐变 CTA
- 紫/品红装饰光球
- Sparkles / Rocket / PartyPopper 类图标
- 英文大写 tracking 装饰
- 营销话术：高颜值 / 高速 / 智能 / 专业级 / 闪电 / 极速
- 额外装饰微交互；动效只用于状态变化本身
