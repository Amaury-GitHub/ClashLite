package main

import (
	"archive/zip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/walk"
)

var LocalVersion string
var RemoteVersion string
var ClashStatus string
var TunStatus string

func GetStatus() {
	//读取Clash运行状态
	Command := exec.Command("powershell", "Get-Process clash-windows-amd64")
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err := Command.Output()
	if err == nil {
		ClashStatus = "运行中"
	} else {
		ClashStatus = "未运行"
	}
	//读取Tun运行状态
	Command = exec.Command("powershell", "Get-NetAdapter Clash")
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err = Command.Output()
	if err == nil {
		TunStatus = "运行中"
	} else {
		TunStatus = "未运行"
	}
	//10s循环
	time.AfterFunc(10*time.Second, GetStatus)
}
func GetOnline() {
	//更新Country.mmdb
	Command := exec.Command("powershell", "(New-Object Net.WebClient).DownloadFile('https://cdn.jsdelivr.net/gh/Hackl0us/GeoIP2-CN@release/Country.mmdb','Country.mmdb.temp')")
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err := Command.Output()
	if err == nil {
		os.Rename("Country.mmdb.temp", "Country.mmdb")
	}
	//获取在线Clash版本
	Command = exec.Command("powershell", "(New-Object Net.WebClient).DownloadString('https://github.com/Dreamacro/clash/releases/tag/premium')")
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	Output, err := Command.Output()
	if err == nil {
		RemoteVersion = string(Output)[strings.Index(string(Output), "Premium")+8 : strings.Index(string(Output), "Dreamacro")-4]
	}
	//1h循环
	time.AfterFunc(1*time.Hour, GetOnline)
}
func StartClash() {
	//读取Clash本地版本
	Command := exec.Command("cmd", "/c", "clash-windows-amd64 -v")
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	Output, _ := Command.Output()
	LocalVersion = string(Output)[6:16]
	//启动Clash
	Command = exec.Command("cmd", "/c", "clash-windows-amd64 -d .")
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	Command.Start()
}
func StopClash() {
	//卸载Clash Tunnel网卡
	Command := exec.Command("powershell", `pnputil /remove-device (Get-PnpDevice | Where-Object{$_.Name -eq "Clash Tunnel"}).InstanceId`)
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	Command.Run()
	//停止Clash
	Command = exec.Command("taskkill", "/f", "/im", "clash-windows-amd64.exe")
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	Command.Run()
}
func init() {
	//阻止多次启动
	Mutex, _ := syscall.UTF16PtrFromString("Clashlite")
	_, _, err := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateMutexW").Call(0, 0, uintptr(unsafe.Pointer(Mutex)))
	if int(err.(syscall.Errno)) != 0 {
		os.Exit(1)
	}

	//读取Clash本地版本
	Command := exec.Command("cmd.exe", "/c", "clash-windows-amd64 -v")
	Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	Output, _ := Command.Output()
	LocalVersion = string(Output)[6:16]

	//循环
	go GetStatus()
	go GetOnline()
}

func main() {
	//定义托盘图标文字
	MainWindow, _ := walk.NewMainWindow()
	Icon, _ := walk.Resources.Icon("icon.ico")
	NotifyIcon, _ := walk.NewNotifyIcon(MainWindow)
	defer NotifyIcon.Dispose()
	NotifyIcon.SetIcon(Icon)
	NotifyIcon.SetToolTip("ClashLite")
	NotifyIcon.SetVisible(true)
	//定义左键显示
	NotifyIcon.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		NotifyIcon.ShowMessage("Clash: "+ClashStatus+"\r\n"+"Tun: "+TunStatus, "本地版本: "+LocalVersion+"\r\n"+"在线版本: "+RemoteVersion)
	})
	//定义右键菜单
	blank1 := walk.NewAction()
	blank1.SetText("-")
	blank2 := walk.NewAction()
	blank2.SetText("-")
	blank3 := walk.NewAction()
	blank3.SetText("-")
	blank4 := walk.NewAction()
	blank4.SetText("-")
	blank5 := walk.NewAction()
	blank5.SetText("-")
	blank6 := walk.NewAction()
	blank6.SetText("-")
	Start := walk.NewAction()
	Start.SetText("启动Clash")
	Stop := walk.NewAction()
	Stop.SetText("停止Clash")
	Update := walk.NewAction()
	Update.SetText("升级Clash")
	Razord := walk.NewAction()
	Razord.SetText("打开Razord")
	Yacd := walk.NewAction()
	Yacd.SetText("打开Yacd")
	EnableLoopback := walk.NewAction()
	EnableLoopback.SetText("打开EnableLoopback")
	Exit := walk.NewAction()
	Exit.SetText("Exit")

	NotifyIcon.ContextMenu().Actions().Add(Start)
	NotifyIcon.ContextMenu().Actions().Add(blank1)
	NotifyIcon.ContextMenu().Actions().Add(Stop)
	NotifyIcon.ContextMenu().Actions().Add(blank2)
	NotifyIcon.ContextMenu().Actions().Add(Update)
	NotifyIcon.ContextMenu().Actions().Add(blank3)
	NotifyIcon.ContextMenu().Actions().Add(Razord)
	NotifyIcon.ContextMenu().Actions().Add(blank4)
	NotifyIcon.ContextMenu().Actions().Add(Yacd)
	NotifyIcon.ContextMenu().Actions().Add(blank5)
	NotifyIcon.ContextMenu().Actions().Add(EnableLoopback)
	NotifyIcon.ContextMenu().Actions().Add(blank6)
	NotifyIcon.ContextMenu().Actions().Add(Exit)

	//启动Clash
	Start.Triggered().Attach(func() {
		//停止Clash
		StopClash()
		//启动Clash
		StartClash()
	})
	//停止Clash
	Stop.Triggered().Attach(func() {
		//停止Clash
		StopClash()
	})
	//升级Clash
	Update.Triggered().Attach(func() {
		//获取在线Clash版本
		Command := exec.Command("powershell", "(New-Object Net.WebClient).DownloadString('https://github.com/Dreamacro/clash/releases/tag/premium')")
		Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		Output, err := Command.Output()
		if err != nil {
			return
		}
		RemoteVersion = string(Output)[strings.Index(string(Output), "Premium")+8 : strings.Index(string(Output), "Dreamacro")-4]
		//如果版本一样不更新
		if RemoteVersion == LocalVersion {
			return
		}
		//生成最新版本的下载地址
		DownloadLink := "(New-Object Net.WebClient).DownloadFile('https://github.com/Dreamacro/clash/releases/download/premium/clash-windows-amd64-" + RemoteVersion + ".zip','clash.zip')"
		//下载压缩包
		Command = exec.Command("powershell", DownloadLink)
		Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_, err = Command.Output()
		if err != nil {
			return
		}
		//停止Clash
		StopClash()
		//删除文件
		err = os.Remove("clash-windows-amd64.exe")
		if err != nil {
			return
		}
		//获取当前目录
		DestinationPath, _ := os.Getwd()
		//打开压缩包
		Archive, err := zip.OpenReader("clash.zip")
		if err != nil {
			return
		}
		defer Archive.Close()
		//解压缩
		DestinationFile := filepath.Join(DestinationPath, Archive.File[0].Name)
		ArchiveFileOpen, err := Archive.File[0].Open()
		if err != nil {
			ArchiveFileOpen.Close()
			return
		}
		DestinationFileOpen, err := os.OpenFile(DestinationFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, Archive.File[0].Mode())
		if err != nil {
			ArchiveFileOpen.Close()
			DestinationFileOpen.Close()
			return
		}
		_, err = io.Copy(DestinationFileOpen, ArchiveFileOpen)
		if err != nil {
			ArchiveFileOpen.Close()
			DestinationFileOpen.Close()
			return
		}
		ArchiveFileOpen.Close()
		DestinationFileOpen.Close()
		Archive.Close()
		//删除压缩包
		os.Remove("clash.zip")
		//启动Clash
		StartClash()
	})
	//打开Razord
	Razord.Triggered().Attach(func() {
		Command := exec.Command("cmd", "/c", "start", "http://localhost:9090/ui/razord")
		Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		Command.Start()
	})
	//打开Yacd
	Yacd.Triggered().Attach(func() {
		Command := exec.Command("cmd", "/c", "start", "http://localhost:9090/ui/yacd")
		Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		Command.Start()
	})
	//打开EnableLoopback
	EnableLoopback.Triggered().Attach(func() {
		Command := exec.Command("cmd", "/c", "start", "EnableLoopback.exe")
		Command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		Command.Start()
	})
	//Exit
	Exit.Triggered().Attach(func() {
		StopClash()
		walk.App().Exit(0)
	})

	MainWindow.Run()
}
