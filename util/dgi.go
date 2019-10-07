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
	dg     *Datagram
	offset Dgsize_t
	read   *bytes.Reader
}

func NewDatagramIterator(dg *Datagram) *DatagramIterator {
	dgi := &DatagramIterator{dg: dg, read: bytes.NewReader(dg.Bytes())}
	return dgi
}

func (dgi *DatagramIterator) panic(len int8) {
	panic(DatagramIteratorEOF{
		fmt.Sprintf("datagram iterator eof, read length: %d buff length: %d", len, dgi.read.Len()),
	})
}

func (dgi *DatagramIterator) readBool() bool {
	val := dgi.readUint8()
	if val != 0 {
		return true
	} else {
		return false
	}
}

func (dgi *DatagramIterator) readInt8() int8 {
	var val int8
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(1)
	}

	dgi.offset += 1
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readInt16() int16 {
	var val int16
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(2)
	}

	dgi.offset += 2
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readInt32() int32 {
	var val int32
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(4)
	}

	dgi.offset += 4
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readInt64() int64 {
	var val int64
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(8)
	}

	dgi.offset += 8
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readUint8() uint8 {
	var val uint8
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(1)
	}

	dgi.offset += 1
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readUint16() uint16 {
	var val uint16
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(2)
	}

	dgi.offset += 2
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readUint32() uint32 {
	var val uint32
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(4)
	}

	dgi.offset += 4
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readUint64() uint64 {
	var val uint64
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(8)
	}

	dgi.offset += 8
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readSize() Dgsize_t {
	var val Dgsize_t
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(Dgsize)
	}

	dgi.offset += Dgsize
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readChannel() Channel_t {
	var val Channel_t
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(Chansize)
	}

	dgi.offset += Chansize
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readDoid() Doid_t {
	var val Doid_t
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(Doidsize)
	}

	dgi.offset += Doidsize
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readZone() Zone_t {
	var val Zone_t
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(Zonesize)
	}

	dgi.offset += Zonesize
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readFloat32() float32 {
	var val float32
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(4)
	}

	dgi.offset += 4
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readFloat64() float64 {
	var val float64
	if err := binary.Read(dgi.read, binary.LittleEndian, &val); err != nil {
		dgi.panic(8)
	}

	dgi.offset += 8
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return val
}

func (dgi *DatagramIterator) readString() string {
	sz := dgi.readSize()
	buff := make([]byte, sz)
	if _, err := dgi.read.Read(buff); err != nil {
		dgi.panic(int8(sz))
	}

	dgi.offset += sz
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return string(buff)
}

func (dgi *DatagramIterator) readBlob() []uint8 {
	return dgi.readData(dgi.readSize())
}

func (dgi *DatagramIterator) readDatagram() *Datagram {
	data := dgi.readBlob()
	dg := NewDatagram()
	dg.Write(data)
	return &dg
}

func (dgi *DatagramIterator) readData(length Dgsize_t) []uint8 {
	buff := make([]uint8, int32(length))
	if _, err := dgi.read.Read(buff); err != nil {
		dgi.panic(int8(length))
	}

	dgi.offset += length
	dgi.read.Seek(int64(dgi.offset), io.SeekStart)
	return buff
}

func (dgi *DatagramIterator) readRemainder() []uint8 {
	sz := Dgsize_t(dgi.dg.Len()) - dgi.offset
	return dgi.readData(sz)
}

// Shorthand for unpackDtype
func (dgi *DatagramIterator) unpackField(field dc.Field, buffer bytes.Buffer) {
	dgi.unpackDtype(field.FieldType(), buffer)
}

func (dgi *DatagramIterator) unpackDtype(dtype dc.BaseType, buffer bytes.Buffer) {
	fixed := (dtype.Type() == dc.T_METHOD || dtype.Type() == dc.T_STRUCT) || dtype.HasRange()

	if dtype.HasFixedSize() && !fixed {
		if array := dtype.(*dc.ArrayType); array != nil && array.ElementType().HasRange() && dtype.Type() == dc.T_ARRAY {
			for n := 0; n < int(array.Size()); n++ {
				dgi.unpackDtype(array.ElementType(), buffer)
			}
		}

		num := dtype.(*dc.NumericType)
		data := dgi.readData(Dgsize_t(dtype.Size()))

		if num != nil && num.HasRange() {
			if !num.WithinRange(data, 0) {
				panic(FieldConstraintViolation{
					fmt.Sprintf("field constraint violation: failed to unpack numeric type %s", dtype.Alias()),
				})
			}
		}

		buffer.Write(data)
	}

	switch dtype.Type() {
	case dc.T_VARSTRING, dc.T_VARBLOB, dc.T_VARARRAY:
		var elemCount uint64
		array := dtype.(*dc.ArrayType)
		len := dgi.readSize()

		netlen := make([]byte, 4)
		binary.BigEndian.PutUint32(netlen, uint32(len))
		buffer.Write(netlen)

		if dtype.Type() == dc.T_VARARRAY {
			sz := buffer.Len()

			for elemCount = 0; buffer.Len()-sz < int(len); elemCount++ {
				dgi.unpackDtype(array.ElementType(), buffer)
			}
		} else {
			data := dgi.readData(len)
			buffer.Write(data)
			elemCount = uint64(len)
		}

		if !array.WithinRange(nil, elemCount) {
			panic(FieldConstraintViolation{
				fmt.Sprintf("field constraint violation: failed to unpack array type %s", dtype.Alias()),
			})
		}
	case dc.T_STRUCT:
		strct := dtype.(*dc.Struct)
		fields := strct.GetNumFields()
		for n := 0; n < fields; n++ {
			dgi.unpackDtype(strct.GetField(n).FieldType(), buffer)
		}
	case dc.T_METHOD:
		method := dtype.(*dc.Method)
		params := method.GetNumParameters()
		for n := 0; n < params; n++ {
			dgi.unpackDtype(method.GetParameter(n).Type(), buffer)
		}
	}
}

func (dgi *DatagramIterator) skipField(field dc.Field) {
	dgi.skipDtype(field.FieldType())
}

func (dgi *DatagramIterator) skipDtype(dtype dc.BaseType) {
	if dtype.HasFixedSize() {
		len := dtype.Size()
		dgi.skip(Dgsize_t(len))
	}

	switch dtype.Type() {
	case dc.T_VARSTRING, dc.T_VARBLOB, dc.T_VARARRAY:
		len := dgi.readSize()
		dgi.skip(len)
	case dc.T_STRUCT:
		strct := dtype.(*dc.Struct)
		fields := strct.GetNumFields()
		for n := 0; n < fields; n++ {
			dgi.skipDtype(strct.GetField(n).FieldType())
		}
	case dc.T_METHOD:
		method := dtype.(*dc.Method)
		params := method.GetNumParameters()
		for n := 0; n < params; n++ {
			dgi.skipDtype(method.GetParameter(n).Type())
		}
	}
}

func (dgi *DatagramIterator) receipientCount() uint8 {
	if dgi.read.Len() == 0 {
		return 0
	}

	return dgi.dg.Bytes()[0]
}

func (dgi *DatagramIterator) sender() Channel_t {
	offset := dgi.offset

	dgi.offset = 1 + Dgsize_t(dgi.receipientCount())*Chansize
	sender := dgi.readChannel()

	dgi.offset = offset
	return sender
}

func (dgi *DatagramIterator) messageType() uint16 {
	offset := dgi.offset

	dgi.offset = 1 + Dgsize_t(dgi.receipientCount())*(Chansize+1)
	msg := dgi.readUint16()

	dgi.offset = offset
	return msg
}

func (dgi *DatagramIterator) tell() Dgsize_t {
	return dgi.offset
}

func (dgi *DatagramIterator) seek(pos Dgsize_t) {
	dgi.offset = pos
}

func (dgi *DatagramIterator) seekPayload() {
	dgi.seek(0)
	dgi.offset = 1 + Dgsize_t(dgi.receipientCount())*Chansize
}

func (dgi *DatagramIterator) skip(len Dgsize_t) {
	if dgi.offset+len > Dgsize_t(dgi.dg.Len()) {
		dgi.panic(int8(len))
	}

	dgi.offset += len
}
