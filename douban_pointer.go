/*
 Copyright (c) 2017 mattn.jp@gmail.com

 Permission is hereby granted, free of charge, to any person obtaining a copy of this 
 software and associated documentation files (the "Software"), to deal in the Software 
 without restriction, including without limitation the rights to use, copy, modify, 
 merge, publish, distribute, sub license, and/or sell copies of the Software, and 
 to permit persons to whom the Software is furnished to do so, 
 subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all copies 
 or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING 
 BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,FITNESS FOR A PARTICULAR PURPOSE AND NON 
 INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, 
 DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */


/* 
 pointer.go is a derivative from github.com/mattn/pointer and modified for removing the mutex 
 because in PostgreSQL, the session is a session-per-process model. there is no chance that
 multi-thread accessing the same map 
*/

package main

// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

var (
	store = map[unsafe.Pointer]interface{}{}
)

func Save(v interface{}) unsafe.Pointer {
	if v == nil {
		return nil
	}

	// Generate real fake C pointer.
	// This pointer will not store any data, but will bi used for indexing purposes.
	// Since Go doest allow to cast dangling pointer to unsafe.Pointer, we do rally allocate one byte.
	// Why we need indexing, because Go doest allow C code to store pointers to Go data.
	var ptr unsafe.Pointer = C.malloc(C.size_t(1))
	if ptr == nil {
		panic("can't allocate 'cgo-pointer hack index pointer': ptr == nil")
	}

	store[ptr] = v

	return ptr
}

func Restore(ptr unsafe.Pointer) (v interface{}) {
	if ptr == nil {
		return nil
	}

	v = store[ptr]
	return
}

func Unref(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}

	delete(store, ptr)

	C.free(ptr)
}

