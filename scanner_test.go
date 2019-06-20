// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCustom string

type ByteRedefine []byte

type MM struct {
	AA string
	TestBytes []ByteRedefine
}
type Cap struct {
	Name    string
	Version uint
}
type ProtoHandshake struct {
	Version    uint64
	Name       string
	Caps       []Cap
	ListenPort uint64
	ID         []byte // secp256k1 public key
	// Ignore additional fields (for forward compatibility).
	Rest []ByteRedefine `rlp:"tail"`
}

type  AAA MM

func TestScanner_Cus2222tom1(t *testing.T) {
	bytes33,_ := hex.DecodeString("05096c6f63616c6e6f6465020c636861696e536572766963650010636f6e73656e73757353657276696365000040783993bbcec68f37585983cea8b4325b04fc0876dd54d28275a651d38bdd04743ea7edb5c14d6857310466ee60851da3ad4a386aebab8e3528f4d09f92ccb6df00")
	fmt.Println(len(bytes33))
	reader := bytes.NewReader(bytes33)
	binary.ReadUvarint(reader)
	fmt.Println(reader.Len())
	xxx := []byte{3,2,1}
	aaa := ProtoHandshake {
		Version: 123,
		Name:"XXXXX",
		Caps:[]Cap{
			Cap{
				Name:"xxx",
				Version:0,
			},
		},
		ListenPort:0,
		ID:[]byte{1,2,3,4,5,6,},
		Rest:[]ByteRedefine{xxx},
	}

	_, err :=Marshal(aaa)
	if err != nil {
		log.Fatal(err)
	}


	sss := ProtoHandshake{}
	err = Unmarshal(bytes33, &sss)
	if err != nil {
		log.Fatal(err)
	}
}
func TestScanner_Slice_interface(t *testing.T) {
	var testV []interface{}
	testV = append(testV, 1)
	testV = append(testV, "xxxx")
	testV = append(testV, []byte{1,2,3})

	sBytes, err := Marshal(testV)
	if err != nil {
		log.Fatal(err)
	}
	var testV2 []interface{}
	testV2 = append(testV2, 1)
	testV2 = append(testV2, "")
	testV2 = append(testV2, []byte{})

	ss := reflect.Indirect(reflect.ValueOf(&testV2))
	val2 := ss.Index(0)
	for {
		if val2.Kind() == reflect.Interface {
			val2 = val2.Elem()
		}else{
			break
		}
	}
	fmt.Println(val2.String())
	err = Unmarshal(sBytes, &testV2)
	if err != nil {
		log.Fatal(err)
	}
}

func TestScanner_Array_interface(t *testing.T) {
	var testV [3]interface{}
	testV[0] = 1
	testV[1] = "xxxx"
	testV[2] = []byte{1,2,3}


	sBytes, err := Marshal(testV)
	if err != nil {
		log.Fatal(err)
	}
	var testV2 [3]interface{}
	testV2[0] = 0
	testV2[1] = ""
	testV2[2] = []byte{}

	ss := reflect.Indirect(reflect.ValueOf(&testV))
	val2 := ss.Index(0)
	for {
		if val2.Kind() == reflect.Interface {
			val2 = val2.Elem()
		}else{
			break
		}
	}
	fmt.Println(val2.String())
	err = Unmarshal(sBytes, &testV2)
	if err != nil {
		log.Fatal(err)
	}
}

// GetBinaryCodec retrieves a custom binary codec.
func (s *testCustom) GetBinaryCodec() Codec {
	return new(stringCodec)
}

func TestScanner(t *testing.T) {
	rt := reflect.Indirect(reflect.ValueOf(s0v)).Type()
	codec, err := scan(rt)
	assert.NoError(t, err)
	assert.NotNil(t, codec)

	var b bytes.Buffer
	e := NewEncoder(&b)
	err = codec.EncodeTo(e, reflect.Indirect(reflect.ValueOf(s0v)))
	assert.NoError(t, err)

	//e.Flush()
	assert.Equal(t, s0b, b.Bytes())
}

func TestScanner_Custom(t *testing.T) {
	v := testCustom("test")
	rt := reflect.Indirect(reflect.ValueOf(v)).Type()
	codec, err := scan(rt)
	assert.NoError(t, err)
	assert.NotNil(t, codec)
}

