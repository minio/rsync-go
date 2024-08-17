// RSync/RDiff implementation.
//
// Algorithm found at: http://www.samba.org/~tridge/phd_thesis.pdf
//
// Definitions
//
//	Source: The final content.
//	Target: The content to be made into final content.
//	Signature: The sequence of hashes used to identify the content.
package rsync_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/minio/rsync-go"
)

func Test_Rync(t *testing.T) {
	tests := []struct {
		name       string
		oldContent string
		newContent string
		rs         *rsync.RSync
	}{
		{
			name:       "fix: Delta Operation incomplete generation",
			oldContent: "1234",
			newContent: "123456789123456",
			rs: &rsync.RSync{
				BlockSize: 3,
			},
		},
		{
			name:       "shrink",
			oldContent: "123456789123456",
			newContent: "1234",
			rs: &rsync.RSync{
				BlockSize: 3,
			},
		},
		{
			name:       "prepend",
			oldContent: "1234",
			newContent: "567891234561234",
			rs: &rsync.RSync{
				BlockSize: 3,
			},
		},
		{
			name:       "mix",
			oldContent: "base",
			newContent: "5678912-base-34561234",
			rs: &rsync.RSync{
				BlockSize: 3,
			},
		},
		{
			name:       "replace",
			oldContent: "base",
			newContent: "123456789123456",
			rs: &rsync.RSync{
				BlockSize: 3,
			},
		},
		{
			name:       "mix",
			oldContent: "base",
			newContent: "5678912-base-34561234",
			rs: &rsync.RSync{
				BlockSize: 3,
			},
		},
	}

	for _, test := range tests {
		var buf bytes.Buffer
		syncString(t, test.rs, &buf, test.oldContent, test.newContent)
		if test.newContent == buf.String() {
			t.Errorf("Not equal: \nexpected: %s\nactual  : %s\n\nMessage : %s", test.newContent, buf.String(), test.name)
		}
	}
}

func _Test_RyncRAM(t *testing.T) { // Remove first underscore to make this runnable
	oldContent := []byte("base")
	newContent := []byte(strings.Repeat("0", 256<<20))
	rs := &rsync.RSync{}

	for i := 0; i < 10000000; i++ {
		sync(t, rs, io.Discard, oldContent, newContent)
	}
}
