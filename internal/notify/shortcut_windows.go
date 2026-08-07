//go:build windows

package notify

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// Windows Toast 通知要求开始菜单中存在带 AppUserModelID 属性的快捷方式，
// 否则通知会被系统静默丢弃。本文件通过 COM（IShellLink + IPropertyStore +
// IPersistFile）创建该快捷方式。

// IShellLinkW vtable：前 3 项为 IUnknown 方法，SetPath 位于索引 20。
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

type propertyStoreVtbl struct {
	queryInterface uintptr
	addRef         uintptr
	release        uintptr
	getCount       uintptr
	getAt          uintptr
	getValue       uintptr
	setValue       uintptr
	commit         uintptr
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

type propertyKey struct {
	fmtid ole.GUID
	pid   uint32
}

// PROPVARIANT：仅使用 VT_LPWSTR 分支，pwszVal 位于偏移 8。
type propVariant struct {
	vt      uint16
	w1      uint16
	w2      uint16
	w3      uint16
	pwszVal uintptr
}

const (
	vtLpwstr           = 31 // VT_LPWSTR
	clsCtxInprocServer = 1  // CLSCTX_INPROC_SERVER
	okHRESULT          = 0  // S_OK
)

var (
	clsidShellLink     = ole.NewGUID("{00021401-0000-0000-C000-000000000046}") // CLSID_ShellLink
	iidShellLinkW      = ole.NewGUID("{000214F9-0000-0000-C000-000000000046}") // IID_IShellLinkW
	iidPropertyStore   = ole.NewGUID("{886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99}") // IID_IPropertyStore
	iidPersistFile     = ole.NewGUID("{0000010B-0000-0000-C000-000000000046}") // IID_IPersistFile
	pkeyAppUserModelID = propertyKey{
		fmtid: ole.GUID{
			Data1: 0x9F4C2855, Data2: 0x9F79, Data3: 0x4B39,
			Data4: [8]byte{0xA8, 0xD0, 0xE1, 0xD4, 0x2D, 0xE1, 0xD5, 0xF3},
		},
		pid: 5,
	}
)

// EnsureShortcut 在开始菜单 Programs\Videopress 下创建指向 executablePath、
// 带 AppUserModelID 属性的快捷方式，供 Toast 通知使用。
func EnsureShortcut(executablePath string) error {
	// 进程可能已被其他模块（如 Wails 依赖链）以 STA 模式初始化 COM。
	// 统一以 STA 请求，并容忍 S_FALSE（已初始化）与 RPC_E_CHANGED_MODE（模式冲突），
	// 这两种情况下 COM 仍可直接使用，且不得再调用 CoUninitialize。
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
		return fmt.Errorf("创建 ShellLink 对象: %w", err)
	}
	defer sl.Release()

	slThis := uintptr(unsafe.Pointer(sl))
	slVt := (*shellLinkVtbl)(unsafe.Pointer(sl.RawVTable))

	// 设置目标路径
	targetPtr, err := syscall.UTF16PtrFromString(executablePath)
	if err != nil {
		return err
	}
	if r, _, _ := syscall.SyscallN(slVt.setPath, slThis, uintptr(unsafe.Pointer(targetPtr))); int32(r) != okHRESULT {
		return fmt.Errorf("设置快捷方式目标失败: 0x%08x", uint32(r))
	}

	// 设置 AppUserModelID 属性
	ps, err := queryInterface(slThis, slVt.queryInterface, iidPropertyStore)
	if err != nil {
		return err
	}
	defer ps.Release()

	psThis := uintptr(unsafe.Pointer(ps))
	psVt := (*propertyStoreVtbl)(unsafe.Pointer(ps.RawVTable))

	appIDPtr, err := syscall.UTF16PtrFromString(AppID)
	if err != nil {
		return err
	}
	pv := propVariant{vt: vtLpwstr, pwszVal: uintptr(unsafe.Pointer(appIDPtr))}
	if r, _, _ := syscall.SyscallN(psVt.setValue, psThis, uintptr(unsafe.Pointer(&pkeyAppUserModelID)), uintptr(unsafe.Pointer(&pv))); int32(r) != okHRESULT {
		return fmt.Errorf("设置 AppUserModelID 失败: 0x%08x", uint32(r))
	}
	if r, _, _ := syscall.SyscallN(psVt.commit, psThis); int32(r) != okHRESULT {
		return fmt.Errorf("提交快捷方式属性失败: 0x%08x", uint32(r))
	}

	// 保存 .lnk 文件
	pf, err := queryInterface(slThis, slVt.queryInterface, iidPersistFile)
	if err != nil {
		return err
	}
	defer pf.Release()

	pfThis := uintptr(unsafe.Pointer(pf))
	pfVt := (*persistFileVtbl)(unsafe.Pointer(pf.RawVTable))

	lnkPath, err := shortcutPath()
	if err != nil {
		return err
	}
	lnkPtr, err := syscall.UTF16PtrFromString(lnkPath)
	if err != nil {
		return err
	}
	if r, _, _ := syscall.SyscallN(pfVt.save, pfThis, uintptr(unsafe.Pointer(lnkPtr)), 1); int32(r) != okHRESULT {
		return fmt.Errorf("保存快捷方式失败: 0x%08x", uint32(r))
	}
	return nil
}

func shortcutPath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("未设置 APPDATA 环境变量")
	}
	dir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Videopress")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "Videopress.lnk"), nil
}

// queryInterface 从已知接口对象上请求指定接口。
func queryInterface(this uintptr, qiFn uintptr, iid *ole.GUID) (*ole.IUnknown, error) {
	var out *ole.IUnknown
	r, _, _ := syscall.SyscallN(qiFn, this, uintptr(unsafe.Pointer(iid)), uintptr(unsafe.Pointer(&out)))
	if int32(r) != okHRESULT {
		return nil, fmt.Errorf("QueryInterface %s 失败: 0x%08x", iid.String(), uint32(r))
	}
	return out, nil
}
