package main

import (
	"bytes"
	"encoding/gob"
	"time"
)

func encode(gl GameLog) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(gl)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decode(data []byte) (GameLog, error) {
	var gl GameLog
	// 将字节切片包装成 Reader 供解码器使用
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&gl)
	if err != nil {
		return GameLog{}, err
	}
	return gl, nil
}

// don't touch below this line

type GameLog struct {
	CurrentTime time.Time
	Message     string
	Username    string
}
