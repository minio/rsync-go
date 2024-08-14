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
	"math/rand/v2"
	"testing"

	"github.com/minio/rsync-go"
)

func FuzzSync(f *testing.F) {
	f.Add(bytes.Repeat([]byte("0"), 8192))
	f.Fuzz(func(t *testing.T, message []byte) {
		if len(message) == 0 {
			return
		}

		rs := &rsync.RSync{
			BlockSize: 1 + rand.IntN(len(message))*2,
		}

		n := 1 + rand.IntN(len(message))
		if n > len(message) {
			n = len(message)
		}

		var mode string
		var oldContent, newContent []byte
		switch rand.IntN(4) {
		case 0:
			mode = "expand"
			oldContent = message[:n]
			newContent = message
		case 1:
			mode = "shrink"
			oldContent = message
			newContent = message[:n]
		case 2:
			mode = "prepend"
			oldContent = message[n:]
			newContent = message
		case 3:
			m := 1 + rand.IntN(n)
			if m > n {
				m = n
			}

			mode = "mix"
			oldContent = message[m:n]
			newContent = message
		}

		var buf bytes.Buffer
		sync(t, rs, &buf, oldContent, newContent)
		if bytes.Compare(newContent, buf.Bytes()) != 0 {
			t.Errorf("Not equal: \nexpected : %#v\nactual   : %#v\n\nmessage  : %#v\nblocksize: %d\nmode     : %s",
				newContent, buf.Bytes(), message, rs.BlockSize, mode)
		}
	})
}

func syncString(t *testing.T, rs *rsync.RSync, w io.Writer, oldContent, newContent string) {
	sync(t, rs, w, []byte(oldContent), []byte(newContent))
}

func sync(t *testing.T, rs *rsync.RSync, w io.Writer, oldContent, newContent []byte) {
	oldReader := bytes.NewReader(oldContent)

	// here we store the whole signature in a byte slice,
	// but it could just as well be sent over a network connection for example
	signatures := make([]rsync.BlockHash, 0, 10)
	err := rs.CreateSignature(oldReader, func(bl rsync.BlockHash) error {
		signatures = append(signatures, bl)
		return nil
	})
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	var currentReader io.Reader = bytes.NewReader(newContent)
	opsOut := make(chan rsync.Operation)

	go func() {
		defer close(opsOut)

		err := rs.CreateDelta(currentReader, signatures, func(op rsync.Operation) error {
			opsOut <- op
			return nil
		})
		if err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}
	}()

	// Already read once for the signatures so we rewind the reader.
	_, err = oldReader.Seek(0, io.SeekStart)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	err = rs.ApplyDelta(w, oldReader, opsOut)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
}
