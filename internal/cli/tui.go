package app

import (
	"fmt"
	"io"
	"os"
	"time"
)

type menuOption struct {
	name        string
	description string
}

func showInteractiveMenu(w io.Writer) string {
	// 启用 Windows 虚拟终端处理
	_ = enableVirtualTerminal()

	options := []menuOption{
		{name: "small", description: "小档案 (33% 比例 - 480px 限制)"},
		{name: "standard", description: "中等   (50% 比例 - 720px 限制) [默认]"},
		{name: "quality", description: "大档案 (100% 比例 - 无限制)"},
	}

	selectedIndex := 1 // 默认选中中等
	countdown := 5
	userInteracted := false

	drawMenu(w, options, selectedIndex, countdown, userInteracted, true)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	inputChan := make(chan int, 10)
	doneChan := make(chan bool)

	// 启动后台协程非阻塞监听键盘事件
	go func() {
		for {
			select {
			case <-doneChan:
				return
			default:
				if hasKey() {
					key := readKey()
					inputChan <- key
				}
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			if !userInteracted {
				countdown--
				if countdown <= 0 {
					close(doneChan)
					fmt.Fprint(w, "\n\n已自动确认默认选项：中等\n\n")
					return "standard"
				}
				drawMenu(w, options, selectedIndex, countdown, userInteracted, false)
			}
		case key := <-inputChan:
			if !userInteracted {
				userInteracted = true
				countdown = 0
			}

			if key == 224 || key == 0 {
				// Windows 虚拟键前缀，需要再读一字节
				time.Sleep(10 * time.Millisecond)
				if hasKey() {
					nextKey := readKey()
					if nextKey == 72 { // 方向键 上
						selectedIndex = (selectedIndex - 1 + len(options)) % len(options)
					} else if nextKey == 80 { // 方向键 下
						selectedIndex = (selectedIndex + 1) % len(options)
					}
				}
			} else if key == 13 || key == 10 { // 回车键
				close(doneChan)
				fmt.Fprintf(w, "\n\n已选择并确认：%s\n\n", options[selectedIndex].description)
				return options[selectedIndex].name
			} else if key == 49 { // 按键 '1'
				selectedIndex = 0
			} else if key == 50 { // 按键 '2'
				selectedIndex = 1
			} else if key == 51 { // 按键 '3'
				selectedIndex = 2
			}

			drawMenu(w, options, selectedIndex, countdown, userInteracted, false)
		}
	}
}

func drawMenu(w io.Writer, options []menuOption, selectedIndex int, countdown int, userInteracted bool, isFirst bool) {
	if !isFirst {
		// 回退 7 行，清除之前的渲染重新绘制
		fmt.Fprint(w, "\033[7A")
	}

	fmt.Fprintln(w, "请选择压缩规格：")
	for i, opt := range options {
		if i == selectedIndex {
			fmt.Fprintf(w, " %s\n", green(fmt.Sprintf("> %s", opt.description)))
		} else {
			fmt.Fprintf(w, "   %s\n", gray(opt.description))
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "【提示】使用上下方向键 %s 切换，%s 确认。按任意键停止倒计时。\n", cyan("↑↓"), cyan("回车"))

	if !userInteracted {
		fmt.Fprintf(w, "倒计时: %s 秒...   \r", yellow(fmt.Sprintf("%d", countdown)))
	} else {
		fmt.Fprintf(w, "%s               \r", green("已启用手动选择。"))
	}
}

func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isTerminalFd(f.Fd())
	}
	return false
}
