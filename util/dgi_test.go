package util

import (
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
