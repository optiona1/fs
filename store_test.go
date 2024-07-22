package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T) {
	original := "momsbestpicture"
	pathKey := PathKey{
		Pathname: "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff",
		Filename: "6804429f74181a63c50c3d81d733a12f14a353ff",
	}
	pathname := CASPathTransformFunc(original)
	assert.Equal(t, pathKey, pathname)
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte(fmt.Sprintf("some png bytes %d", i))

		if err := s.writeStream(key, bytes.NewBuffer(data)); err != nil {
			t.Error(err)
		}
		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, _ := io.ReadAll(r)

		fmt.Println(string(b))
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		if err = s.Delete(key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); ok {
			t.Errorf("expected to NOT have key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransfromFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Errorf("clear error: %s", err)
	}
}
