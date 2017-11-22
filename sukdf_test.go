package sukdf

import (
	"testing"
	"math/rand"
	"golang.org/x/crypto/scrypt"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func BenchmarkSukdf_Compute(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		str := RandStringBytesMaskImprSrc(12)
		t.Logf("%s \n", str)
		s := New(str, WithMaxBacktracks(15000))
		x, ok := s.Compute()
		if !ok {
			t.Log("hash compute maxed on iterations")
		}
		t.Logf("%d \n", len(x))
		t.Logf("%x", x)
	}
}

func TestSukdf_Compute(t *testing.T) {
	s := New("dnkfdkf")
	x, ok := s.Compute()
	if !ok {
		t.Fatal("hash compute failed")
	}
	t.Logf("%d \n", len(x))
	t.Logf("%x", x)
}

func BenchmarkScrypt(b *testing.B) {
	salt, _ := generateRandomBytes(16)
	dk, _ := scrypt.Key([]byte("fdsv82hfmv"), salt, 16384, 8, 1, 32)
	b.Logf("%x", dk)
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// err == nil only if len(b) == n
	if err != nil {
		return nil, err
	}

	return b, nil
}
