// Package ringbuffer はリングバッファを実装します。
package ringbuffer

import (
	"fmt"
)

// IRingBuffer はリングバッファの読み書きインターフェースです.
type IRingBuffer[T any] interface {
	Write(v T)
	Read() (T, bool)
}

// RingBuffer は固定長のリングバッファです.
type RingBuffer[T any] struct {
	buf   []T    // データを格納する領域
	write uint64 // 次にデータを書き込む位置
	read  uint64 // 次にデータを読み込む位置
	mask  uint64 // len(rb.buf) - 1 をしたくないため
}

// Write はリングバッファへ書き込みます。
func (rb *RingBuffer[T]) Write(v T) {
	// 満杯ならreadを進める
	if rb.write-rb.read == uint64(len(rb.buf)) {
		rb.read++
	}

	// データを書き込んで、writeを進める
	rb.buf[rb.write&rb.mask] = v
	rb.write++
}

// Read はリングバッファから読み込みます。
func (rb *RingBuffer[T]) Read() (T, bool) {
	var zero T

	// バッファが空
	if rb.write == rb.read {
		return zero, false
	}

	idx := rb.read & rb.mask
	ret := rb.buf[idx]
	rb.buf[idx] = zero

	//	ret := rb.buf[rb.read&rb.mask]
	rb.read++

	return ret, true
}

// NewRingBuffer は指定したサイズのリングバッファを作成します。
func NewRingBuffer[T any](size uint64) (*RingBuffer[T], error) {
	// 2のn乗であるかチェック
	if size > 0 && size&(size-1) == 0 {
		return &RingBuffer[T]{
			buf:  make([]T, size),
			mask: size - 1,
		}, nil
	}
	return nil, fmt.Errorf("2の累乗を指定してください")
}
