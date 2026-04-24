package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

const (
	typeString  = "string"
	typeError   = "error"
	typeInteger = "integer"
	typeBulk    = "bulk"
	typeArray   = "array"
	typeNull    = "null"
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

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() ([]byte, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(line, "\r\n") {
		return nil, fmt.Errorf("invalid RESP line ending")
	}
	return []byte(line[:len(line)-2]), nil
}

func (r *Resp) readInteger() (int, error) {
	line, err := r.readLine()
	if err != nil {
		return 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}

func (r *Resp) Read() (Value, error) {
	respType, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch respType {
	case STRING:
		return r.readString()
	case ERROR:
		return r.readError()
	case INTEGER:
		return r.readRespInteger()
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %q", respType)
	}
}

func (r *Resp) readString() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{typ: typeString, str: string(line)}, nil
}

func (r *Resp) readError() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{typ: typeError, str: string(line)}, nil
}

func (r *Resp) readRespInteger() (Value, error) {
	num, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}
	return Value{typ: typeInteger, num: num}, nil
}

func (r *Resp) readArray() (Value, error) {
	length, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}
	if length < 0 {
		return Value{typ: typeNull}, nil
	}

	v := Value{
		typ:   typeArray,
		array: make([]Value, 0, length),
	}

	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return Value{}, err
		}
		v.array = append(v.array, val)
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	length, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}
	if length == -1 {
		return Value{typ: typeNull}, nil
	}
	if length < -1 {
		return Value{}, fmt.Errorf("invalid bulk string length: %d", length)
	}

	buf := make([]byte, length+2)
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return Value{}, err
	}
	if buf[length] != '\r' || buf[length+1] != '\n' {
		return Value{}, fmt.Errorf("invalid bulk string ending")
	}

	return Value{typ: typeBulk, bulk: string(buf[:length])}, nil
}

func (v Value) Marshal() []byte {
	switch v.typ {
	case typeArray:
		return v.marshalArray()
	case typeBulk:
		return v.marshalBulk()
	case typeString:
		return v.marshalString()
	case typeInteger:
		return v.marshalInteger()
	case typeNull:
		return v.marshalNull()
	case typeError:
		return v.marshalError()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalInteger() []byte {
	var bytes []byte
	bytes = append(bytes, INTEGER)
	bytes = append(bytes, strconv.Itoa(v.num)...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len(v.array))...)
	bytes = append(bytes, '\r', '\n')

	for i := range v.array {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	_, err := w.writer.Write(v.Marshal())
	return err
}
