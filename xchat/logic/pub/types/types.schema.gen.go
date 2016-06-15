package types

import (
	"io"
	"time"
	"unsafe"
)

var (
	_ = unsafe.Sizeof(0)
	_ = io.ReadFull
	_ = time.Now()
)

type ChatMessage struct {
	ChatID   uint64
	ChatType string
	ID       uint64
	User     string
	Ts       int64
	Msg      string
	Updated  int64
}

func (d *ChatMessage) Size() (s uint64) {

	{
		l := uint64(len(d.ChatType))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.User))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.Msg))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 32
	return
}
func (d *ChatMessage) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.ChatID >> 0)

		buf[1+0] = byte(d.ChatID >> 8)

		buf[2+0] = byte(d.ChatID >> 16)

		buf[3+0] = byte(d.ChatID >> 24)

		buf[4+0] = byte(d.ChatID >> 32)

		buf[5+0] = byte(d.ChatID >> 40)

		buf[6+0] = byte(d.ChatID >> 48)

		buf[7+0] = byte(d.ChatID >> 56)

	}
	{
		l := uint64(len(d.ChatType))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+8] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+8] = byte(t)
			i++

		}
		copy(buf[i+8:], d.ChatType)
		i += l
	}
	{

		buf[i+0+8] = byte(d.ID >> 0)

		buf[i+1+8] = byte(d.ID >> 8)

		buf[i+2+8] = byte(d.ID >> 16)

		buf[i+3+8] = byte(d.ID >> 24)

		buf[i+4+8] = byte(d.ID >> 32)

		buf[i+5+8] = byte(d.ID >> 40)

		buf[i+6+8] = byte(d.ID >> 48)

		buf[i+7+8] = byte(d.ID >> 56)

	}
	{
		l := uint64(len(d.User))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+16] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+16] = byte(t)
			i++

		}
		copy(buf[i+16:], d.User)
		i += l
	}
	{

		buf[i+0+16] = byte(d.Ts >> 0)

		buf[i+1+16] = byte(d.Ts >> 8)

		buf[i+2+16] = byte(d.Ts >> 16)

		buf[i+3+16] = byte(d.Ts >> 24)

		buf[i+4+16] = byte(d.Ts >> 32)

		buf[i+5+16] = byte(d.Ts >> 40)

		buf[i+6+16] = byte(d.Ts >> 48)

		buf[i+7+16] = byte(d.Ts >> 56)

	}
	{
		l := uint64(len(d.Msg))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+24] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+24] = byte(t)
			i++

		}
		copy(buf[i+24:], d.Msg)
		i += l
	}
	{

		buf[i+0+24] = byte(d.Updated >> 0)

		buf[i+1+24] = byte(d.Updated >> 8)

		buf[i+2+24] = byte(d.Updated >> 16)

		buf[i+3+24] = byte(d.Updated >> 24)

		buf[i+4+24] = byte(d.Updated >> 32)

		buf[i+5+24] = byte(d.Updated >> 40)

		buf[i+6+24] = byte(d.Updated >> 48)

		buf[i+7+24] = byte(d.Updated >> 56)

	}
	return buf[:i+32], nil
}

func (d *ChatMessage) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.ChatID = 0 | (uint64(buf[i+0+0]) << 0) | (uint64(buf[i+1+0]) << 8) | (uint64(buf[i+2+0]) << 16) | (uint64(buf[i+3+0]) << 24) | (uint64(buf[i+4+0]) << 32) | (uint64(buf[i+5+0]) << 40) | (uint64(buf[i+6+0]) << 48) | (uint64(buf[i+7+0]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+8] & 0x7F)
			for buf[i+8]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+8]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.ChatType = string(buf[i+8 : i+8+l])
		i += l
	}
	{

		d.ID = 0 | (uint64(buf[i+0+8]) << 0) | (uint64(buf[i+1+8]) << 8) | (uint64(buf[i+2+8]) << 16) | (uint64(buf[i+3+8]) << 24) | (uint64(buf[i+4+8]) << 32) | (uint64(buf[i+5+8]) << 40) | (uint64(buf[i+6+8]) << 48) | (uint64(buf[i+7+8]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+16] & 0x7F)
			for buf[i+16]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+16]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.User = string(buf[i+16 : i+16+l])
		i += l
	}
	{

		d.Ts = 0 | (int64(buf[i+0+16]) << 0) | (int64(buf[i+1+16]) << 8) | (int64(buf[i+2+16]) << 16) | (int64(buf[i+3+16]) << 24) | (int64(buf[i+4+16]) << 32) | (int64(buf[i+5+16]) << 40) | (int64(buf[i+6+16]) << 48) | (int64(buf[i+7+16]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+24] & 0x7F)
			for buf[i+24]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+24]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.Msg = string(buf[i+24 : i+24+l])
		i += l
	}
	{

		d.Updated = 0 | (int64(buf[i+0+24]) << 0) | (int64(buf[i+1+24]) << 8) | (int64(buf[i+2+24]) << 16) | (int64(buf[i+3+24]) << 24) | (int64(buf[i+4+24]) << 32) | (int64(buf[i+5+24]) << 40) | (int64(buf[i+6+24]) << 48) | (int64(buf[i+7+24]) << 56)

	}
	return i + 32, nil
}

type ChatNotifyMessage struct {
	ChatID   uint64
	ChatType string
	User     string
	Ts       int64
	Msg      string
	Updated  int64
}

func (d *ChatNotifyMessage) Size() (s uint64) {

	{
		l := uint64(len(d.ChatType))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.User))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.Msg))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 24
	return
}
func (d *ChatNotifyMessage) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.ChatID >> 0)

		buf[1+0] = byte(d.ChatID >> 8)

		buf[2+0] = byte(d.ChatID >> 16)

		buf[3+0] = byte(d.ChatID >> 24)

		buf[4+0] = byte(d.ChatID >> 32)

		buf[5+0] = byte(d.ChatID >> 40)

		buf[6+0] = byte(d.ChatID >> 48)

		buf[7+0] = byte(d.ChatID >> 56)

	}
	{
		l := uint64(len(d.ChatType))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+8] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+8] = byte(t)
			i++

		}
		copy(buf[i+8:], d.ChatType)
		i += l
	}
	{
		l := uint64(len(d.User))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+8] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+8] = byte(t)
			i++

		}
		copy(buf[i+8:], d.User)
		i += l
	}
	{

		buf[i+0+8] = byte(d.Ts >> 0)

		buf[i+1+8] = byte(d.Ts >> 8)

		buf[i+2+8] = byte(d.Ts >> 16)

		buf[i+3+8] = byte(d.Ts >> 24)

		buf[i+4+8] = byte(d.Ts >> 32)

		buf[i+5+8] = byte(d.Ts >> 40)

		buf[i+6+8] = byte(d.Ts >> 48)

		buf[i+7+8] = byte(d.Ts >> 56)

	}
	{
		l := uint64(len(d.Msg))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+16] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+16] = byte(t)
			i++

		}
		copy(buf[i+16:], d.Msg)
		i += l
	}
	{

		buf[i+0+16] = byte(d.Updated >> 0)

		buf[i+1+16] = byte(d.Updated >> 8)

		buf[i+2+16] = byte(d.Updated >> 16)

		buf[i+3+16] = byte(d.Updated >> 24)

		buf[i+4+16] = byte(d.Updated >> 32)

		buf[i+5+16] = byte(d.Updated >> 40)

		buf[i+6+16] = byte(d.Updated >> 48)

		buf[i+7+16] = byte(d.Updated >> 56)

	}
	return buf[:i+24], nil
}

func (d *ChatNotifyMessage) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.ChatID = 0 | (uint64(buf[i+0+0]) << 0) | (uint64(buf[i+1+0]) << 8) | (uint64(buf[i+2+0]) << 16) | (uint64(buf[i+3+0]) << 24) | (uint64(buf[i+4+0]) << 32) | (uint64(buf[i+5+0]) << 40) | (uint64(buf[i+6+0]) << 48) | (uint64(buf[i+7+0]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+8] & 0x7F)
			for buf[i+8]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+8]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.ChatType = string(buf[i+8 : i+8+l])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+8] & 0x7F)
			for buf[i+8]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+8]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.User = string(buf[i+8 : i+8+l])
		i += l
	}
	{

		d.Ts = 0 | (int64(buf[i+0+8]) << 0) | (int64(buf[i+1+8]) << 8) | (int64(buf[i+2+8]) << 16) | (int64(buf[i+3+8]) << 24) | (int64(buf[i+4+8]) << 32) | (int64(buf[i+5+8]) << 40) | (int64(buf[i+6+8]) << 48) | (int64(buf[i+7+8]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+16] & 0x7F)
			for buf[i+16]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+16]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.Msg = string(buf[i+16 : i+16+l])
		i += l
	}
	{

		d.Updated = 0 | (int64(buf[i+0+16]) << 0) | (int64(buf[i+1+16]) << 8) | (int64(buf[i+2+16]) << 16) | (int64(buf[i+3+16]) << 24) | (int64(buf[i+4+16]) << 32) | (int64(buf[i+5+16]) << 40) | (int64(buf[i+6+16]) << 48) | (int64(buf[i+7+16]) << 56)

	}
	return i + 24, nil
}

type XMessage struct {
	Msg interface{}
}

func (d *XMessage) Size() (s uint64) {

	{
		var v uint64
		switch d.Msg.(type) {

		case ChatMessage:
			v = 0 + 1

		case ChatNotifyMessage:
			v = 1 + 1

		}

		{

			t := v
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		switch tt := d.Msg.(type) {

		case ChatMessage:

			{
				s += tt.Size()
			}

		case ChatNotifyMessage:

			{
				s += tt.Size()
			}

		}
	}
	return
}
func (d *XMessage) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{
		var v uint64
		switch d.Msg.(type) {

		case ChatMessage:
			v = 0 + 1

		case ChatNotifyMessage:
			v = 1 + 1

		}

		{

			t := uint64(v)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		switch tt := d.Msg.(type) {

		case ChatMessage:

			{
				nbuf, err := tt.Marshal(buf[i+0:])
				if err != nil {
					return nil, err
				}
				i += uint64(len(nbuf))
			}

		case ChatNotifyMessage:

			{
				nbuf, err := tt.Marshal(buf[i+0:])
				if err != nil {
					return nil, err
				}
				i += uint64(len(nbuf))
			}

		}
	}
	return buf[:i+0], nil
}

func (d *XMessage) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{
		v := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			v = t

		}
		switch v {

		case 0 + 1:
			var tt ChatMessage

			{
				ni, err := tt.Unmarshal(buf[i+0:])
				if err != nil {
					return 0, err
				}
				i += ni
			}

			d.Msg = tt

		case 1 + 1:
			var tt ChatNotifyMessage

			{
				ni, err := tt.Unmarshal(buf[i+0:])
				if err != nil {
					return 0, err
				}
				i += ni
			}

			d.Msg = tt

		default:
			d.Msg = nil
		}
	}
	return i + 0, nil
}
