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
	var order binary.ByteOrder = binary.LittleEndian
	if data[5] == 2 {
		order = binary.BigEndian
	}
	shoff := order.Uint64(data[40:48])
	shnum := int(order.Uint16(data[60:62]))
	shstrndx := int(order.Uint16(data[62:64]))
	f := &File{}
	if shoff+64 > uint64(len(data)) {
		return f, nil
	}
	// 扩展节数：e_shnum 为 0 时真实节数在 0 号节头的 sh_size。
	if shnum == 0 {
		shnum = int(order.Uint64(data[shoff+32 : shoff+40]))
	}
	// 扩展名字表索引：e_shstrndx 为 SHN_XINDEX 时真实索引在 0 号节头的 sh_link。
	if shstrndx == 0xffff {
		shstrndx = int(order.Uint32(data[shoff+40 : shoff+44]))
	}
	if shnum == 0 || shoff+uint64(shnum)*64 > uint64(len(data)) {
		return f, nil
	}
	var strtab []byte
	if shstrndx < shnum {
		base := shoff + uint64(shstrndx)*64
		off := order.Uint64(data[base+24 : base+32])
		size := order.Uint64(data[base+32 : base+40])
		if off+size <= uint64(len(data)) {
			strtab = data[off : off+size]
		}
	}
	f.byName = make(map[string]*Section, shnum)
	for i := 0; i < shnum; i++ {
		c.Scanned++
		base := shoff + uint64(i)*64
		nameOff := int(order.Uint32(data[base : base+4]))
		name := ""
		if nameOff < len(strtab) {
			rest := strtab[nameOff:]
			end := len(rest)
			for j, b := range rest {
				if b == 0 {
					end = j
					break
				}
			}
			name = string(rest[:end])
		}
		f.Sections = append(f.Sections, Section{
			Name:   name,
			Type:   order.Uint32(data[base+4 : base+8]),
			Offset: order.Uint64(data[base+24 : base+32]),
			Size:   order.Uint64(data[base+32 : base+40]),
		})
	}
	for i := range f.Sections {
		if _, ok := f.byName[f.Sections[i].Name]; !ok {
			f.byName[f.Sections[i].Name] = &f.Sections[i]
		}
	}
	return f, nil
}

// Lookup 按名字找节。
func (f *File) Lookup(name string, c *Counter) *Section {
	c.Scanned++
	return f.byName[name]
}
