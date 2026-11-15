package ringbuffer

import (
	"math/rand"
	"strconv"
	"testing"
)

var benchmarkResult int

// User は構造体を格納するベンチマーク用のデータです。
type User struct {
	Name string
	Age  int
	Data []byte
}

func BenchmarkWrite(b *testing.B) {
	rb, err := NewRingBuffer[int](1024)
	if err != nil {
		b.Fatal(err)
	}

	var sum int

	for b.Loop() {
		rb.Write(100)
		v, _ := rb.Read()
		sum += v
	}

	benchmarkResult = sum

}

func BenchmarkBatch(b *testing.B) {
	rb, err := NewRingBuffer[int](1024)
	if err != nil {
		b.Fatal(err)
	}

	var sum int

	for b.Loop() {
		for i := 0; i < 100; i++ {
			rb.Write(i)
		}

		for i := 0; i < 100; i++ {
			v, _ := rb.Read()
			sum += v
		}
	}

	benchmarkResult = sum
}

func BenchmarkStringBatch(b *testing.B) {
	rb, err := NewRingBuffer[User](1024)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		for i := 0; i < 100; i++ {
			rb.Write(User{Name: strconv.Itoa(i), Age: i, Data: []byte{0, 0, 0, 0, 0, 0, 0}})
		}

		for i := 0; i < 100; i++ {
			rb.Read()
		}
	}

}

func TestRingBufferOverwrite(t *testing.T) {
	rb, err := NewRingBuffer[int](4)
	if err != nil {
		t.Fatal(err)
	}

	// 容量4に対して6件書き込む
	for _, v := range []int{10, 20, 30, 40, 50, 60} {
		rb.Write(v)
	}

	// 古い2件は上書きされているはず
	want := []int{30, 40, 50, 60}

	for _, expected := range want {
		got, ok := rb.Read()

		if !ok {
			t.Fatal("データが存在するはずなのにReadが失敗しました")
		}

		if got != expected {
			t.Fatalf("got=%d, want=%d", got, expected)
		}
	}

	// すべて読み出した後は空
	_, ok := rb.Read()
	if ok {
		t.Fatal("空のバッファからデータを読み出せました")
	}
}

func TestRingBufferCapacityOne(t *testing.T) {
	rb, err := NewRingBuffer[int](1)
	if err != nil {
		t.Fatal(err)
	}

	rb.Write(10)
	rb.Write(20)
	rb.Write(30)

	got, ok := rb.Read()
	if !ok {
		t.Fatal("Readに失敗しました")
	}

	if got != 30 {
		t.Fatalf("got=%d, want=30", got)
	}

	// 2回目は空のはず
	_, ok = rb.Read()
	if ok {
		t.Fatal("空のバッファから読み出せました")
	}
}

func TestRingBufferCounterOverflow(t *testing.T) {
	rb, err := NewRingBuffer[int](4)
	if err != nil {
		t.Fatal(err)
	}

	// カウンタを最大値の直前に設定
	rb.write = ^uint64(0) - 1
	rb.read = rb.write

	rb.Write(10)
	rb.Write(20)
	rb.Write(30)

	want := []int{10, 20, 30}

	for _, expected := range want {
		got, ok := rb.Read()
		if !ok {
			t.Fatal("Readに失敗しました")
		}

		if got != expected {
			t.Fatalf("got=%d, want=%d", got, expected)
		}
	}

	if _, ok := rb.Read(); ok {
		t.Fatal("空のはずなのに読み出せました")
	}
}

func TestRingBufferInterleaved(t *testing.T) {
	rb, err := NewRingBuffer[int](4)
	if err != nil {
		t.Fatal(err)
	}

	checkRead := func(want int) {
		t.Helper()

		got, ok := rb.Read()
		if !ok {
			t.Fatalf("Read() failed, want=%d", want)
		}
		if got != want {
			t.Fatalf("got=%d, want=%d", got, want)
		}
	}

	rb.Write(10)
	rb.Write(20)
	checkRead(10)

	rb.Write(30)
	rb.Write(40)
	checkRead(20)

	rb.Write(50)
	checkRead(30)
	checkRead(40)
	checkRead(50)

	// 全件読み出したので空のはず
	if _, ok := rb.Read(); ok {
		t.Fatal("buffer should be empty")
	}
}

func TestRingBufferRandom(t *testing.T) {
	const capacity = 8

	rb, err := NewRingBuffer[int](capacity)
	if err != nil {
		t.Fatal(err)
	}

	// 正解モデル
	var model []int

	// 再現可能な乱数
	rng := rand.New(rand.NewSource(42))

	for i := range 100_000 {
		if rng.Intn(2) == 0 {
			// Write
			value := rng.Intn(1000)
			rb.Write(value)

			if len(model) == capacity {
				model = model[1:]
			}
			model = append(model, value)
		} else {
			// Read
			got, ok := rb.Read()

			if len(model) == 0 {
				if ok {
					t.Fatalf("step=%d: expected empty", i)
				}
				continue
			}

			want := model[0]
			model = model[1:]

			if !ok || got != want {
				t.Fatalf(
					"step=%d: got=(%d,%v), want=%d",
					i, got, ok, want,
				)
			}
		}
	}
}

func FuzzRingBuffer(f *testing.F) {
	// 最初に試す入力
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{0, 0, 0, 0})
	f.Add([]byte{255, 128, 64, 32})
	// f.Add([]byte{2, 4, 6, 8, 10, 12, 14, 16, 18, 1})

	f.Fuzz(func(t *testing.T, data []byte) {
		const capacity = 8

		rb, err := NewRingBuffer[int](capacity)
		if err != nil {
			t.Fatal(err)
		}

		var model []int

		for i, b := range data {
			if b&1 == 0 {
				// 偶数ならWrite
				value := int(b)
				rb.Write(value)

				if len(model) == capacity {
					model = model[1:]
				}
				model = append(model, value)
			} else {
				// 奇数ならRead
				got, ok := rb.Read()

				if len(model) == 0 {
					if ok {
						t.Fatalf("step=%d: expected empty", i)
					}
					continue
				}

				want := model[0]
				model = model[1:]

				if !ok || got != want {
					t.Fatalf(
						"step=%d: got=(%d,%v), want=%d",
						i, got, ok, want,
					)
				}
			}
		}
	})
}
