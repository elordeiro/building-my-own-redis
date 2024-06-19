package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
	CRLF    = "\r\n"
)

type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

type Resp struct {
	reader *bufio.Reader
}

type Writer struct {
	writer io.Writer
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	bytes := v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Deserializer ---------------------------------------------------------------
func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n++
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return int(i64), n, nil
}

func (r *Resp) readArray() (v Value, err error) {
	v.typ = "array"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	v.array = make([]Value, 0)
	for range len {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		v.array = append(v.array, val)
	}

	return v, nil
}

func (r *Resp) readBulk() (v Value, err error) {
	v.typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, nil
	}

	bulk := make([]byte, len)
	r.reader.Read(bulk)
	v.bulk = string(bulk)

	// Read trailing byte from CRLF
	r.readLine()

	return v, nil
}

func (r *Resp) Read() (v Value, err error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknow type: %v", string(_type))
		return Value{}, nil
	}
}

// ----------------------------------------------------------------------------

// Serializer -----------------------------------------------------------------
func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshalNull()
	case "error":
		return v.marshalError()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() (bytes []byte) {
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, CRLF...)
	return bytes
}

func (v Value) marshalBulk() (bytes []byte) {
	bytes = append(bytes, BULK)
	strconv.AppendInt(bytes, int64(len((v.bulk))), 10)
	bytes = append(bytes, CRLF...)
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, CRLF...)

	return bytes
}

func (v Value) marshalArray() (bytes []byte) {
	len := len(v.array)
	bytes = append(bytes, ARRAY)
	strconv.AppendInt(bytes, int64(len), 10)
	bytes = append(bytes, CRLF...)

	for i := range len {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshalError() (bytes []byte) {
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, CRLF...)

	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

// ----------------------------------------------------------------------------
