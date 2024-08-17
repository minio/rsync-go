package rsync_test

import (
	"fmt"
	"io"
	"log"
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
	err := rs.CreateSignature(oldReader, writeSignature)
	if err != nil {
		log.Fatal(err)
	}

	var currentReader io.Reader
	currentReader = strings.NewReader("I am the new content")

	opsOut := make(chan rsync.Operation)
	writeOperation := func(op rsync.Operation) error {
		opsOut <- op
		return nil
	}

	go func() {
		defer close(opsOut)
		err := rs.CreateDelta(currentReader, sig, writeOperation)
		if err != nil {
			log.Fatal(err)
		}
	}()

	var newWriter strings.Builder
	_, err = oldReader.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}

	err = rs.ApplyDelta(&newWriter, oldReader, opsOut)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(newWriter.String())
	// Output: I am the new content
}
