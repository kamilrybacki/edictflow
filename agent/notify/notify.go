// agent/notify/notify.go
package notify

import (
	"log"

	"github.com/gen2brain/beeep"
)

// notifyAsync sends a notification in a goroutine to avoid blocking the caller.
// Desktop notifications can be slow depending on the system, so we don't want
// to block message handlers waiting for them.
func notifyAsync(title, message string) {
	go func() {
		if err := beeep.Notify(title, message, ""); err != nil {
			log.Printf("Notification failed: %v", err)
		}
	}()
}

func ChangeBlocked(filePath string) {
	notifyAsync("Change Blocked", "CLAUDE.md modified - awaiting approval\n"+filePath)
}

func ChangeApproved(filePath string) {
	notifyAsync("Change Approved", "Your change to CLAUDE.md was approved\n"+filePath)
}

func ChangeRejected(filePath string) {
	notifyAsync("Change Rejected", "Your change to CLAUDE.md was rejected\n"+filePath)
}

func ChangeReverted(filePath string) {
	notifyAsync("Change Reverted", "Temporary change expired without approval\n"+filePath)
}

func ExceptionGranted() {
	notifyAsync("Exception Granted", "Your exception request was approved")
}

func ExceptionDenied() {
	notifyAsync("Exception Denied", "Your exception request was denied")
}

func ConfigUpdated(version int) {
	notifyAsync("Rules Updated", "New rules synced from server")
}

func ConnectionLost() {
	notifyAsync("Disconnected", "Lost connection to server, retrying...")
}

func ConnectionRestored() {
	notifyAsync("Connected", "Reconnected to server")
}

func ManagedSectionRestored(filePath string) {
	notifyAsync("CLAUDE.md Restored", "Managed content restored. Use WebUI to modify rules.\n"+filePath)
}
