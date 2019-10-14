package util

import (
	"astrongo/dclass/parse"
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDatagramIterator_Numerical(t *testing.T) {
	var dg Datagram
	var dgi *DatagramIterator

	dg = NewDatagram()
	dg.AddInt8(8)
	dgi = NewDatagramIterator(&dg)
	require.EqualValues(t, dgi.readInt8(), 8)

	dg = NewDatagram()
	dg.AddInt32(1234)
	dg.AddInt64(-123456789)
	dgi = NewDatagramIterator(&dg)
	require.EqualValues(t, dgi.readInt32(), 1234)
	require.EqualValues(t, dgi.readInt64(), -123456789)

	dg = NewDatagram()
	dg.AddFloat32(12.378839)
	dg.AddFloat64(128883.218389123)
	dgi = NewDatagramIterator(&dg)
	require.EqualValues(t, dgi.readFloat32(), float32(12.378839))
	require.EqualValues(t, dgi.readFloat64(), float64(128883.218389123))

	dg = NewDatagram()
	dg.AddBool(true)
	dgi = NewDatagramIterator(&dg)
	require.True(t, dgi.readBool())
}

func TestDatagramIterator_ReadString(t *testing.T) {
	var dg Datagram
	var dgi *DatagramIterator

	dg = NewDatagram()
	dg.AddString("hello")
	dgi = NewDatagramIterator(&dg)
	require.EqualValues(t, dgi.readString(), "hello")
}

func TestDatagramIterator_ReadData(t *testing.T) {
	var dg Datagram
	var dgi *DatagramIterator

	dg = NewDatagram()
	dg.AddDataBlob([]byte{'a', 'b', 'c', 'd'})
	dgi = NewDatagramIterator(&dg)
	require.ElementsMatch(t, dgi.readBlob(), []uint8{'a', 'b', 'c', 'd'})

	dg = NewDatagram()
	dg.AddString("test123")
	dgi = NewDatagramIterator(&dg)
	require.ElementsMatch(t, dgi.readData(8), []uint8{7, 0, 0, 0, 't', 'e', 's', 't'})

	dg = NewDatagram()
	dg.AddString("test123")
	dgi = NewDatagramIterator(&dg)
	dgi.readData(8)
	require.ElementsMatch(t, dgi.readRemainder(), []byte{'1', '2', '3'})
}

func TestDatagramIterator_Unpack(t *testing.T) {
	var dg Datagram
	var dgi *DatagramIterator
	var buff *bytes.Buffer
	buff = &bytes.Buffer{}

	dct, err := parse.ParseFile("util/test.dc")
	if err != nil {
		t.Fatalf("test dclass parse failed: %s", err)
	}

	dcf := dct.Traverse()
	cls, _ := dcf.ClassByName("methods")
	dg = NewDatagram()
	dg.AddUint32(123456789)
	dg.AddInt16(-1234)
	dg.AddUint8(32)
	dg.AddFloat64(3.14239285)
	dg.AddInt8(-100)
	dg.AddInt16(-10000)
	dgi = NewDatagramIterator(&dg)
	dgi.unpackDtype(cls, buff)
	dg = NewDatagram()
	dg.Write(buff.Bytes())
	dgi = NewDatagramIterator(&dg)
	require.EqualValues(t, dgi.readUint32(), 123456789)

	errChan := make(chan string)
	go func() {
		buff.Truncate(0)
		cls, _ = dcf.ClassByName("constraints1")
		dg = NewDatagram()
		dg.AddUint8(123)
		dgi = NewDatagramIterator(&dg)
		defer func() {
			if r := recover(); r == nil {
				errChan <- "numeric constraint violation test failed"
			}
			errChan <- ""
		}()
		dgi.unpackDtype(cls, buff)
	}()
	if err := <-errChan; err != "" {
		t.Errorf(err)
	}

	go func() {
		buff.Truncate(0)
		cls, _ = dcf.ClassByName("constraints2")
		dg = NewDatagram()
		dg.AddString("6Chars")
		dgi = NewDatagramIterator(&dg)
		defer func() {
			if r := recover(); r == nil {
				errChan <- "array constraint violation test failed"
			}
			errChan <- ""
		}()
		dgi.unpackDtype(cls, buff)
	}()
	if err := <-errChan; err != "" {
		t.Errorf(err)
	}

	go func() {
		buff.Truncate(0)
		cls, _ = dcf.ClassByName("arrays")
		dg = NewDatagram()
		dg.AddSize(3)
		dg.AddInt8(3)
		dg.AddInt8(4)
		dg.AddInt8(5)
		dg.AddSize(12)
		for n := 0; n < 12; n++ {
			dg.AddInt8(3)
		}
		dg.AddSize(4)
		for n := 0; n < 4; n++ {
			dg.AddInt8(99)
		}

		dgi = NewDatagramIterator(&dg)
		//defer func() {
		//	if r := recover(); r == nil {
		//		errChan <- "array constraint violation test failed"
		//	} else { fmt.Println(r) }
		//	errChan <- ""
		//}()
		dgi.unpackDtype(cls, buff)
	}()
	if err := <-errChan; err != "" {
		t.Errorf(err)
	}
}
