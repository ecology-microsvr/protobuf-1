// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protoimpl

import (
	"fmt"
	"sync"
)

type Message interface {
	Reset()
	String() string
	ProtoMessage()
}

type ExtensionDesc struct {
	ExtendedType  Message     // nil pointer to the type that is being extended
	ExtensionType interface{} // nil pointer to the extension type
	Field         int32       // field number
	Name          string      // fully-qualified name of extension, for text formatting
	Tag           string      // protobuf tag style
	Filename      string      // name of the file in which the extension is defined
}

func ExtensionFieldsOf(p interface{}) ExtensionFields {
	switch p := p.(type) {
	case *map[int32]Extension:
		return (*extensionMap)(p)
	case *XXX_InternalExtensions:
		return (*extensionSyncMap)(p)
	default:
		panic(fmt.Sprintf("invalid extension fields type: %T", p))
	}
}

type ExtensionFields = extensionFields

type extensionFields interface {
	Len() int
	Has(int32) bool
	Get(int32) Extension
	Set(int32, Extension)
	Clear(int32)
	Range(f func(int32, Extension) bool)

	HasInit() bool
	sync.Locker
}

// Deprecated: Do not use.
type Extension = extensionField

type extensionField struct {
	// When an extension is stored in a message using SetExtension
	// only desc and value are set. When the message is marshaled
	// Raw will be set to the encoded form of the message.
	//
	// When a message is unmarshaled and contains extensions, each
	// extension will have only Raw set. When such an extension is
	// accessed using GetExtension (or GetExtensions) desc and value
	// will be set.
	Desc *ExtensionDesc

	// value is a concrete value for the extension field. Let the type of
	// desc.ExtensionType be the "API type" and the type of Extension.value
	// be the "storage type". The API type and storage type are the same except:
	//	* For scalars (except []byte), the API type uses *T,
	//	while the storage type uses T.
	//	* For repeated fields, the API type uses []T, while the storage type
	//	uses *[]T.
	//
	// The reason for the divergence is so that the storage type more naturally
	// matches what is expected of when retrieving the values through the
	// protobuf reflection APIs.
	//
	// The value may only be populated if desc is also populated.
	Value interface{}

	// Raw is the raw encoded bytes for the extension field.
	Raw []byte
}

// Deprecated: Do not use.
type XXX_InternalExtensions extensionSyncMap

type extensionSyncMap struct {
	p *struct {
		mu sync.Mutex
		m  extensionMap
	}
}

func (m extensionSyncMap) Len() int {
	if m.p == nil {
		return 0
	}
	return m.p.m.Len()
}
func (m extensionSyncMap) Has(n int32) bool {
	if m.p == nil {
		return false
	}
	return m.p.m.Has(n)
}
func (m extensionSyncMap) Get(n int32) extensionField {
	if m.p == nil {
		return extensionField{}
	}
	return m.p.m.Get(n)
}
func (m *extensionSyncMap) Set(n int32, x extensionField) {
	if m.p == nil {
		m.p = new(struct {
			mu sync.Mutex
			m  extensionMap
		})
	}
	m.p.m.Set(n, x)
}
func (m extensionSyncMap) Clear(n int32) {
	m.p.m.Clear(n)
}
func (m extensionSyncMap) Range(f func(int32, extensionField) bool) {
	if m.p == nil {
		return
	}
	m.p.m.Range(f)
}

func (m extensionSyncMap) HasInit() bool {
	return m.p != nil
}
func (m extensionSyncMap) Lock() {
	m.p.mu.Lock()
}
func (m extensionSyncMap) Unlock() {
	m.p.mu.Unlock()
}

type extensionMap map[int32]extensionField

func (m extensionMap) Len() int {
	return len(m)
}
func (m extensionMap) Has(n int32) bool {
	_, ok := m[n]
	return ok
}
func (m extensionMap) Get(n int32) extensionField {
	return m[n]
}
func (m *extensionMap) Set(n int32, x extensionField) {
	if *m == nil {
		*m = make(map[int32]extensionField)
	}
	(*m)[n] = x
}
func (m *extensionMap) Clear(n int32) {
	delete(*m, n)
}
func (m extensionMap) Range(f func(int32, extensionField) bool) {
	for n, x := range m {
		if !f(n, x) {
			return
		}
	}
}

var globalLock sync.Mutex

func (m extensionMap) HasInit() bool {
	return m != nil
}
func (m extensionMap) Lock() {
	globalLock.Lock()
}
func (m extensionMap) Unlock() {
	globalLock.Lock()
}
