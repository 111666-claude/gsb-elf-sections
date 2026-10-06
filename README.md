# elf-sections

解析 ELF 的节头表与节名字符串表，并按名字建立索引。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/elfsect --sample=extnum
go run ./cmd/elfsect --sample=xindex
go run ./cmd/elfsect --sample=namelen
go run ./cmd/elfsect --sample=be
go run ./cmd/elfsect --sample=work
```

## 口径

- **字节序**：`e_ident` 里的数据编码位决定整份文件的字节序，后续多字节字段都按它读。
- **节数**：`e_shnum` 为 0 时，真实节数在 0 号节头的 `sh_size` 里。
- **名字表**：`e_shstrndx` 为 `SHN_XINDEX`（0xffff）时，真实索引在 0 号节头的 `sh_link` 里。
- **节名**：`sh_name` 是名字表里的字节偏移，名字到第一个零字节为止。
- **查找**：按名字查找走索引，不逐节比较。

## 不变量

- 节数与文件里声明的节数一致。
- 每个节名不含零字节。
- `scanned` 不随节数乘查询次数放大：八百个节查八百次的 `scanned` 不超过 6000。

## 输出契约

场景打印一行 JSON，键随场景而定。`--sample=work` 打印 `{"sections": N, "scanned": N}`。
