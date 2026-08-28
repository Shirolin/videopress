# Videopress 视觉设计契约 — 压片台

## 目的

Videopress 是 Windows 桌面视频压缩工具。界面隐喻为**暖色暗金属压片台**：文件在此排队、压制、出片。不是 SaaS 毛玻璃仪表盘，也不是切割垫工作台。

## 基调

- 不透明金属面板，密实控件条，密集队列
- 单一品牌强调色：**钨丝暖金**（tungsten）
- 功能色仅用于状态：完成（ok）、取消/破坏（danger）
- 中文可读：Noto Sans SC / Noto Serif SC

## 色彩令牌

```css
--desk            #161410    /* 暖暗金属底，窗口背景 */
--desk-hem          #100e0b    /* 标题栏 / 面板 */
--desk-inset        #1c1914    /* 凹槽、输入框 */
--rule              #2c2820    /* 发丝线 */
--rule-strong       #3c382e
--ink               #e8e4d8    /* 暖米白正文 */
--ink-dim           #8a8578
--ink-faint         #5e5a50
--tungsten          #c4a574    /* 唯一品牌强调：主 CTA、进度、字标勾、焦点 */
--tungsten-hover    #d2b484
--tungsten-press    #a88855
--tungsten-ink      #161410    /* 钨色按钮上的文字（深色，非白） */
--ok                #7a9a6a    /* 任务完成等功能成功 */
--danger            #8a3a32    /* 取消 / 破坏性悬停 */
--radius            5px        /* 最大 6px */
```

禁止：`#08080a` 冷黑底、紫/品红渐变、多品牌强调色、钨色按钮白字。

## 字体

| 用途 | 字体 |
|------|------|
| 字标「Videopress」+ 静「压片」 | Noto Serif SC 600 |
| UI 正文与控件 | IBM Plex Sans + Noto Sans SC |
| 规格 / CRF / 码率 | IBM Plex Mono |

禁止作为 `font-family` 出现：Inter、Roboto、Arial、system-ui、Geist、Outfit、Satoshi、Nunito。

## 圆角

默认 `--radius`（5px）。局部最大 6px。禁止 `rounded-2xl` 类默认大圆角。

## 面板

`.hem-panel`：不透明 `--desk-hem` 背景，`1px solid var(--rule)` 边框，`border-radius: var(--radius)`。无 `backdrop-filter`、无光晕、无渐变边框。

## 签名动效（唯一）

队列项变为完成时：行 `translateY(1px)` 按压感 + 钨色进度条填满，`180–220ms cubic-bezier(0.16, 1, 0.3, 1)`。空态拖放区落定 `8px`。其余交互保持安静。`prefers-reduced-motion: reduce` 时跳过。

## 保留（产品性格）

- 拖放区 + 文件队列 + 预设/并发/硬解/复制音频/强制/跳过
- 高级编码/CRF/FPS + 设置 + i18n + GPU 检测
- 队列有项时拖放区紧凑化 + 统计条
- 控件条密度
- 真实拖放（已实现则保留文案）
- 取消红、完成绿（使用 `--danger` / `--ok` 去饱和）

## 禁止（装饰性 slop）

- 毛玻璃 `backdrop-filter` 于 chrome 面板
- `body::before/::after` 紫/品红光球
- 紫品红渐变 CTA / 进度条
- `.neon-hover` 紫色光晕
- Sparkles / Rocket / PartyPopper 类图标
- 英文大写 tracking 装饰
- 营销话术：高颜值 / 高速 / 智能 / 专业级 / 闪电 / 极速
- 白字钨色按钮
- 额外装饰微交互

## 文案

界面文案陈述功能，不推销。i18n 键名与语义不变；仅调整过于营销的中英字串。
