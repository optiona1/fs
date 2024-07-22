package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const DefalutRootFolder = "ggnetwork"

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathTransfromFunc func(string) PathKey

type PathKey struct {
	Root     string
	Pathname string
	Filename string
}

func (p PathKey) FullPathWithRoot() string {
	return fmt.Sprintf("%s/%s/%s", p.Root, p.Pathname, p.Filename)
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

type StoreOpts struct {
	// Root is the folder name of the root, containing all the folders/files of the system.
	Root              string
	PathTransfromFunc PathTransfromFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		Pathname: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransfromFunc == nil {
		opts.PathTransfromFunc = DefaultPathTransformFunc
	}
	if opts.Root == "" {
		opts.Root = DefalutRootFolder
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	pathKey := CASPathTransformFunc(key)
	pathKey.Root = s.Root
	fullPathWithRoot := pathKey.FullPathWithRoot()
	_, err := os.Stat(fullPathWithRoot)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func (s *Store) Clear() error {
	return os.Remove(s.Root)
}

func (s *Store) Delete(key string) error {
	pathKey := CASPathTransformFunc(key)
	pathKey.Root = s.Root
	fullPathWithRoot := pathKey.FullPathWithRoot()

	defer func() {
		log.Printf("deleted [%s] from disk.", fullPathWithRoot)
	}()

	err := os.Remove(fullPathWithRoot)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	parentDir := filepath.Dir(fullPathWithRoot)
	for {
		err := os.Remove(parentDir)
		if err != nil {
			if !os.IsNotExist(err) && !os.IsPermission(err) && !isNotEmptyError(err) {
				return fmt.Errorf("failed to delete directory: %v", err)
			}
			break
		}

		parentDir = filepath.Dir(parentDir)

		if parentDir == "." || parentDir == s.Root || parentDir == "/" {
			break
		}
	}
	return nil
}

func isNotEmptyError(err error) bool {
	return os.IsExist(err)
}

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}

func (s *Store) Read(key string) (io.Reader, error) {

	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransfromFunc(key)
	pathKey.Root = s.Root

	return os.Open(pathKey.FullPathWithRoot())
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransfromFunc(key)
	pathKey.Root = s.Root

	if err := os.MkdirAll(s.Root+"/"+pathKey.Pathname, os.ModePerm); err != nil {
		return err
	}

	fullPathWithRoot := pathKey.FullPathWithRoot()

	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk, %s", n, fullPathWithRoot)

	return nil
}
