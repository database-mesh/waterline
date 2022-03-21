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

type TcAct struct {
	linkIdx        int
	attrs          netlink.QdiscAttrs
	qdisc          netlink.Qdisc
	totalBandWidth string
}

func NewTcAct(iface string, totalBandWidth string) (*TcAct, error) {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, err
	}

	attrs := netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	}

	return &TcAct{
		linkIdx:        link.Attrs().Index,
		attrs:          attrs,
		totalBandWidth: totalBandWidth,
	}, nil
}

func (t *TcAct) AddHtbQdisc() error {
	t.qdisc = netlink.NewHtb(t.attrs)
	return netlink.QdiscAdd(t.qdisc)
}

func (t *TcAct) DeleteHtbQdisc() error {
	return netlink.QdiscDel(t.qdisc)
}

func (t *TcAct) AddClasses(qos v1alpha1.SQLTrafficQoS) error {
	rules := qos.Spec.Groups
	// sort by classid or rate
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ClassId < rules[j].ClassId
	})

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

func (t *TcAct) addRootHandle() error {
	attrs := netlink.QdiscAttrs{
		LinkIndex: t.linkIdx,
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
	return netlink.ClassAdd(class)
}

func (t *TcAct) addClass(idx int, rule v1alpha1.TrafficQoSGroup) error {
	attrs := netlink.ClassAttrs{
		LinkIndex: t.linkIdx,
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

	if rule.Ceil != "" {
		ceilValue, err := resource.ParseQuantity(rule.Ceil)
		if err != nil {
			return err
		}

		htbClassAttrs.Ceil = uint64(ceilValue.Value())
	}

	class := netlink.NewHtbClass(attrs, htbclassattrs)
	return netlink.ClassAdd(class)
}

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
