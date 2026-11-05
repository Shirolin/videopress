package sendto

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
	"golang.org/x/sys/windows/registry"
)

const launcherName = "快速压缩视频.lnk"
const appLnkName = "Videopress.lnk"
const startMenuSubDir = "Videopress"

type CreateLnkFunc func(lnkPath string, targetPath string, arguments string) error
type RemoveFunc func(name string) error

// lnkTargetReader 读取快捷方式的真实目标路径接口，默认实现为 readLnkTargetPath。
var lnkTargetReader = readLnkTargetPath

// lnkCreator 创建快捷方式实现，默认实现为 createLnk。
var lnkCreator = createLnk

func Install(executablePath string) (string, error) {
	sendToDir, err := sendToDir()
	if err != nil {
		return "", err
	}
	return InstallAt(sendToDir, executablePath, createLnk)
}

func Uninstall() error {
	sendToDir, err := sendToDir()
	if err != nil {
		return err
	}
	return UninstallAt(sendToDir, os.Remove)
}

func InstallAt(sendToDir string, executablePath string, createLnk CreateLnkFunc) (string, error) {
	launcherPath := filepath.Join(sendToDir, launcherName)
	if err := createLnk(launcherPath, executablePath, "--sendto"); err != nil {
		return "", err
	}
	return launcherPath, nil
}

func UninstallAt(sendToDir string, remove RemoveFunc) error {
	lnkPath := filepath.Join(sendToDir, launcherName)
	_ = remove(lnkPath)

	// 兼容老版本清除
	oldCmdPath := filepath.Join(sendToDir, "快速压缩视频.cmd")
	_ = remove(oldCmdPath)

	return nil
}

// InstallDesktop 创建桌面快捷方式
func InstallDesktop(executablePath string) error {
	desktopDir, err := getDesktopDir()
	if err != nil {
		return err
	}
	lnkPath := filepath.Join(desktopDir, appLnkName)
	return createLnk(lnkPath, executablePath, "")
}

// UninstallDesktop 卸载桌面快捷方式
func UninstallDesktop() error {
	desktopDir, err := getDesktopDir()
	if err != nil {
		return err
	}
	lnkPath := filepath.Join(desktopDir, appLnkName)
	if _, err := os.Stat(lnkPath); err == nil {
		return os.Remove(lnkPath)
	}
	return nil
}

// InstallStartMenu 创建开始菜单快捷方式
func InstallStartMenu(executablePath string) error {
	startMenuDir, err := getStartMenuProgramsDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(startMenuDir, startMenuSubDir)
	_ = os.MkdirAll(appDir, 0755)

	lnkPath := filepath.Join(appDir, appLnkName)
	return createLnk(lnkPath, executablePath, "")
}

// UninstallStartMenu 卸载开始菜单快捷方式
func UninstallStartMenu() error {
	startMenuDir, err := getStartMenuProgramsDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(startMenuDir, startMenuSubDir)
	lnkPath := filepath.Join(appDir, appLnkName)
	if _, err := os.Stat(lnkPath); err == nil {
		if err := os.Remove(lnkPath); err != nil {
			return err
		}
	}
	// 仅删除空目录；若用户曾向该目录放入其他文件则静默保留
	_ = os.Remove(appDir)
	return nil
}

// RegisterContextMenu 注册右键直接压缩菜单 (注册表 HKCU，免管理员)
func RegisterContextMenu(executablePath string, lang string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("创建右键注册表项失败: %w", err)
	}
	defer k.Close()

	menuText := "使用 Videopress 压缩"
	if lang == "en" {
		menuText = "Compress with Videopress"
	}

	if err := k.SetStringValue("MUIVerb", menuText); err != nil {
		return fmt.Errorf("设置 MUIVerb 失败: %w", err)
	}
	if err := k.SetStringValue("Icon", executablePath); err != nil {
		return fmt.Errorf("设置 Icon 失败: %w", err)
	}

	cmdK, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress\command`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("创建 Command 子项失败: %w", err)
	}
	defer cmdK.Close()

	formattedCmd := fmt.Sprintf(`"%s" "%%1"`, executablePath)
	if err := cmdK.SetStringValue("", formattedCmd); err != nil {
		return fmt.Errorf("设置默认值失败: %w", err)
	}
	return nil
}

// UnregisterContextMenu 卸载右键菜单
func UnregisterContextMenu() error {
	// 先删除子项 command
	_ = registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress\command`)
	// 再删除主项
	err := registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress`)
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("删除右键菜单注册表项失败: %w", err)
	}
	return nil
}

func createLnk(lnkPath string, targetPath string, arguments string) error {
	initialized := false
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err == nil {
		initialized = true
	} else {
		if hr, ok := err.(*ole.OleError); !ok || (hr.Code() != 1 && hr.Code() != 0x80010106) {
			return fmt.Errorf("初始化 COM: %w", err)
		}
	}
	if initialized {
		defer ole.CoUninitialize()
	}

	sl, err := ole.CreateInstance(clsidShellLink, iidShellLinkW)
	if err != nil {
		return fmt.Errorf("创建 ShellLink: %w", err)
	}
	defer sl.Release()

	slThis := uintptr(unsafe.Pointer(sl))
	slVt := (*shellLinkVtbl)(unsafe.Pointer(sl.RawVTable))

	targetPtr, err := syscall.UTF16PtrFromString(targetPath)
	if err != nil {
		return err
	}
	if r, _, _ := syscall.SyscallN(slVt.setPath, slThis, uintptr(unsafe.Pointer(targetPtr))); int32(r) < 0 {
		return fmt.Errorf("设置快捷方式目标失败: 0x%08x", uint32(r))
	}

	if arguments != "" {
		argPtr, err := syscall.UTF16PtrFromString(arguments)
		if err != nil {
			return err
		}
		if r, _, _ := syscall.SyscallN(slVt.setArguments, slThis, uintptr(unsafe.Pointer(argPtr))); int32(r) < 0 {
			return fmt.Errorf("设置快捷方式参数失败: 0x%08x", uint32(r))
		}
	}

	var pf *ole.IUnknown
	r, _, _ := syscall.SyscallN(slVt.queryInterface, slThis, uintptr(unsafe.Pointer(iidPersistFile)), uintptr(unsafe.Pointer(&pf)))
	if int32(r) < 0 {
		return fmt.Errorf("QueryInterface IPersistFile: 0x%08x", uint32(r))
	}
	defer pf.Release()

	pfThis := uintptr(unsafe.Pointer(pf))
	pfVt := (*persistFileVtbl)(unsafe.Pointer(pf.RawVTable))

	lnkPtr, err := syscall.UTF16PtrFromString(lnkPath)
	if err != nil {
		return err
	}
	if r, _, _ := syscall.SyscallN(pfVt.save, pfThis, uintptr(unsafe.Pointer(lnkPtr)), 1); int32(r) < 0 {
		return fmt.Errorf("保存快捷方式失败: 0x%08x", uint32(r))
	}
	return nil
}

func sendToDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("未设置 APPDATA 环境变量")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "SendTo"), nil
}

func getDesktopDir() (string, error) {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return "", fmt.Errorf("未设置 USERPROFILE 环境变量")
	}
	return filepath.Join(userProfile, "Desktop"), nil
}

func getStartMenuProgramsDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("未设置 APPDATA 环境变量")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs"), nil
}

func sendToLnkPath() (string, error) {
	dir, err := sendToDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, launcherName), nil
}

func desktopLnkPath() (string, error) {
	dir, err := getDesktopDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appLnkName), nil
}

func startMenuLnkPath() (string, error) {
	dir, err := getStartMenuProgramsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, startMenuSubDir, appLnkName), nil
}

// IsSendToInstalled 检查 SendTo 快捷方式是否存在且目标有效
func IsSendToInstalled() bool {
	p, err := sendToLnkPath()
	if err != nil {
		return false
	}
	return isLnkTargetValid(p)
}

// IsDesktopInstalled 检查桌面快捷方式是否存在且目标有效
func IsDesktopInstalled() bool {
	p, err := desktopLnkPath()
	if err != nil {
		return false
	}
	return isLnkTargetValid(p)
}

// IsStartMenuInstalled 检查开始菜单快捷方式是否存在且目标有效
func IsStartMenuInstalled() bool {
	p, err := startMenuLnkPath()
	if err != nil {
		return false
	}
	return isLnkTargetValid(p)
}

func isLnkTargetValid(lnkPath string) bool {
	if _, err := os.Stat(lnkPath); err != nil {
		return false
	}
	target, err := lnkTargetReader(lnkPath)
	if err != nil || target == "" {
		return false
	}
	_, err = os.Stat(target)
	return err == nil
}

// IShellLinkW 与 IPersistFile COM vtable 定义，用于通过系统 API 读取 .lnk 真实目标路径。
type shellLinkVtbl struct {
	queryInterface uintptr
	addRef         uintptr
	release        uintptr
	getPath        uintptr
	getIDList      uintptr
	setIDList      uintptr
	getDescription uintptr
	setDescription uintptr
	getWorkingDir  uintptr
	setWorkingDir  uintptr
	getArguments   uintptr
	setArguments   uintptr
	getHotkey      uintptr
	setHotkey      uintptr
	getShowCmd     uintptr
	setShowCmd     uintptr
	getIconLoc     uintptr
	setIconLoc     uintptr
	setRelPath     uintptr
	resolve        uintptr
	setPath        uintptr
}

type persistFileVtbl struct {
	queryInterface uintptr
	addRef         uintptr
	release        uintptr
	getClassID     uintptr
	isDirty        uintptr
	load           uintptr
	save           uintptr
	saveCompleted  uintptr
	getCurFile     uintptr
}

var (
	clsidShellLink = ole.NewGUID("{00021401-0000-0000-C000-000000000046}") // CLSID_ShellLink
	iidShellLinkW  = ole.NewGUID("{000214F9-0000-0000-C000-000000000046}") // IID_IShellLinkW
	iidPersistFile = ole.NewGUID("{0000010B-0000-0000-C000-000000000046}") // IID_IPersistFile
)

// readLnkTargetPath 通过 Windows COM (IShellLinkW + IPersistFile) 读取快捷方式的目标路径，
// 完整支持中文/Unicode 路径与特殊字符，避免手写二进制解析导致的编码与偏移问题。
func readLnkTargetPath(lnkPath string) (string, error) {
	initialized := false
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err == nil {
		initialized = true
	} else {
		if hr, ok := err.(*ole.OleError); !ok || (hr.Code() != 1 && hr.Code() != 0x80010106) {
			return "", fmt.Errorf("初始化 COM: %w", err)
		}
	}
	if initialized {
		defer ole.CoUninitialize()
	}

	sl, err := ole.CreateInstance(clsidShellLink, iidShellLinkW)
	if err != nil {
		return "", fmt.Errorf("创建 ShellLink: %w", err)
	}
	defer sl.Release()

	slThis := uintptr(unsafe.Pointer(sl))
	slVt := (*shellLinkVtbl)(unsafe.Pointer(sl.RawVTable))

	var pf *ole.IUnknown
	r, _, _ := syscall.SyscallN(slVt.queryInterface, slThis, uintptr(unsafe.Pointer(iidPersistFile)), uintptr(unsafe.Pointer(&pf)))
	if int32(r) < 0 {
		return "", fmt.Errorf("QueryInterface IPersistFile: 0x%08x", uint32(r))
	}
	defer pf.Release()

	pfThis := uintptr(unsafe.Pointer(pf))
	pfVt := (*persistFileVtbl)(unsafe.Pointer(pf.RawVTable))

	lnkPtr, err := syscall.UTF16PtrFromString(lnkPath)
	if err != nil {
		return "", err
	}
	if r, _, _ := syscall.SyscallN(pfVt.load, pfThis, uintptr(unsafe.Pointer(lnkPtr)), 0); int32(r) < 0 {
		return "", fmt.Errorf("加载 lnk 失败: 0x%08x", uint32(r))
	}

	var buf [syscall.MAX_PATH]uint16
	if r, _, _ := syscall.SyscallN(slVt.getPath, slThis, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0, 0); int32(r) < 0 {
		return "", fmt.Errorf("获取 lnk 目标路径失败: 0x%08x", uint32(r))
	}
	return syscall.UTF16ToString(buf[:]), nil
}

// IsContextMenuInstalled 检查右键注册表项是否存在且目标可执行文件有效
func IsContextMenuInstalled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress\command`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue("")
	if err != nil {
		return false
	}
	target := extractExePath(val)
	if target == "" {
		return false
	}
	_, err = os.Stat(target)
	return err == nil
}

// extractExePath 从诸如 `"C:\path\to.exe" "%1"` 或 `C:\path\to.exe` 中提取可执行文件路径
func extractExePath(cmdStr string) string {
	cmdStr = strings.TrimSpace(cmdStr)
	if strings.HasPrefix(cmdStr, `"`) {
		end := strings.Index(cmdStr[1:], `"`)
		if end != -1 {
			return cmdStr[1 : end+1]
		}
	}
	fields := strings.Fields(cmdStr)
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}

// RefreshIfStale 检查已安装的集成快捷方式/右键菜单，若指向旧路径或失效则使用 executablePath 自动自愈重写。
// 未安装的项绝不会被新建。
func RefreshIfStale(executablePath string, lang string) error {
	if executablePath == "" {
		return nil
	}
	// 避免在临时目录（如 mygo dev 编译产物）下执行自愈，破坏用户固定的快捷方式
	if isInTempDir(executablePath) {
		return nil
	}

	var errs []error

	// 1. 自愈右键上下文菜单
	if err := refreshContextMenuIfInstalled(executablePath, lang); err != nil {
		errs = append(errs, fmt.Errorf("刷新右键菜单: %w", err))
	}

	// 2. 自愈 SendTo
	if p, err := sendToLnkPath(); err == nil {
		if err := refreshLnkIfStale(p, executablePath, "--sendto"); err != nil {
			errs = append(errs, fmt.Errorf("刷新 SendTo: %w", err))
		}
	}

	// 3. 自愈桌面快捷方式
	if p, err := desktopLnkPath(); err == nil {
		if err := refreshLnkIfStale(p, executablePath, ""); err != nil {
			errs = append(errs, fmt.Errorf("刷新桌面快捷方式: %w", err))
		}
	}

	// 4. 自愈开始菜单快捷方式
	if p, err := startMenuLnkPath(); err == nil {
		if err := refreshLnkIfStale(p, executablePath, ""); err != nil {
			errs = append(errs, fmt.Errorf("刷新开始菜单: %w", err))
		}
	}

	return errors.Join(errs...)
}

func isInTempDir(p string) bool {
	temp := os.TempDir()
	if temp == "" {
		return false
	}
	cleanP := strings.ToLower(filepath.Clean(p))
	cleanTemp := strings.ToLower(filepath.Clean(temp))
	return strings.HasPrefix(cleanP, cleanTemp)
}

func refreshContextMenuIfInstalled(executablePath string, lang string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress\command`, registry.QUERY_VALUE)
	if err != nil {
		return nil // 未安装则不创建
	}
	val, _, err := k.GetStringValue("")
	k.Close()
	if err != nil {
		return nil
	}

	expectedCmd := fmt.Sprintf(`"%s" "%%1"`, executablePath)
	if strings.TrimSpace(val) != strings.TrimSpace(expectedCmd) {
		return RegisterContextMenu(executablePath, lang)
	}
	return nil
}

func refreshLnkIfStale(lnkPath string, executablePath string, arguments string) error {
	if _, err := os.Stat(lnkPath); err != nil {
		return nil // 未安装则不创建
	}

	target, err := lnkTargetReader(lnkPath)
	if err == nil && target != "" && strings.EqualFold(filepath.Clean(target), filepath.Clean(executablePath)) {
		return nil // 目标路径完全匹配且有效，无需刷新
	}

	return lnkCreator(lnkPath, executablePath, arguments)
}
