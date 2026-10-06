package elfsect

import (
	"encoding/binary"
	"testing"
)

func TestSectionCount(t *testing.T) {
	data := Build(binary.LittleEndian, []string{".text"}, false, false)
	f, _ := Read(data, &Counter{})
	if len(f.Sections) != 3 {
		t.Fatalf("应有 3 个节：%+v", f.Sections)
	}
}

func TestSectionType(t *testing.T) {
	data := Build(binary.LittleEndian, []string{".text"}, false, false)
	f, _ := Read(data, &Counter{})
	if f.Sections[1].Type != shtProgbits {
		t.Fatal("类型应是 PROGBITS")
	}
}

func TestShstrtabType(t *testing.T) {
	data := Build(binary.LittleEndian, []string{".text"}, false, false)
	f, _ := Read(data, &Counter{})
	if f.Sections[2].Type != shtStrtab {
		t.Fatal("最后一个节应是 STRTAB")
	}
}

func TestSectionOffset(t *testing.T) {
	data := Build(binary.LittleEndian, []string{".text"}, false, false)
	f, _ := Read(data, &Counter{})
	if f.Sections[1].Offset != 64 {
		t.Fatalf("首个命名节偏移应是 64：%+v", f.Sections[1])
	}
}

func TestBadMagic(t *testing.T) {
	if _, err := Read([]byte("xxxx"), &Counter{}); err == nil {
		t.Fatal("头不对应报错")
	}
}

func TestCounterCounts(t *testing.T) {
	data := Build(binary.LittleEndian, []string{".text"}, false, false)
	c := &Counter{}
	Read(data, c)
	if c.Scanned == 0 {
		t.Fatal("应统计比较次数")
	}
}
