// Package elfsect 解析 ELF 节头表。
package elfsect

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Counter 记录比较次数。
type Counter struct {
	Scanned int
}

// Section 是一个节。
type Section struct {
	Name   string `json:"name"`
	Type   uint32 `json:"type"`
	Offset uint64 `json:"offset"`
	Size   uint64 `json:"size"`
}

// File 是解析结果。
type File struct {
	Sections []Section
	byName   map[string]*Section
}

// Read 解析节头表。
func Read(data []byte, c *Counter) (*File, error) {
	if len(data) < 64 || string(data[0:4]) != "\x7fELF" {
		return nil, errors.New("bad-magic")
	}
	var order binary.ByteOrder
	switch data[5] {
	case 1:
		order = binary.LittleEndian
	case 2:
		order = binary.BigEndian
	default:
		return nil, errors.New("bad-encoding")
	}
	shoff := order.Uint64(data[40:48])
	shnum := uint64(order.Uint16(data[60:62]))
	shstrndx := uint64(order.Uint16(data[62:64]))
	f := &File{byName: make(map[string]*Section)}
	if shoff == 0 {
		return f, nil
	}
	// 扩展节数与扩展名字表索引都藏在 0 号节头里。
	if shnum == 0 || shstrndx == 0xffff {
		if shoff+64 > uint64(len(data)) {
			return f, nil
		}
		if shnum == 0 {
			shnum = order.Uint64(data[shoff+32 : shoff+40])
		}
		if shstrndx == 0xffff {
			shstrndx = uint64(order.Uint32(data[shoff+40 : shoff+44]))
		}
	}
	if shnum == 0 || shoff+shnum*64 > uint64(len(data)) {
		return f, nil
	}
	var strtab []byte
	if shstrndx < shnum {
		base := shoff + shstrndx*64
		off := order.Uint64(data[base+24 : base+32])
		size := order.Uint64(data[base+32 : base+40])
		if off+size <= uint64(len(data)) {
			strtab = data[off : off+size]
		}
	}
	for i := uint64(0); i < shnum; i++ {
		c.Scanned++
		base := shoff + i*64
		nameOff := int(order.Uint32(data[base : base+4]))
		name := ""
		if nameOff < len(strtab) {
			rest := strtab[nameOff:]
			if end := bytes.IndexByte(rest, 0); end >= 0 {
				rest = rest[:end]
			}
			name = string(rest)
		}
		f.Sections = append(f.Sections, Section{
			Name:   name,
			Type:   order.Uint32(data[base+4 : base+8]),
			Offset: order.Uint64(data[base+24 : base+32]),
			Size:   order.Uint64(data[base+32 : base+40]),
		})
	}
	for i := range f.Sections {
		f.byName[f.Sections[i].Name] = &f.Sections[i]
	}
	return f, nil
}

// Lookup 按名字找节。
func (f *File) Lookup(name string, c *Counter) *Section {
	c.Scanned++
	return f.byName[name]
}
