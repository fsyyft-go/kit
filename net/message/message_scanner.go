package message

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	cockroachdbErrors "github.com/cockroachdb/errors"
)

func scanMessage(data []byte, atEOF bool) (int, []byte, error) {
	var advance int
	var token []byte
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdbErrors.Newf("解包过程发生异常：$[1]v。", r)
		}
	}()

	if atEOF {
		err = io.EOF
	} else {
		if len(data) > 4 {
			messageLength := uint16(0)

			if errReadLength := binary.Read(bytes.NewReader(data[2:4]), binary.BigEndian, &messageLength); nil != errReadLength {
				err = errReadLength
			} else if int(messageLength+4) <= len(data) {
				advance = int(messageLength + 4)
				token = data[:int(messageLength+4)]
			}
		}
	}

	return advance, token, err
}

func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)

	scanner.Split(scanMessage)

	return scanner
}
