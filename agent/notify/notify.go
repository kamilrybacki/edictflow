// agent/notify/notify.go
package notify

import (
	"github.com/gen2brain/beeep"
)

const appName = "Edictflow"

func ChangeBlocked(filePath string) {
	beeep.Notify(
		"Change Blocked",
		"CLAUDE.md modified - awaiting approval\n"+filePath,
		"",
	)
}

func ChangeApproved(filePath string) {
	beeep.Notify(
		"Change Approved",
		"Your change to CLAUDE.md was approved\n"+filePath,
		"",
	)
}

func ChangeRejected(filePath string) {
	beeep.Notify(
		"Change Rejected",
		"Your change to CLAUDE.md was rejected\n"+filePath,
		"",
	)
}

func ChangeReverted(filePath string) {
	beeep.Notify(
		"Change Reverted",
		"Temporary change expired without approval\n"+filePath,
		"",
	)
}

func ExceptionGranted() {
	beeep.Notify(
		"Exception Granted",
		"Your exception request was approved",
		"",
	)
}

func ExceptionDenied() {
	beeep.Notify(
		"Exception Denied",
		"Your exception request was denied",
		"",
	)
}

func ConfigUpdated(version int) {
	beeep.Notify(
		"Rules Updated",
		"New rules synced from server",
		"",
	)
}

func ConnectionLost() {
	beeep.Notify(
		"Disconnected",
		"Lost connection to server, retrying...",
		"",
	)
}

func ConnectionRestored() {
	beeep.Notify(
		"Connected",
		"Reconnected to server",
		"",
	)
}

func ManagedSectionRestored(filePath string) {
	beeep.Notify(
		"CLAUDE.md Restored",
		"Managed content restored. Use WebUI to modify rules.\n"+filePath,
		"",
	)
}
