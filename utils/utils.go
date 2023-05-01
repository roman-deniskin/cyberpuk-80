package utils

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	// Кнопки
	MB_OK               = 0x00000000 // окно сообщения содержит одну кнопку: OK.
	MB_OKCANCEL         = 0x00000001 // окно сообщения содержит две кнопки: OK и Cancel.
	MB_ABORTRETRYIGNORE = 0x00000002 // окно сообщения содержит три кнопки: Abort, Retry и Ignore.
	MB_YESNOCANCEL      = 0x00000003 // окно сообщения содержит три кнопки: Yes, No и Cancel.
	MB_YESNO            = 0x00000004 // окно сообщения содержит две кнопки: Yes и No.
	MB_RETRYCANCEL      = 0x00000005 // окно сообщения содержит две кнопки: Retry и Cancel.
	// Значки
	MB_ICONERROR       = 0x00000010 // окно сообщения содержит значок ошибки.
	MB_ICONQUESTION    = 0x00000020 // окно сообщения содержит значок вопроса.
	MB_ICONWARNING     = 0x00000030 // окно сообщения содержит значок предупреждения.
	MB_ICONINFORMATION = 0x00000040 // окно сообщения содержит значок информации.
	// Модификаторы
	MB_DEFBUTTON1  = 0x00000000 // первая кнопка (обычно OK или Yes) выбрана по умолчанию.
	MB_DEFBUTTON2  = 0x00000100 // вторая кнопка (обычно Cancel или No) выбрана по умолчанию.
	MB_DEFBUTTON3  = 0x00000200 // третья кнопка (обычно Retry или Ignore) выбрана по умолчанию.
	MB_DEFBUTTON4  = 0x00000300 // четвертая кнопка (если доступна) выбрана по умолчанию.
	MB_APPLMODAL   = 0x00000000 // окно сообщения блокирует только текущее приложение.
	MB_SYSTEMMODAL = 0x00001000 // окно сообщения блокирует все приложения на рабочем столе.
	MB_TASKMODAL   = 0x00002000 // окно сообщения блокирует текущую задачу.
)

func MessageBox(title, text string, style int) int {
	user32 := windows.NewLazySystemDLL("user32.dll")
	MessageBoxW := user32.NewProc("MessageBoxW")

	ret, _, _ := MessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(style),
	)
	return int(ret)
}
