package elfsect

import (
	"encoding/binary"
	"fmt"
)

const (
	shtProgbits = 1
	shtStrtab   = 3
)

// Build 拼一个 64 位 ELF，节为：0 号空节 + names + .shstrtab。
func Build(order binary.ByteOrder, names []string, extended, xindex bool) []byte {
	strtab := []byte{0}
	offsets := make([]int, len(names))
	for i, n := range names {
		offsets[i] = len(strtab)
		strtab = append(strtab, []byte(n)...)
		strtab = append(strtab, 0)
	}
	shstrName := len(strtab)
	strtab = append(strtab, []byte(".shstrtab")...)
	strtab = append(strtab, 0)

	shnum := len(names) + 2
	dataStart := 64
	dataLen := 8 * len(names)
	shstrOff := dataStart + dataLen
	shoff := shstrOff + len(strtab)
	out := make([]byte, shoff+shnum*64)
	copy(out, []byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0})
	if order == binary.BigEndian {
		out[5] = 2
	}
	put16 := func(off int, v uint16) { order.PutUint16(out[off:], v) }
	put32 := func(off int, v uint32) { order.PutUint32(out[off:], v) }
	put64 := func(off int, v uint64) { order.PutUint64(out[off:], v) }
	put16(16, 2)
	put16(18, 0x3e)
	put32(20, 1)
	put64(24, 0x400000)
	put64(40, uint64(shoff))
	put16(52, 64)
	put16(58, 64)
	put16(60, uint16(shnum))
	put16(62, uint16(shnum-1))
	if extended {
		put16(60, 0)
	}
	if xindex {
		put16(62, 0xffff)
	}
	for i := range names {
		base := dataStart + i*8
		for k := 0; k < 8; k++ {
			out[base+k] = byte(i + 1)
		}
	}
	copy(out[shstrOff:], strtab)
	// 0 号节头
	put64(shoff+32, 0)
	if extended {
		put64(shoff+32, uint64(shnum))
	}
	if xindex {
		put32(shoff+40, uint32(shnum-1))
	}
	for i := range names {
		base := shoff + (i+1)*64
		put32(base, uint32(offsets[i]))
		put32(base+4, shtProgbits)
		put64(base+24, uint64(dataStart+i*8))
		put64(base+32, 8)
	}
	base := shoff + (len(names)+1)*64
	put32(base, uint32(shstrName))
	put32(base+4, shtStrtab)
	put64(base+24, uint64(shstrOff))
	put64(base+32, uint64(len(strtab)))
	return out
}

// Scenario 是一个固定场景。
type Scenario struct {
	Name string
	Run  func() map[string]any
}

// Samples 返回全部场景。
func Samples() []Scenario {
	return []Scenario{
		{Name: "extnum", Run: func() map[string]any {
			data := Build(binary.LittleEndian, []string{".text"}, true, false)
			f, _ := Read(data, &Counter{})
			return map[string]any{"sections": len(f.Sections)}
		}},
		{Name: "xindex", Run: func() map[string]any {
			data := Build(binary.LittleEndian, []string{".text"}, false, true)
			f, _ := Read(data, &Counter{})
			sec := f.Lookup(".text", &Counter{})
			if sec == nil {
				return map[string]any{"namelen": 0}
			}
			return map[string]any{"namelen": len(sec.Name)}
		}},
		{Name: "namelen", Run: func() map[string]any {
			data := Build(binary.LittleEndian, []string{".text", ".data"}, false, false)
			f, _ := Read(data, &Counter{})
			if len(f.Sections) < 2 {
				return map[string]any{"namelen": 0}
			}
			return map[string]any{"namelen": len(f.Sections[1].Name)}
		}},
		{Name: "be", Run: func() map[string]any {
			data := Build(binary.BigEndian, []string{".text"}, false, false)
			f, _ := Read(data, &Counter{})
			return map[string]any{"sections": len(f.Sections)}
		}},
	}
}

// Work 跑规模线场景。
func Work(n int) map[string]any {
	names := make([]string, n)
	for i := range names {
		names[i] = fmt.Sprintf(".s%04d", i)
	}
	data := Build(binary.LittleEndian, names, false, false)
	f, _ := Read(data, &Counter{})
	c := &Counter{}
	for i := 0; i < n; i++ {
		f.Lookup(fmt.Sprintf(".s%04d", i), c)
	}
	return map[string]any{"sections": n, "scanned": c.Scanned}
}
