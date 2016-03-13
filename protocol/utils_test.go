package protocol

import "testing"

func TestToID(t *testing.T) {
	message := "This !s a completely regular t3st1ng nick!~~"
	result := ToID(message)
	expect := UserID("thissacompletelyregulart3st1ngnick")
	if result != expect {
		t.Errorf("ToId(%q) = %+v, want %+v", message, result, expect)
	}
}
