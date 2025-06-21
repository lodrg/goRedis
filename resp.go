package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RESP 数据类型
const (
	RESP_SIMPLE_STRING = '+'
	RESP_ERROR         = '-'
	RESP_INTEGER       = ':'
	RESP_BULK_STRING   = '$'
	RESP_ARRAY         = '*'
)

// RESPValue 表示一个 RESP 值
type RESPValue struct {
	Type   byte
	Str    string
	Num    int64
	IsNull bool
	Array  []*RESPValue
}

// NewRESPValue 创建新的 RESP 值
func NewRESPValue(respType byte) *RESPValue {
	return &RESPValue{
		Type: respType,
	}
}

// ParseRESP 从 reader 解析 RESP 数据
func ParseRESP(reader *bufio.Reader) (*RESPValue, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil, fmt.Errorf("empty line")
	}

	value := NewRESPValue(line[0])

	switch line[0] {
	case RESP_SIMPLE_STRING:
		value.Str = line[1:]
		return value, nil

	case RESP_ERROR:
		value.Str = line[1:]
		return value, nil

	case RESP_INTEGER:
		num, err := strconv.ParseInt(line[1:], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %v", err)
		}
		value.Num = num
		return value, nil

	case RESP_BULK_STRING:
		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length: %v", err)
		}

		if length == -1 {
			value.IsNull = true
			return value, nil
		}

		// 读取指定长度的字符串
		data := make([]byte, length)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			return nil, fmt.Errorf("failed to read bulk string: %v", err)
		}

		value.Str = string(data)

		// 读取结尾的 \r\n
		_, err = reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read bulk string terminator: %v", err)
		}

		return value, nil

	case RESP_ARRAY:
		count, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid array length: %v", err)
		}

		if count == -1 {
			value.IsNull = true
			return value, nil
		}

		value.Array = make([]*RESPValue, count)
		for i := 0; i < count; i++ {
			elem, err := ParseRESP(reader)
			if err != nil {
				return nil, fmt.Errorf("failed to parse array element %d: %v", i, err)
			}
			value.Array[i] = elem
		}

		return value, nil

	default:
		return nil, fmt.Errorf("unknown RESP type: %c", line[0])
	}
}

// SerializeRESP 将 RESP 值序列化为字节数组
func (v *RESPValue) SerializeRESP() []byte {
	var buf bytes.Buffer

	switch v.Type {
	case RESP_SIMPLE_STRING:
		buf.WriteByte(RESP_SIMPLE_STRING)
		buf.WriteString(v.Str)
		buf.WriteString("\r\n")

	case RESP_ERROR:
		buf.WriteByte(RESP_ERROR)
		buf.WriteString(v.Str)
		buf.WriteString("\r\n")

	case RESP_INTEGER:
		buf.WriteByte(RESP_INTEGER)
		buf.WriteString(strconv.FormatInt(v.Num, 10))
		buf.WriteString("\r\n")

	case RESP_BULK_STRING:
		buf.WriteByte(RESP_BULK_STRING)
		if v.IsNull {
			buf.WriteString("-1\r\n")
		} else {
			buf.WriteString(strconv.Itoa(len(v.Str)))
			buf.WriteString("\r\n")
			buf.WriteString(v.Str)
			buf.WriteString("\r\n")
		}

	case RESP_ARRAY:
		buf.WriteByte(RESP_ARRAY)
		buf.WriteString(strconv.Itoa(len(v.Array)))
		buf.WriteString("\r\n")
		for _, elem := range v.Array {
			buf.Write(elem.SerializeRESP())
		}
	}

	return buf.Bytes()
}

// ToString 将 RESP 值转换为字符串（用于调试）
func (v *RESPValue) ToString() string {
	switch v.Type {
	case RESP_SIMPLE_STRING:
		return fmt.Sprintf("SimpleString: %s", v.Str)
	case RESP_ERROR:
		return fmt.Sprintf("Error: %s", v.Str)
	case RESP_INTEGER:
		return fmt.Sprintf("Integer: %d", v.Num)
	case RESP_BULK_STRING:
		if v.IsNull {
			return "BulkString: null"
		}
		return fmt.Sprintf("BulkString: %s", v.Str)
	case RESP_ARRAY:
		if v.IsNull {
			return "Array: null"
		}
		parts := make([]string, len(v.Array))
		for i, elem := range v.Array {
			parts[i] = elem.ToString()
		}
		return fmt.Sprintf("Array[%d]: [%s]", len(v.Array), strings.Join(parts, ", "))
	default:
		return "Unknown"
	}
}
