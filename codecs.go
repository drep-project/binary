// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

import (
	"encoding/binary"
	"errors"
	"math/big"
	"reflect"
)

var (
	MAXSLICESIZE = ^uint(0) >> 1
)

// Constants
var (
	LittleEndian = binary.LittleEndian
	BigEndian    = binary.BigEndian
)

// Codec represents a single part Codec, which can encode and decode something.
type Codec interface {
	EncodeTo(*Encoder, reflect.Value) error
	DecodeTo(*Decoder, reflect.Value) error
}

// ------------------------------------------------------------------------------

type reflectArrayCodec struct {
	elemCodec Codec // The codec of the array's elements
}

// Encode encodes a value into the encoder.
func (c *reflectArrayCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Type().Len()
	for i := 0; i < l; i++ {
		v := reflect.Indirect(rv.Index(i).Addr())
		if err = c.elemCodec.EncodeTo(e, v); err != nil {
			return
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectArrayCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	l := rv.Type().Len()
	for i := 0; i < l; i++ {
		if err = c.elemCodec.DecodeTo(d, rv.Index(i)); err != nil {
			return
		}
	}
	return
}

// ------------------------------------------------------------------------------

type reflectSliceCodec struct {
	elemCodec Codec // The codec of the slice's elements
}

// Encode encodes a value into the encoder.
func (c *reflectSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		v := reflect.Indirect(rv.Index(i).Addr())
		if err = c.elemCodec.EncodeTo(e, v); err != nil {
			return
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil && l > 0 {
		if l < uint64(MAXSLICESIZE) {
			rv.Set(reflect.MakeSlice(rv.Type(), int(l), int(l)))
			for i := 0; i < int(l); i++ {
				if err = c.elemCodec.DecodeTo(d, rv.Index(i)); err != nil {
					return
				}
			}
		}else {
			panic(errors.New("slice len exceed max size"))
		}
	}
	return
}

// ------------------------------------------------------------------------------

type byteSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *byteSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	b := rv.Bytes()
	e.WriteUvarint(uint64(len(b)))
	e.Write(b)
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *byteSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		if l < uint64(MAXSLICESIZE) {
			data := make([]byte, int(l), int(l))
			if _, err = d.Read(data); err == nil {
				rv.Set(reflect.ValueOf(data))
			}
		}else {
			panic(errors.New("slice len exceed max size"))
		}
	}
	return
}

type byteArrayCodec struct{}

// Encode encodes a value into the encoder.
func (c *byteArrayCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	if rv.CanAddr() {
		e.Write(rv.Slice(0, rv.Len()).Bytes())
	}else{
		arr := reflect.MakeSlice(reflect.TypeOf([]byte{}),rv.Len(), rv.Len())
		reflect.Copy(arr, rv)
		e.Write(arr.Bytes())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *byteArrayCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	l := rv.Len()
	byteArray := make([]byte, l)
	if _, err = d.Read(byteArray[:]); err == nil {
		reflect.Copy(rv, reflect.ValueOf(byteArray))
	}
	return
}

// ------------------------------------------------------------------------------

type boolSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	if l > 0 {
		v := rv.Interface().([]bool)
		e.Write(boolsToBinary(&v))
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *boolSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		if l < uint64(MAXSLICESIZE) {
			buf := make([]byte, l)
			_, err = d.r.Read(buf)
			rv.Set(reflect.ValueOf(binaryToBools(&buf)))
		}else {
			panic(errors.New("slice len exceed max size"))
		}
	}
	return
}

type boolArrayCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolArrayCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	arr := make([]byte, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		ele := rv.Index(i).Interface().(bool)
		if ele {
			arr[i] = 1
		} else {
			arr[i] = 0
		}
	}
	e.Write(arr[:])
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *boolArrayCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	l := rv.Len()
	buf := make([]byte, l)
	_, err = d.r.Read(buf)
	for index, val := range binaryToBools(&buf) {
		rv.Index(index).Set(reflect.ValueOf(val))
	}
	return
}

// ------------------------------------------------------------------------------

type varintSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		e.WriteVarint(rv.Index(i).Int())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *varintSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil && l > 0 {
		if l < uint64(MAXSLICESIZE) {
			slice := reflect.MakeSlice(rv.Type(), int(l), int(l))
			for i := 0; i < int(l); i++ {
				var v int64
				if v, err = binary.ReadVarint(d.r); err == nil {
					slice.Index(i).SetInt(v)
				}
			}
			rv.Set(slice)
		}else {
			panic(errors.New("slice len exceed max size"))
		}
	}
	return
}

type varintArrayCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintArrayCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	for i := 0; i < l; i++ {
		e.WriteVarint(rv.Index(i).Int())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *varintArrayCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	l := rv.Len()
	for i := 0; i < int(l); i++ {
		if v, err := binary.ReadVarint(d.r); err == nil {
			rv.Index(i).SetInt(v)
		}
	}
	return
}

// ------------------------------------------------------------------------------

type varuintSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *varuintSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		e.WriteUvarint(rv.Index(i).Uint())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *varuintSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l, v uint64
	if l, err = binary.ReadUvarint(d.r); err == nil && l > 0 {
		if l < uint64(MAXSLICESIZE) {
			slice := reflect.MakeSlice(rv.Type(), int(l), int(l))
			for i := 0; i < int(l); i++ {
				if v, err = d.ReadUvarint(); err == nil {
					slice.Index(i).SetUint(v)
				}
			}
			rv.Set(slice)
		}else {
			panic(errors.New("slice len exceed max size"))
		}
	}
	return
}

type varuintArrayCodec struct{}

// Encode encodes a value into the encoder.
func (c *varuintArrayCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	for i := 0; i < l; i++ {
		e.WriteUvarint(rv.Index(i).Uint())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *varuintArrayCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	l := rv.Len()
	for i := 0; i < int(l); i++ {
		if v, err := d.ReadUvarint(); err == nil {
			rv.Index(i).SetUint(v)
		}
	}
	return
}

// ------------------------------------------------------------------------------

type reflectStructCodec []fieldCodec

type fieldCodec struct {
	Index int   // The index of the field
	Codec Codec // The codec to use for this field
}

// Encode encodes a value into the encoder.
func (c *reflectStructCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	for _, i := range *c {
		if err = i.Codec.EncodeTo(e, rv.Field(i.Index)); err != nil {
			return err
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectStructCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	for _, i := range *c {
		if v := rv.Field(i.Index); v.CanSet() {
			if err = i.Codec.DecodeTo(d, v); err != nil {
				return err
			}
		}
	}
	return
}

type reflectPtrCodec struct {
	Codec Codec
}

// Encode encodes a value into the encoder.
func (c *reflectPtrCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	if rv.IsNil() {
		e.writeBool(true)
	} else {
		e.writeBool(false)
		return c.Codec.EncodeTo(e, rv.Elem())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectPtrCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	isNil, err := d.ReadBool()
	if isNil {
		return nil
	} else {
		v := rv.Type().Elem()
		zeroV := reflect.New(v)
		rv.Set(zeroV)
		return c.Codec.DecodeTo(d, rv.Elem())
	}
	return
}

// ------------------------------------------------------------------------------

// customCodec represents a custom binary marshaling.
type customCodec struct {
	marshaler      *reflect.Method
	unmarshaler    *reflect.Method
	ptrMarshaler   *reflect.Method
	ptrUnmarshaler *reflect.Method
}

// Encode encodes a value into the encoder.
func (c *customCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	m := c.GetMarshalBinary(rv)
	if m == nil {
		return errors.New("MarshalBinary not found on " + rv.Type().String())
	}

	ret := m.Call([]reflect.Value{})
	if !ret[1].IsNil() {
		err = ret[1].Interface().(error)
		return
	}

	// Write the marshaled byte slice
	buffer := ret[0].Bytes()
	e.WriteUvarint(uint64(len(buffer)))
	e.Write(buffer)
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *customCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	m := c.GetUnmarshalBinary(rv)

	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil {
		buffer := make([]byte, l)
		_, err = d.r.Read(buffer)
		ret := m.Call([]reflect.Value{reflect.ValueOf(buffer)})
		if !ret[0].IsNil() {
			err = ret[0].Interface().(error)
		}

	}
	return
}

func (c *customCodec) GetMarshalBinary(rv reflect.Value) *reflect.Value {
	if c.marshaler != nil {
		m := rv.Method(c.marshaler.Index)
		return &m
	}

	if c.ptrMarshaler != nil {
		m := rv.Addr().Method(c.ptrMarshaler.Index)
		return &m
	}

	return nil
}

func (c *customCodec) GetUnmarshalBinary(rv reflect.Value) *reflect.Value {
	if c.unmarshaler != nil {
		m := rv.Method(c.unmarshaler.Index)
		return &m
	}

	if c.ptrUnmarshaler != nil {
		m := rv.Addr().Method(c.ptrUnmarshaler.Index)
		return &m
	}

	return nil
}

// ------------------------------------------------------------------------------

type reflectMapCodec struct {
	key Codec // Codec for the key
	val Codec // Codec for the value
}

// Encode encodes a value into the encoder.
func (c *reflectMapCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	e.WriteUvarint(uint64(rv.Len()))
	for _, key := range rv.MapKeys() {
		value := rv.MapIndex(key)
		if err = c.writeKey(e, key); err != nil {
			return err
		}

		if err = c.val.EncodeTo(e, value); err != nil {
			return err
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectMapCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil {
		t := rv.Type()
		vt := t.Elem()
		rv.Set(reflect.MakeMap(t))
		for i := 0; i < int(l); i++ {

			var kv reflect.Value
			if kv, err = c.readKey(d, t.Key()); err != nil {
				return
			}

			vv := reflect.New(vt).Elem()
			if err = c.val.DecodeTo(d, vv); err != nil {
				return
			}

			rv.SetMapIndex(kv, vv)
		}
	}
	return
}

// Write key writes a key to the encoder
func (c *reflectMapCodec) writeKey(e *Encoder, key reflect.Value) (err error) {
	switch key.Kind() {

	case reflect.Int16:
		e.WriteUint16(uint16(key.Int()))
	case reflect.Int32:
		e.WriteUint32(uint32(key.Int()))
	case reflect.Int64:
		e.WriteUint64(uint64(key.Int()))

	case reflect.Uint16:
		e.WriteUint16(uint16(key.Uint()))
	case reflect.Uint32:
		e.WriteUint32(uint32(key.Uint()))
	case reflect.Uint64:
		e.WriteUint64(uint64(key.Uint()))

	case reflect.String:
		str := key.String()
		e.WriteUint16(uint16(len(str)))
		e.Write(stringToBinary(str))
	default:
		err = c.key.EncodeTo(e, key)
	}
	return
}

// Read key reads a key from the decoder
func (c *reflectMapCodec) readKey(d *Decoder, keyType reflect.Type) (key reflect.Value, err error) {
	switch keyType.Kind() {
	case reflect.Int16:
		var v uint16
		if v, err = d.ReadUint16(); err == nil {
			key = reflect.ValueOf(int16(v))
		}
	case reflect.Int32:
		var v uint32
		if v, err = d.ReadUint32(); err == nil {
			key = reflect.ValueOf(int32(v))
		}
	case reflect.Int64:
		var v uint64
		if v, err = d.ReadUint64(); err == nil {
			key = reflect.ValueOf(int64(v))
		}

	case reflect.Uint16:
		var v uint16
		if v, err = d.ReadUint16(); err == nil {
			key = reflect.ValueOf(v)
		}
	case reflect.Uint32:
		var v uint32
		if v, err = d.ReadUint32(); err == nil {
			key = reflect.ValueOf(v)
		}
	case reflect.Uint64:
		var v uint64
		if v, err = d.ReadUint64(); err == nil {
			key = reflect.ValueOf(v)
		}

	// String keys must have max length of 65536
	case reflect.String:
		var l uint16
		var b []byte

		if l, err = d.ReadUint16(); err == nil {
			if b, err = d.Slice(int(l)); err == nil {
				key = reflect.ValueOf(string(b))
			}
		}

	// Default to a reflect-based approach
	default:
		key = reflect.New(keyType).Elem()
		err = c.key.DecodeTo(d, key)
	}
	return
}

// ------------------------------------------------------------------------------

type stringCodec struct{}

// Encode encodes a value into the encoder.
func (c *stringCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	str := rv.String()
	e.WriteUvarint(uint64(len(str)))
	e.Write(stringToBinary(str))
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *stringCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	var b []byte

	if l, err = d.ReadUvarint(); err == nil {
		if b, err = d.Slice(int(l)); err == nil {
			rv.SetString(string(b))
		}
	}
	return
}

// ------------------------------------------------------------------------------

type boolCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeBool(rv.Bool())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *boolCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out bool
	if out, err = d.ReadBool(); err == nil {
		rv.SetBool(out)
	}
	return
}

// ------------------------------------------------------------------------------

type varintCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.WriteVarint(rv.Int())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *varintCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var v int64
	if v, err = binary.ReadVarint(d.r); err != nil {
		return
	}
	rv.SetInt(v)
	return
}

// ------------------------------------------------------------------------------

type varuintCodec struct{}

// Encode encodes a value into the encoder.
func (c *varuintCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.WriteUvarint(rv.Uint())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *varuintCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var v uint64
	if v, err = binary.ReadUvarint(d.r); err != nil {
		return
	}
	rv.SetUint(v)
	return
}

// ------------------------------------------------------------------------------

type complex64Codec struct{}

// Encode encodes a value into the encoder.
func (c *complex64Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeComplex64(complex64(rv.Complex()))
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *complex64Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out complex64
	out, err = d.readComplex64()
	rv.SetComplex(complex128(out))
	return
}

// ------------------------------------------------------------------------------

type complex128Codec struct{}

// Encode encodes a value into the encoder.
func (c *complex128Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeComplex128(rv.Complex())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *complex128Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out complex128
	out, err = d.readComplex128()
	rv.SetComplex(out)
	return
}

// ------------------------------------------------------------------------------

type float32Codec struct{}

// Encode encodes a value into the encoder.
func (c *float32Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.WriteFloat32(float32(rv.Float()))
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float32Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var v float32
	if v, err = d.ReadFloat32(); err == nil {
		rv.SetFloat(float64(v))
	}
	return
}

// ------------------------------------------------------------------------------

type float64Codec struct{}

// Encode encodes a value into the encoder.
func (c *float64Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.WriteFloat64(rv.Float())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float64Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var v float64
	if v, err = d.ReadFloat64(); err == nil {
		rv.SetFloat(v)
	}
	return
}

// ------------------------------------------------------------------------------

type bigIntCodec struct{}

// Encode encodes a value into the encoder.
func (c *bigIntCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	val := rv.Interface().(big.Int)
	contents := val.Bytes()
	e.WriteUvarint(uint64(len(contents)))
	e.Write(contents)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *bigIntCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	len, err := d.ReadUvarint()
	if err != nil {
		return err
	}
	if len < uint64(MAXSLICESIZE) {
		contents := make([]byte, len)
		_, err = d.Read(contents)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(*new(big.Int).SetBytes(contents)))
		return nil
	}else {
		panic(errors.New("slice len exceed max size"))
	}
}

//
type reflectInterfaceCodec struct {
	Codec Codec
}

// Encode encodes a value into the encoder.
func (c *reflectInterfaceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	if rv.IsNil() {
		e.writeBool(true)
	} else {
		e.writeBool(false)
		realValue := getUnderlyingType(rv)
		realCodeC, err := scanType(realValue.Type())
		if err != nil {
			return err
		}
		realCodeC.EncodeTo(e, realValue)
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectInterfaceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	isNil, err := d.ReadBool()
	if isNil {
		return nil
	} else {
		realValue := getUnderlyingType(rv.Addr())
		realCodeC, err := scanType(realValue.Type())
		if err != nil {
			return err
		}
		realCodeC.DecodeTo(d, realValue)
	}
	return
}

//
type reflectInterfaceArrayCodec struct {
}

// Encode encodes a value into the encoder.
func (c *reflectInterfaceArrayCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Type().Len()
	for i := 0; i < l; i++ {
		val := rv.Index(i)
		if val.IsNil() {
			e.writeBool(true)
			continue
		}
		e.writeBool(false)
		realValue := getUnderlyingType(val)
		realCodeC, err := scanType(realValue.Type())
		if err = realCodeC.EncodeTo(e, realValue); err != nil {
			return err
		}
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectInterfaceArrayCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	l := rv.Type().Len()
	for i := 0; i < l; i++ {
		isNil, err := d.ReadBool()
		if err != nil {
			return err
		}
		if isNil {
			continue
		}
		val := rv.Index(i)
		realValue := getUnderlyingType(val)
		realCodeC, err := scanType(realValue.Type())
		newVal := reflect.New(realValue.Type()).Elem()
		if err = realCodeC.DecodeTo(d, newVal); err != nil {
			return err
		}
		val.Set(newVal)
	}
	return nil
}

//
type reflectInterfaceSliceCodec struct {
}

// Encode encodes a value into the encoder.
func (c *reflectInterfaceSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		val := reflect.Indirect(rv.Index(i).Addr())
		if val.IsNil() {
			e.writeBool(true)
			continue
		}
		e.writeBool(false)
		realValue := getUnderlyingType(val)
		realCodeC, err := scanType(realValue.Type())
		if err = realCodeC.EncodeTo(e, realValue); err != nil {
			return err
		}
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectInterfaceSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil && l > 0 {
		for i := 0; i < int(l); i++ {
			isNil, err := d.ReadBool()
			if err != nil {
				return err
			}
			if isNil {
				continue
			}
			val := rv.Index(i)
			realValue := getUnderlyingType(val)
			realCodeC, err := scanType(realValue.Type())
			newVal := reflect.New(realValue.Type()).Elem()
			if err = realCodeC.DecodeTo(d, newVal); err != nil {
				return err
			}
			val.Set(newVal)
		}
	}
	return nil
}

func getUnderlyingType(val reflect.Value) reflect.Value {
	for {
		if val.Kind() == reflect.Interface {
			val = val.Elem()
		} else {
			break
		}
	}
	return val
}
