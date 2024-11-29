/* SPDX-FileCopyrightText: 2024 Tristan Partin <tristan@partin.io>
 *
 * SPDX-License-Identifier: MIT-0
 */

package notify

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	_ "unsafe"
)

type NotifyAccessValue string

const (
	NotifyAccessAll  NotifyAccessValue = "all"
	NotifyAccessExec NotifyAccessValue = "exec"
	NotifyAccessMain NotifyAccessValue = "main"
	NotifyAccessNone NotifyAccessValue = "none"
)

type NotifyWatchdogValue string

const (
	NotifyWatchdogOne     NotifyWatchdogValue = "1"
	NotifyWatchdogTrigger NotifyWatchdogValue = "trigger"
)

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64

// Notify may be called by a service to notify the service manager about state
// changes.
func Notify(msg string) error {
	socketPath := os.Getenv("NOTIFY_SOCKET")
	if socketPath == "" {
		return nil
	}

	if socketPath[0] != '/' && socketPath[0] != '@' {
		return syscall.EAFNOSUPPORT
	}

	if socketPath[0] == '@' {
		socketPath = "\000" + socketPath[1:]
	}

	conn, err := net.Dial("unixgram", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(msg))

	return err
}

// Format string based bariant of Notify
func Notifyf(format string, a ...any) error {
	return Notify(fmt.Sprintf(format, a...))
}

// Refer to the BARRIER=1 message
func Barrier() error {
	return Notify("BARRIER=1")
}

// Refer to the BUSERROR message
func BusError(dbusErrorCode string) error {
	return Notifyf("BUSERROR=%s", dbusErrorCode)
}

// Refer to the ERRNO message
func Errno(err syscall.Errno) error {
	return Notifyf("ERRNO=%d", int(err))
}

// Refer to the EXIT_STATUS message
func ExitStatus(exitStatus uint8) error {
	return Notifyf("EXIT_STATUS=%d", exitStatus)
}

// Refer to the EXTEND_TIMEOUT_USEC message
func ExtendTimeoutUsec(microsecs int64) error {
	return Notifyf("EXTEND_TIMEOUT_USEC=%d", microsecs)
}

// Refer to the FDSTORE=1, FDNAME, and FDPOLL=0 messages
func FdStore(name string, poll bool) error {
	var msg strings.Builder

	// 9 = len("FDSTORE=1")
	// 1 + 8 + 255 = len("\nFDNAME=name"), name has a max length of 255
	// 1 + 8 = len("\nFDPOLL=0")
	msg.Grow(9 + 1 + 8 + 255 + 1 + 8)
	msg.WriteString("FDSTORE=1")

	if name != "" {
		msg.WriteString(fmt.Sprintf("\nFDNAME=%s", name))
	}

	if !poll {
		msg.WriteString("\nFDPOLL=0")
	}

	return Notify(msg.String())
}

// Refer to the FDSTOREREMOVE message
func FdStoreRemove(name string) error {
	return Notifyf("FDSTOREREMOVE=1\nFDNAME=%s", name)
}

// Refer to the MAINPID message
func MainPID(pid int32) error {
	return Notifyf("MAINPID=%d", pid)
}

// Refer to the NOTIFYACCESS message
func NotifyAccess(value NotifyAccessValue) error {
	return Notifyf("NOTIFYACCESS=%s", value)
}

// Refer to the RELOADING=1 message
func Reloading() error {
	microsecs := nanotime() / 1000
	return Notifyf("RELOADING=1\nMONOTONIC_USEC=%d", microsecs)
}

// Refer to the READY=1 message
func Ready() error {
	return Notify("READY=1")
}

// Refer to the STATUS message
func Status(status string) error {
	return Notifyf("STATUS=%s", status)
}

// Refer to the STOPPING=1 message
func Stopping() error {
	return Notify("STOPPING=1")
}

// Refer to the WATCHDOG message
func Watchdog(value NotifyWatchdogValue) error {
	return Notifyf("WATCHDOG=%s", value)
}

// Refer to the WATCHDOG_USEC message
func WatchdogUsec(microsecs int64) error {
	return Notifyf("WATCHDOG_USEC=%d", microsecs)
}
