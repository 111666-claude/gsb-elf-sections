// Package elfsect 解析 ELF 节头表。
package elfsect

import (
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
	shoff := binary.LittleEndian.Uint64(data[40:48])
	shnum := int(binary.LittleEndian.Uint16(data[60:62]))
	shstrndx := int(binary.LittleEndian.Uint16(data[62:64]))
	f := &File{}
	if shnum == 0 || shoff+uint64(shnum)*64 > uint64(len(data)) {
		return f, nil
	}
	var strtab []byte
	if shstrndx < shnum {
		base := shoff + uint64(shstrndx)*64
		off := binary.LittleEndian.Uint64(data[base+24 : base+32])
		size := binary.LittleEndian.Uint64(data[base+32 : base+40])
		if off+size <= uint64(len(data)) {
			strtab = data[off : off+size]
		}
	}
	for i := 0; i < shnum; i++ {
		c.Scanned++
		base := shoff + uint64(i)*64
		nameOff := int(binary.LittleEndian.Uint32(data[base : base+4]))
		name := ""
		if nameOff < len(strtab) {
			name = string(strtab[nameOff:])
		}
		f.Sections = append(f.Sections, Section{
			Name:   name,
			Type:   binary.LittleEndian.Uint32(data[base+4 : base+8]),
			Offset: binary.LittleEndian.Uint64(data[base+24 : base+32]),
			Size:   binary.LittleEndian.Uint64(data[base+32 : base+40]),
		})
	}
	return f, nil
}

// Lookup 按名字找节。
func (f *File) Lookup(name string, c *Counter) *Section {
	for i := range f.Sections {
		c.Scanned++
		if f.Sections[i].Name == name {
			return &f.Sections[i]
		}
	}
	return nil
}
