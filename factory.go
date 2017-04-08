package main

import (
	"math/rand"
	"net/url"
	"time"
)

// Generator ...
type Generator func(n int) string

// DefaultGenerator ...
var DefaultGenerator = randomString

// Factory ...
type Factory struct {
	store     Store
	generator Generator
}

// NewFactory ...
func NewFactory(generator Generator, store Store) *Factory {
	return &Factory{
		store:     store,
		generator: generator,
	}
}

// Gen ...
func (f *Factory) Gen(uri string) (key string, err error) {
	// we don't return the parsed url because #hash are converted to uri-compatible
	// and we don't want to encode/decode all the time, there is no need for that,
	// we save the url as the user expects if the uri validation passed.
	_, err = url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}

	key = f.generator(8)

	for {
		if v := f.store.Get(key); v == "" {
			break
		}

		key = f.generator(8)
	}

	return key, nil
}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randomString(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
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
