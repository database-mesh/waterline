//  Copyright 2022 Database Mesh Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package bpf

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/mlycore/log"
)

const (
	SockFilter = "sock_filter.o"
	TcPktMap   = "/sys/fs/bpf/tc/globals/my_pkt"
)

// TODO: add loader to load this program to net dev

type Loader struct {
}

// TODO: add port

func (l *Loader) Load() error {
	m, err := l.LoadTcPkgMap()
	if err != nil {
		return err
	}

	if err := l.LoadSockFilter(m); err != nil {
		return err
	}

	return nil

}
func (l *Loader) LoadSockFilter(tcPkt *ebpf.Map) error {
	spec, err := ebpf.LoadCollectionSpec(SockFilter)
	if err != nil {
		return err
	}

	var objs Objs
	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		return err
	}

	if err = objs.JmpTable.Update(uint32(0), uint32(objs.TailProg1.FD()), ebpf.UpdateAny); err != nil {
		return fmt.Errorf("jmptable err: %s", err)
	}

	reader, err := ringbuf.NewReader(objs.MyPktEvt)
	if err != nil {
		return err
	}

	for {
		evt, query, err := l.ReadRecord(objs, reader)
		if err != nil {
			log.Warnln(err)
			continue
		}

		evt.ClassId = l.CalcQos(query)
		if err := tcPkt.Update(uint32(0), &evt, ebpf.UpdateAny); err != nil {
			log.Warnln(err)
		}
	}
	return nil
}
func (l *Loader) LoadTcPkgMap() (*ebpf.Map, error) {
	return ebpf.LoadPinnedMap(TcPktMap, nil)
}
func (l *Loader) CalcQoS(query string) uint32 {
	return 0
}
func (l *Loader) ReadRecord(objs Objs, reader *ringbuf.Reader) (Event, string, error) {
	record, err := reader.Read()
	if err != nil {
		return Event{}, "", fmt.Errorf("reading from reader: %s", err)
	}

	var evt Event
	if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &evt); err != nil {
		return Event{}, "", fmt.Errorf("parsing ringbuf event: %s", err)
	}

	var (
		key     uint32
		value   uint8
		entries = objs.Buf.Iterate()
		chars   = make([]byte, evt.PktLen)
	)

	// PktLen contains COM_TYPE 1byte, so evt.PktLen-1 here
	for i := 0; i < int(evt.PktLen-1); i++ {
		entries.Next(&key, &value)
		chars = append(chars, value)
	}

	return evt, string(chars), nil
}

// func load() error {
// 	m, err := loadTcPktMap()
// 	if err != nil {
// 		return err
// 	}

// 	if err := loadSockFilter(m); err != nil {
// 		return err
// 	}

// 	return nil
// }

type Objs struct {
	Prog         *ebpf.Program `ebpf:"sql_filter"`
	TailProg1    *ebpf.Program `ebpf:"sql_filter_1"`
	FilterHelper *ebpf.Map     `ebpf:"filter_helper"`
	Buf          *ebpf.Map     `ebpf:"buf"`
	JmpTable     *ebpf.Map     `ebpf:"jmp_table"`
	MyPktEvt     *ebpf.Map     `ebpf:"my_pkt_evt"`
}

// func loadSockFilter(tcPkt *ebpf.Map) error {
// 	spec, err := ebpf.LoadCollectionSpec(SockFilter)
// 	if err != nil {
// 		return err
// 	}

// 	var objs Objs
// 	if err := spec.LoadAndAssign(&objs, nil); err != nil {
// 		return err
// 	}

// 	if err = objs.JmpTable.Update(uint32(0), uint32(objs.TailProg1.FD()), ebpf.UpdateAny); err != nil {
// 		return fmt.Errorf("jmptable err: %s", err)
// 	}

// 	reader, err := ringbuf.NewReader(objs.MyPktEvt)
// 	if err != nil {
// 		return err
// 	}

// 	for {
// 		evt, query, err := readRecord(objs, reader)
// 		if err != nil {
// 			log.Warnln(err)
// 			continue
// 		}

// 		evt.ClassId = calcQos(query)
// 		if err := tcPkt.Update(uint32(0), &evt, ebpf.UpdateAny); err != nil {
// 			log.Warnln(err)
// 		}
// 	}
// }

// func loadTcPktMap() (*ebpf.Map, error) {
// 	return ebpf.LoadPinnedMap(TcPktMap, nil)
// }

// func calcQos(query string) uint32 {
// 	return 0
// }

type Event struct {
	Seq    uint8
	Sport  uint16
	Dport  uint16
	Saddr  uint32
	Daddr  uint32
	PktLen uint32
	// tcp payload offset
	Offset  uint32
	ClassId uint32
}

// readRecord read record from ringbuf
// func readRecord(objs Objs, reader *ringbuf.Reader) (Event, string, error) {
// 	record, err := reader.Read()
// 	if err != nil {
// 		return Event{}, "", fmt.Errorf("reading from reader: %s", err)
// 	}

// 	var evt Event
// 	if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &evt); err != nil {
// 		return Event{}, "", fmt.Errorf("parsing ringbuf event: %s", err)
// 	}

// 	var (
// 		key     uint32
// 		value   uint8
// 		entries = objs.Buf.Iterate()
// 		chars   = make([]byte, evt.PktLen)
// 	)

// 	// PktLen contains COM_TYPE 1byte, so evt.PktLen-1 here
// 	for i := 0; i < int(evt.PktLen-1); i++ {
// 		entries.Next(&key, &value)
// 		chars = append(chars, value)
// 	}

// 	return evt, string(chars), nil
// }
