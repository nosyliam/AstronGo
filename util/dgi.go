package util

import (
	"astrongo/dclass/dc"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type DatagramIteratorEOF struct {
	err string
}

type FieldConstraintViolation struct {
	err string
}

type DatagramIterator struct {
	Dg     *Datagram
	offset Dgsize_t
	Read   *bytes.Reader
}

func NewDatagramIterator(dg *Datagram) *DatagramIterator {
	dgi := &DatagramIterator{Dg: dg, Read: bytes.NewReader(dg.Bytes())}
	return dgi
}

func (dgi *DatagramIterator) Copy() *DatagramIterator {
	newDgi := NewDatagramIterator(dgi.Dg)
	newDgi.Seek(dgi.Tell())
	return dgi
}

func (dgi *DatagramIterator) panic(len int8) {
	panic(DatagramIteratorEOF{
		fmt.Sprintf("datagram iterator eof, read length: %d buff length: %d", len, dgi.Read.Len()),
	})
}

func (dgi *DatagramIterator) ReadBool() bool {
	val := dgi.ReadUint8()
	if val != 0 {
		return true
	} else {
		return false
	}
}

func (dgi *DatagramIterator) ReadInt8() int8 {
	var val int8
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(1)
	}

	dgi.offset += 1
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadInt16() int16 {
	var val int16
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(2)
	}

	dgi.offset += 2
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadInt32() int32 {
	var val int32
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(4)
	}

	dgi.offset += 4
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadInt64() int64 {
	var val int64
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(8)
	}

	dgi.offset += 8
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadUint8() uint8 {
	var val uint8
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(1)
	}

	dgi.offset += 1
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadUint16() uint16 {
	var val uint16
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(2)
	}

	dgi.offset += 2
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadUint32() uint32 {
	var val uint32
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(4)
	}

	dgi.offset += 4
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadUint64() uint64 {
	var val uint64
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(8)
	}

	dgi.offset += 8
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadSize() Dgsize_t {
	var val Dgsize_t
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(Dgsize)
	}

	dgi.offset += Dgsize
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadChannel() Channel_t {
	var val Channel_t
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(Chansize)
	}

	dgi.offset += Chansize
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadDoid() Doid_t {
	var val Doid_t
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(Doidsize)
	}

	dgi.offset += Doidsize
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadZone() Zone_t {
	var val Zone_t
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(Zonesize)
	}

	dgi.offset += Zonesize
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadFloat32() float32 {
	var val float32
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(4)
	}

	dgi.offset += 4
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadFloat64() float64 {
	var val float64
	if err := binary.Read(dgi.Read, binary.LittleEndian, &val); err != nil {
		dgi.panic(8)
	}

	dgi.offset += 8
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) ReadString() string {
	sz := dgi.ReadSize()
	buff := make([]byte, sz)
	if _, err := dgi.Read.Read(buff); err != nil {
		dgi.panic(int8(sz))
	}

	dgi.offset += sz
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return string(buff)
}

func (dgi *DatagramIterator) ReadBlob() []uint8 {
	return dgi.ReadData(dgi.ReadSize())
}

func (dgi *DatagramIterator) ReadDatagram() *Datagram {
	data := dgi.ReadBlob()
	dg := NewDatagram()
	dg.Write(data)
	return &dg
}

func (dgi *DatagramIterator) ReadData(length Dgsize_t) []uint8 {
	buff := make([]uint8, int32(length))
	if _, err := dgi.Read.Read(buff); err != nil {
		dgi.panic(int8(length))
	}

	dgi.offset += length
	dgi.Read.Seek(int64(dgi.offset), io.SeekStart)
	return buff
}

func (dgi *DatagramIterator) ReadRemainder() []uint8 {
	sz := Dgsize_t(dgi.Dg.Len()) - dgi.offset
	return dgi.ReadData(sz)
}

func (dgi *DatagramIterator) UnpackFieldtoUint8(field dc.Field) []uint8 {
	b := &bytes.Buffer{}
	dgi.UnpackField(field, b)
	return b.Bytes()
}

// Shorthand for unpackDtype
func (dgi *DatagramIterator) UnpackField(field dc.Field, buffer *bytes.Buffer) {
	dgi.UnpackDtype(field.FieldType(), buffer)
}

func (dgi *DatagramIterator) UnpackDtype(dtype dc.BaseType, buffer *bytes.Buffer) {
	fixed := (dtype.Type() == dc.T_METHOD || dtype.Type() == dc.T_STRUCT) && dtype.HasRange()

	if dtype.HasFixedSize() && !fixed {
		if array, ok := dtype.(*dc.ArrayType); ok && array != nil && array.ElementType().HasRange() && dtype.Type() == dc.T_ARRAY {
			dgi.ReadSize()
			for n := 0; n < int(array.Size()); n++ {
				dgi.UnpackDtype(array.ElementType(), buffer)
			}
		}

		data := dgi.ReadData(Dgsize_t(dtype.Size()))

		if num, ok := dtype.(*dc.NumericType); ok && num != nil && num.HasRange() {
			if !num.WithinRange(data, 0) {
				panic(FieldConstraintViolation{
					fmt.Sprintf("field constraint violation: failed to unpack numeric type %s", dtype.Alias()),
				})
			}
		}

		buffer.Write(data)
		return
	}

	switch dtype.Type() {
	case dc.T_VARSTRING, dc.T_VARBLOB, dc.T_VARARRAY:
		var elemCount uint64
		array := dtype.(*dc.ArrayType)
		len := dgi.ReadSize()

		netlen := make([]byte, 4)
		binary.BigEndian.PutUint32(netlen, uint32(len))
		buffer.Write(netlen)

		if dtype.Type() == dc.T_VARARRAY {
			sz := buffer.Len()

			for elemCount = 0; buffer.Len()-sz < int(len); elemCount++ {
				dgi.UnpackDtype(array.ElementType(), buffer)
			}
		} else {
			data := dgi.ReadData(len)
			buffer.Write(data)
			elemCount = uint64(len)
		}

		if !array.WithinRange(nil, elemCount) {
			panic(FieldConstraintViolation{
				fmt.Sprintf("field constraint violation: failed to unpack array type %s", dtype.Alias()),
			})
		}
	case dc.T_STRUCT:
		var strct *dc.Struct
		if strct, _ = dtype.(*dc.Struct); strct == nil {
			cls := dtype.(*dc.Class)
			strct = &cls.Struct
		}
		fields := strct.GetNumFields()
		for n := 0; n < fields; n++ {
			dgi.UnpackDtype(strct.GetField(n).FieldType(), buffer)
		}
	case dc.T_METHOD:
		method := dtype.(*dc.Method)
		params := method.GetNumParameters()
		for n := 0; n < params; n++ {
			dgi.UnpackDtype(method.GetParameter(n).Type(), buffer)
		}
	}
}

func (dgi *DatagramIterator) SkipField(field dc.Field) {
	dgi.SkipDtype(field.FieldType())
}

func (dgi *DatagramIterator) SkipDtype(dtype dc.BaseType) {
	if dtype.HasFixedSize() {
		len := dtype.Size()
		dgi.Skip(Dgsize_t(len))
		return
	}

	switch dtype.Type() {
	case dc.T_VARSTRING, dc.T_VARBLOB, dc.T_VARARRAY:
		len := dgi.ReadSize()
		dgi.Skip(len)
	case dc.T_STRUCT:
		var strct *dc.Struct
		if strct, _ = dtype.(*dc.Struct); strct == nil {
			cls := dtype.(*dc.Class)
			strct = &cls.Struct
		}
		fields := strct.GetNumFields()
		for n := 0; n < fields; n++ {
			dgi.SkipDtype(strct.GetField(n).FieldType())
		}
	case dc.T_METHOD:
		method := dtype.(*dc.Method)
		params := method.GetNumParameters()
		for n := 0; n < params; n++ {
			dgi.SkipDtype(method.GetParameter(n).Type())
		}
	}
}

func (dgi *DatagramIterator) RecipientCount() uint8 {
	if dgi.Read.Len() == 0 {
		return 0
	}

	return dgi.Dg.Bytes()[0]
}

func (dgi *DatagramIterator) Sender() Channel_t {
	offset := dgi.offset

	dgi.offset = 1 + Dgsize_t(dgi.RecipientCount())*Chansize
	sender := dgi.ReadChannel()

	dgi.offset = offset
	return sender
}

func (dgi *DatagramIterator) MessageType() uint16 {
	offset := dgi.offset

	dgi.offset = 1 + Dgsize_t(dgi.RecipientCount())*(Chansize+1)
	msg := dgi.ReadUint16()

	dgi.offset = offset
	return msg
}

func (dgi *DatagramIterator) Tell() Dgsize_t {
	return dgi.offset
}

func (dgi *DatagramIterator) Seek(pos Dgsize_t) {
	dgi.offset = pos
}

func (dgi *DatagramIterator) SeekPayload() {
	dgi.Seek(0)
	dgi.offset = 1 + Dgsize_t(dgi.RecipientCount())*Chansize
}

func (dgi *DatagramIterator) Skip(len Dgsize_t) {
	if dgi.offset+len > Dgsize_t(dgi.Dg.Len()) {
		dgi.panic(int8(len))
	}

	dgi.offset += len
}
