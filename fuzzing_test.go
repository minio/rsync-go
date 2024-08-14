package rsync_test

import (
	"bytes"
	"io"
	"math/rand/v2"
	"testing"

	"github.com/minio/rsync-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzSync(f *testing.F) {
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

		mode := "expand"
		oldContent := message[:n]
		newContent := message
		if rand.IntN(2) == 0 {
			mode = "shrink"
			oldContent = message
			newContent = message[:n]
		}

		result := sync(t, rs, oldContent, newContent)
		assert.Equalf(t, newContent, result,
			"%d,%s:\t%#v\noldContent:\t%#v\nnewContent:\t%#v",
			rs.BlockSize, mode, message, oldContent, newContent)
	})
}

func syncString(t *testing.T, rs *rsync.RSync, oldContent, newContent string) string {
	return string(sync(t, rs, []byte(oldContent), []byte(newContent)))
}

func sync(t *testing.T, rs *rsync.RSync, oldContent, newContent []byte) []byte {
	oldReader := bytes.NewReader(oldContent)

	// here we store the whole signature in a byte slice,
	// but it could just as well be sent over a network connection for example
	sig := make([]rsync.BlockHash, 0, 10)
	writeSignature := func(bl rsync.BlockHash) error {
		sig = append(sig, bl)
		return nil
	}
	err := rs.CreateSignature(oldReader, writeSignature)
	require.NoError(t, err)

	var currentReader io.Reader = bytes.NewReader(newContent)

	opsOut := make(chan rsync.Operation)
	writeOperation := func(op rsync.Operation) error {
		opsOut <- op
		return nil
	}

	go func() {
		defer close(opsOut)

		err := rs.CreateDelta(currentReader, sig, writeOperation)
		require.NoError(t, err)
	}()

	var buf bytes.Buffer
	_, err = oldReader.Seek(0, io.SeekStart)
	require.NoError(t, err)

	err = rs.ApplyDelta(&buf, oldReader, opsOut)
	require.NoError(t, err)

	return buf.Bytes()
}
