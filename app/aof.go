package main

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"time"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   *sync.Mutex
}

func NewAof(path string) (aof *Aof, err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof = &Aof{
		file: f,
		rd:   bufio.NewReader(f),
		mu:   &sync.Mutex{},
	}

	go func() {
		for {
			aof.mu.Lock()
			aof.file.Sync()
			aof.mu.Unlock()
			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

func (aof *Aof) Write(value Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Read() {
	resp := NewResp(aof.rd)

	for {
		value, err := resp.Read()
		if err != nil {
			return
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler := Handlers[command]

		handler(args)
	}

}
