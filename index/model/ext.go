package model

import (
	"fmt"
	"sort"
	"strconv"
)

func identFromStrings(strings ...string) []byte {
	buf := make([]byte, 0)
	for _, str := range strings {
		buf = append(buf, []byte(str)...)
	}
	return buf
}

func identMapToString(data map[string]string) string {
	keys := make([]string, len(data))
	i := 0
	for k, _ := range data {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	buf := make([]byte, 0)
	for _, k := range keys {
		buf = append(buf, []byte(k)...)
		buf = append(buf, []byte(data[k])...)
	}
	return string(buf)
}

func (m *Entry) Identity() []byte {
	return identFromStrings(
		m.GetName(),
		m.GetAddress(),
		fmt.Sprintf("%v", m.Type),
		strconv.FormatInt(m.GetTimestamp(), 10),
		identMapToString(m.GetMetadata()))
}

func (m *Blob) Identity() []byte {
	return identFromStrings(m.Blocks...)
}

func (m *Commit) Identity() []byte {
	return append(identFromStrings(
		m.GetTree(),
		m.GetCommitter(),
		m.GetMessage(),
		strconv.FormatInt(m.GetTimestamp(), 10),
		identMapToString(m.GetMetadata()),
	), identFromStrings(m.GetParents()...)...)
}

func (m *Object) Identity() []byte {
	return append(
		m.GetBlob().Identity(),
		identFromStrings(
			identMapToString(m.GetMetadata()),
		)...,
	)
}