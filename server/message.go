/*
 * MIT License
 *
 * Copyright (c)  ShiChao
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/*
 * Revision History:
 *     Initial: 2018/03/02        ShiChao
 */

package server

import (
	"encoding/binary"
	"bytes"
	"fmt"
)

const (
	lenIdx = 4
	msgIdx = 8
)

type message struct {
	seq    uint32
	length uint32
	msg    []byte
}

// NewMsg generate message from msg string and seqNumber
func NewMsg(seq uint32, msg string) *message {
	return &message{
		seq,
		uint32(len([]byte(msg))),
		[]byte(msg),
	}
}

func (m *message) Msg() []byte {
	return m.msg
}

func (m *message) Len() int {
	return len(m.msg)
}

// Encode encode message to []byte
func (m *message) Encode() ([]byte, error) {
	var (
		err error
		buf bytes.Buffer
		b   = make([]byte, 4)
	)

	binary.BigEndian.PutUint32(b, m.seq)
	buf.Write(b[:lenIdx])
	binary.BigEndian.PutUint32(b, m.length)
	buf.Write(b[:lenIdx])

	b, err = encode(m.msg)
	if err != nil {
		return []byte{}, err
	}
	_, err = buf.Write(b)
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}


// Decode decode []byte to message
func (m *message) Decode(b []byte) error {
	var err error

	seq, err := decode(b[:lenIdx], lenIdx)
	if err != nil {
		return err
	}

	m.seq = binary.BigEndian.Uint32(seq)

	l, err := decode(b[lenIdx:msgIdx], lenIdx)
	if err != nil {
		return err
	}
	m.length = binary.BigEndian.Uint32(l)

	m.msg, err = decode(b[msgIdx:], int32(m.length))

	return nil
}


// DecodeMsgs decode multiple messages at once
func DecodeMsgs(b []byte) (msgs []*message, err error) {
	for {
		msg := NewMsg(0, "")
		err = msg.Decode(b)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		msgs = append(msgs, msg)
		b = b[msgIdx+msg.Len():]
		if len(b) == 0 {
			return
		}
	}
}

func encode(b []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, b)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func decode(b []byte, n int32) ([]byte, error) {
	var ret = make([]byte, n)
	buf := bytes.NewReader(b)

	err := binary.Read(buf, binary.BigEndian, ret)
	if err != nil {
		return []byte{}, err
	}
	return ret, nil
}
