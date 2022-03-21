// Copyright 2022 Database Mesh Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tc

import (
	v1alpha1 "github.com/database-mesh/waterline/api/v1alpha1"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"k8s.io/apimachinery/pkg/api/resource"
	"sort"
)

type Shaper struct {
	qos            v1alpha1.SQLTrafficQoS
	link           netlink.Link
	totalBandWidth string
}

func NewTcShaper(qos v1alpha1.SQLTrafficQoS, totalBandWidth string) (*Shaper, error) {
	link, err := netlink.LinkByName(qos.Spec.NetworkDevice)
	if err != nil {
		return nil, err
	}

	return &Shaper{
		qos:            qos,
		link:           link,
		totalBandWidth: totalBandWidth,
	}, nil
}

// add htb qidsc, called by AddClasses
func (t *Shaper) addHtbQdisc() error {
	attrs := netlink.QdiscAttrs{
		LinkIndex: t.link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	}

	qdisc := netlink.NewHtb(attrs)
	return netlink.QdiscReplace(qdisc)
}

// add htb root handle, called by AddClasses
func (t *Shaper) addRootHandle() error {
	attrs := netlink.ClassAttrs{
		LinkIndex: t.link.Attrs().Index,
		Parent:    netlink.MakeHandle(1, 0),
		Handle:    netlink.MakeHandle(1, 1),
	}

	bandValue, err := resource.ParseQuantity(t.totalBandWidth)
	if err != nil {
		return err
	}

	htbClassAttrs := netlink.HtbClassAttrs{
		Rate: uint64(bandValue.Value()),
	}

	class := netlink.NewHtbClass(attrs, htbClassAttrs)
	return netlink.ClassReplace(class)
}

func (t *Shaper) AddClasses() error {
	if err := t.addHtbQdisc(); err != nil {
		return err
	}

	if err := t.addRootHandle(); err != nil {
		return err
	}

	rules := t.qos.Spec.Groups
	// sort by rate
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Rate < rules[j].Rate
	})

	for idx, rule := range rules {
		if err := t.addClass(idx, rule); err != nil {
			return err
		}
	}

	return nil
}

// add htb class
func (t *Shaper) addClass(idx int, rule v1alpha1.TrafficQoSGroup) error {
	attrs := netlink.ClassAttrs{
		LinkIndex: t.link.Attrs().Index,
		Parent:    netlink.MakeHandle(1, 1),
		//exclude 0, 1
		Handle: netlink.MakeHandle(1, uint16(idx+2)),
	}

	rateValue, err := resource.ParseQuantity(rule.Rate)
	if err != nil {
		return err
	}

	htbClassAttrs := netlink.HtbClassAttrs{
		Rate: uint64(rateValue.Value()),
	}

	switch t.qos.Spec.QoSClass {
	case v1alpha1.QoSClassGuaranteed:
		if rule.Ceil == "" {
			ceilValue, err := resource.ParseQuantity(rule.Rate)
			if err != nil {
				return err
			}
			htbClassAttrs.Ceil = uint64(ceilValue.Value())
		}
	case v1alpha1.QoSClassBurstable:
		if rule.Ceil == "" {
			htbClassAttrs.Ceil = uint64(float64(rateValue.Value()) * 1.5)
		}
	case v1alpha1.QoSClassBestEffort:
		return nil
	}

	class := netlink.NewHtbClass(attrs, htbClassAttrs)
	return netlink.ClassReplace(class)
}

// add qdisc clsact
func (t *Shaper) addClsact() error {
	attrs := netlink.QdiscAttrs{
		LinkIndex: t.link.Attrs().Index,
		Handle:    netlink.MakeHandle(0xffff, 0),
		Parent:    netlink.HANDLE_CLSACT,
	}

	qdisc := netlink.GenericQdisc{
		QdiscAttrs: attrs,
		QdiscType:  "clsact",
	}

	return netlink.QdiscReplace(&qdisc)
}

// match qdisc func
type matchQdiscFunc = func(qdisc netlink.Qdisc) bool

// delete matched qdisc
func (t *Shaper) delQdisc(f matchQdiscFunc) error {
	qdiscs, err := netlink.QdiscList(t.link)

	if err != nil {
		return err
	}

	for _, v := range qdiscs {
		if f(v) {
			return netlink.QdiscDel(v)
		}
	}

	return nil
}

// match class func
type matchClassFunc = func(class netlink.Class) bool

// delete matched class
func (t *Shaper) delClass(f matchClassFunc) error {
	classes, err := t.ListClass()
	if err != nil {
		return err
	}

	for _, v := range classes {
		if f(v) {
			return netlink.ClassDel(v)
		}
	}

	return nil
}

// ListClass list class
func (t *Shaper) ListClass() ([]netlink.Class, error) {
	return netlink.ClassList(t.link, netlink.MakeHandle(1, 0))
}

// AddFilter add bpf filter, default obj name is "tc.o"
func (t *Shaper) AddFilter() error {
	filterAttrs := netlink.FilterAttrs{
		LinkIndex: t.link.Attrs().Index,
		Parent:    netlink.MakeHandle(1, 0),
		Protocol:  unix.ETH_P_ALL,
	}

	bpfFilter := netlink.BpfFilter{
		FilterAttrs:  filterAttrs,
		ClassId:      netlink.MakeHandle(1, 0),
		Name:         "tc.o",
		DirectAction: true,
	}

	return netlink.FilterAdd(&bpfFilter)
}
