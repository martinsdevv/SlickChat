package domain

import "testing"

func TestRoom_CanPersistAttachments(t *testing.T) {
	t.Parallel()

	r := &Room{ZeroLogging: false, ParanoidMode: false}
	if !r.CanPersistAttachments() {
		t.Fatal("expected persistible attachments")
	}

	r.ZeroLogging = true
	if r.CanPersistAttachments() {
		t.Fatal("zero logging must not persist attachments metadata")
	}

	r.ZeroLogging = false
	r.ParanoidMode = true
	if r.CanPersistAttachments() {
		t.Fatal("paranoid must not persist attachments metadata")
	}
}
