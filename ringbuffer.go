// package main リングバッファを実装
package main

import (
	"fmt"
)

type IRingBuffer[T any] interface {
	Write(v T)
	Read() (T, bool)
}

type RingBuffer[T any] struct {
	buf   []T    // データを格納する領域
	write uint64 // 次にデータを書き込む位置
	read  uint64 // 次にデータを読み込む位置
	mask  uint64 // len(rb.buf) - 1 をしたくないため
}

// リングバッファへ書き込みます。
func (rb *RingBuffer[T]) Write(v T) {
	// 満杯ならreadを進める
	if rb.write-rb.read == uint64(len(rb.buf)) {
		rb.read++
	}

	// データを書き込んで、writeを進める
	rb.buf[rb.write&rb.mask] = v
	rb.write++
}

// リングバッファから読み込みます。
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

// リングバッファを作成します。
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
