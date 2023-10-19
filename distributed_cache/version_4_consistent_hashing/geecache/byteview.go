package geecache

type ByteView struct {
    b []byte
}

func (byteView ByteView) Len() int {
    return len(byteView.b)
}

func (byteView ByteView) ByteSlice() []byte {
    return cloneBytes(byteView.b)
}

func cloneBytes(b []byte) []byte {
    c := make([]byte, len(b))
    copy(c, b)
    return c
}

func (byteView ByteView) String() string {
    return string(byteView.b)
}