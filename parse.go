package main

import (
	"bytes"
	//"errors"
	"fmt"
	"io"
)

var (
	tagEBML = []byte{0x1a, 0x45, 0xdf, 0xa3}
	tagSegment = []byte{0x18, 0x53, 0x80, 0x67}
	tagSeek = []byte{0x11, 0x4d, 0x9b, 0x74}
	tagCluster = []byte{0x1f, 0x43, 0xb6, 0x75}
	tagVoid = []byte{0xec}
	tagInfo = []byte{0x15, 0x49, 0xa9, 0x66}
	tagTrack = []byte{0x16, 0x54, 0xae, 0x6b}
	tagTagging = []byte{0x12, 0x54, 0xc3, 0x67}
)

type Webm struct {
	EBML []byte
	SegmentInfo []byte
	MetaSeekInfo []byte
	Tagging []byte
	Track []byte
	Void []byte
	SegmentTag []byte
	initSegment []byte
	ClusterChannel chan *[]byte
	io.Reader
}

func (w *Webm) Parse() {
	
	initSegment := []byte{}
	//EBML
	tag := make([]byte, len(tagEBML))
	element := w.getElement(tag)
	if !bytes.Equal(tag, tagEBML) {
		fmt.Println("tagEBML")
		return
	}
	w.EBML = element
	initSegment = append(initSegment, element...)

	//SgmentTag
	tag = make([]byte, len(tagSegment))
	io.ReadFull(w, tag)
	if !bytes.Equal(tag, tagSegment) {
		fmt.Println("tagSegment")
		return
	}
	_, _, offset, length := w.getSize()
	element = append(tag, offset...)
	element = append(element, length...)
	w.SegmentTag = element
	initSegment = append(initSegment, element...)

	//MetaSeekInfo
	tag = make([]byte, len(tagSeek))
	element = w.getElement(tag)
	if !bytes.Equal(tag, tagSeek) {
		fmt.Println("tagSeek")
		return
	}
	w.MetaSeekInfo = element
	initSegment = append(initSegment, element...)

	//void
	tag = make([]byte, len(tagVoid))
	element = w.getElement(tag)
	if !bytes.Equal(tag, tagVoid) {
		fmt.Println("tagVoid")
		return
	}
	w.Void = element
	initSegment = append(initSegment, element...)

	//segment info
	tag = make([]byte, len(tagInfo))
	element = w.getElement(tag)
	if !bytes.Equal(tag, tagInfo) {
		fmt.Println("tagInfo")
		return
	}
	w.SegmentInfo = element
	initSegment = append(initSegment, element...)
	
	//track
	tag = make([]byte, len(tagTrack))
	element = w.getElement(tag)
	if !bytes.Equal(tag, tagTrack) {
		fmt.Println("tagTrack")
		return
	}
	w.Track = element
	initSegment = append(initSegment, element...)

	//tagging
	tag = make([]byte, len(tagTagging))
	element = w.getElement(tag)
	if !bytes.Equal(tag, tagTagging) {
		fmt.Println("tagTagging")
		return
	}
	w.Tagging = element
	initSegment = append(initSegment, element...)
	w.initSegment = initSegment

	w.ClusterChannel <- &initSegment
	for {
		tag := make([]byte, len(tagCluster))
		element := w.getElement(tag)
		if !bytes.Equal(tag, tagCluster) {
			fmt.Println("tagCluster")
			return
		}
		w.ClusterChannel <- &element
	}
}

func (w *Webm) getElement(tag []byte) []byte {
	io.ReadFull(w, tag)
	l, _, offset, length := w.getSize()
	element := append(tag, offset...)
	element = append(element, length...)
	buf := make([]byte, l)
	n, err := io.ReadFull(w, buf)
	if err != nil || n == 0 {
		fmt.Println(n, err)
	}
	element = append(element, buf...)
	return element
}

func (w *Webm) getSize() (int, int, []byte, []byte) {
	n := make([]byte, 1)
	io.ReadFull(w, n)
	j := 0
	for i := 0; i < 8; i++ {
		if (n[0] >> (7-uint8(i))) > 0 {
			j = i
			break
		}
	}
	d := make([]byte, j)
	io.ReadFull(w, d)
	k := 0
	for i := 0; i < j; i++ {
		k += int(uint64(d[i]) << uint8((j - 1 - i)*8))
	}
	return k, j, n, d
}