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

type Message struct {
	ChatID uint64
	MsgID  uint64
	User   string
	Ts     int64
	Msg    string
}

func (d *Message) Size() (s uint64) {

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
func (d *Message) Marshal(buf []byte) ([]byte, error) {
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

		buf[0+8] = byte(d.MsgID >> 0)

		buf[1+8] = byte(d.MsgID >> 8)

		buf[2+8] = byte(d.MsgID >> 16)

		buf[3+8] = byte(d.MsgID >> 24)

		buf[4+8] = byte(d.MsgID >> 32)

		buf[5+8] = byte(d.MsgID >> 40)

		buf[6+8] = byte(d.MsgID >> 48)

		buf[7+8] = byte(d.MsgID >> 56)

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
	return buf[:i+24], nil
}

func (d *Message) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.ChatID = 0 | (uint64(buf[i+0+0]) << 0) | (uint64(buf[i+1+0]) << 8) | (uint64(buf[i+2+0]) << 16) | (uint64(buf[i+3+0]) << 24) | (uint64(buf[i+4+0]) << 32) | (uint64(buf[i+5+0]) << 40) | (uint64(buf[i+6+0]) << 48) | (uint64(buf[i+7+0]) << 56)

	}
	{

		d.MsgID = 0 | (uint64(buf[i+0+8]) << 0) | (uint64(buf[i+1+8]) << 8) | (uint64(buf[i+2+8]) << 16) | (uint64(buf[i+3+8]) << 24) | (uint64(buf[i+4+8]) << 32) | (uint64(buf[i+5+8]) << 40) | (uint64(buf[i+6+8]) << 48) | (uint64(buf[i+7+8]) << 56)

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
	return i + 24, nil
}
