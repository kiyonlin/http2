package http2

import (
	"sync"
)

const FrameData uint8 = 0x0

// Data defines a FrameData
type Data struct {
	noCopy     noCopy
	endStream  bool
	hasPadding bool
	b          []byte // data bytes
}

var dataPool = sync.Pool{
	New: func() interface{} {
		return &Data{}
	},
}

// AcquireData ...
func AcquireData() (data *Data) {
	data = dataPool.Get().(*Data)
	return
}

// ReleaseData ...
func ReleaseData(data *Data) {
	data.Reset()
	dataPool.Put(data)
}

// Reset ...
func (data *Data) Reset() {
	data.endStream = false
	data.b = data.b[:0]
}

// CopyTo copies data to d.
func (data *Data) CopyTo(d *Data) {
	d.hasPadding = data.hasPadding
	d.endStream = data.endStream
	d.b = append(d.b[:0], data.b...)
}

// SetEndStream ...
func (data *Data) SetEndStream(value bool) {
	data.endStream = value
}

// Data returns the byte slice of the data readed/to be sendStream.
func (data *Data) Data() []byte {
	return data.b
}

// SetData resets data byte slice and sets b.
func (data *Data) SetData(b []byte) {
	data.b = append(data.b[:0], b...)
}

// Padding returns true if the data will be/was hasPaddingded.
func (data *Data) Padding() bool {
	return data.hasPadding
}

// SetPadding sets hasPaddingding to the data if true. If false the data won't be hasPaddingded.
func (data *Data) SetPadding(value bool) {
	data.hasPadding = value
}

// Append appends b to data
func (data *Data) Append(b []byte) {
	data.b = append(data.b, b...)
}

func (data *Data) Len() uint32 {
	return uint32(len(data.b))
}

// Write writes b to data
func (data *Data) Write(b []byte) (int, error) {
	n := len(b)
	data.Append(b)
	return n, nil
}

// ReadFrame reads data from fr.
//
// This function does not reset the Frame.
func (data *Data) ReadFrame(fr *Frame) (err error) {
	payload := cutPadding(fr)
	data.endStream = fr.Has(FlagEndStream)
	if len(payload) == 0 {
		err = ErrZeroPayload
	} else {
		data.b = append(data.b[:0], payload...)
	}
	return
}

// WriteFrame writes the data to the frame payload setting FlagPadded.
//
// This function only resets the frame payload.
func (data *Data) WriteFrame(fr *Frame) {
	// TODO: generate hasPaddingding and set to the frame payload
	fr.SetType(FrameData)

	if data.endStream {
		fr.Add(FlagEndStream)
	}

	if data.hasPadding {
		fr.Add(FlagPadded)
		data.b = addPadding(data.b)
	}

	fr.SetPayload(data.b)
}
