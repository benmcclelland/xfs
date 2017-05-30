# xfs
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/benmcclelland/xfs)

golang package for interfacing with xfs filesystem features

Example Bulk Stat:
```go
h, _ := xfs.NewBulkReq(".")
defer h.Release()

count := int64(0)
for {
	bstats, _ := h.Next()
	if len(bstats) == 0 {
        fmt.Println("Walked", count, "inodes")
		return
	}

	for _, bstat := range bstats {
		count++
		if bstat.Ino == 100 {
			fmt.Println("Found inode 100!")
		}
	}
}
```