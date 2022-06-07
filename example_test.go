package rsync_test

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/minio/rsync-go"
)

func ExampleRsync() {
	oldReader := strings.NewReader("I am the original content")

	rs := &rsync.RSync{}

	// here we store the whole signature in a byte slice,
	// but it could just as well be sent over a network connection for example
	sig := make([]rsync.BlockHash, 0, 10)
	writeSignature := func(bl rsync.BlockHash) error {
		sig = append(sig, bl)
		return nil
	}
	rs.CreateSignature(oldReader, writeSignature)

	var currentReader io.Reader
	currentReader = strings.NewReader("I am the new content")

	opsOut := make(chan rsync.Operation)
	writeOperation := func(op rsync.Operation) error {
		opsOut <- op
		return nil
	}

	go func() {
		defer close(opsOut)
		rs.CreateDelta(currentReader, sig, writeOperation)
	}()

	var newWriter strings.Builder
	oldReader.Seek(0, os.SEEK_SET)

	rs.ApplyDelta(&newWriter, oldReader, opsOut)

	fmt.Println(newWriter.String())
	// Output: I am the new content
}
