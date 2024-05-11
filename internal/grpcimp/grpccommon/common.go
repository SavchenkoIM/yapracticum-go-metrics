package grpccommon

import (
	"bytes"
	"encoding/binary"
	"math"
	"yaprakticum-go-track2/internal/grpcimp"
)

func float64ToByte(f float64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(f))
	return buf[:]
}

func int64ToByte(i int64) []byte {
	var buf [8]byte
	var ui uint64
	if i > 0 {
		ui = uint64(i)
	} else {
		ui = math.MaxUint64 - uint64(i)
	}
	binary.BigEndian.PutUint64(buf[:], ui)
	return buf[:]
}

func MetricDataToByteSlice(data []*grpcimp.MetricData) []byte {
	res := make([]byte, 0)

	for _, v := range data {
		bb := bytes.NewBuffer(make([]byte, 0))
		v := v
		bb.Read([]byte(v.Name))
		bb.Read(int64ToByte(v.Delta))
		bb.Read(float64ToByte(v.Value))
		bb.Read(int64ToByte(int64(v.Type)))
		res = append(res, bb.Bytes()...)
	}

	return res
}
