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

#include <linux/bpf.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <arpa/inet.h>
#include <linux/if_ether.h>
#include "include/bpf_helpers.h"
#include "include/bpf_endian.h"

struct bpf_map_def SEC("maps") filter_helper = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u8),
    .value_size = sizeof(__u16),
    .max_entries = 2,
};

struct bpf_map_def SEC("maps") buf = {
	.type = BPF_MAP_TYPE_ARRAY,
	.key_size = sizeof(__u32),
	.value_size = sizeof(__u8),
	.max_entries = 1<<16,
};

struct bpf_map_def SEC("maps") jmp_table = {
	.type = BPF_MAP_TYPE_PROG_ARRAY,
	.key_size = sizeof(__u32),
	.value_size = sizeof(__u32),
	.max_entries = 100,
};

struct event {
	__u8 seq;
    __u16 sport;
    __u16 dport;
    __u32 saddr;
    __u32 daddr;
	__u32 my_pkt_len;
	__u32 payload_offset;
	__u32 class_id;
};

struct bpf_map_def SEC("maps") my_pkt = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u8),
	.value_size = sizeof(struct event),
	.max_entries = 1,
};

struct bpf_map_def SEC("maps") my_pkt_evt = {
    .type = BPF_MAP_TYPE_RINGBUF,
    .max_entries = 4096,
};


SEC("socket")
int sql_filter(struct __sk_buff *skb) {
    if (skb->protocol != bpf_htons(ETH_P_IP)) {
        return 0;
    }

    struct iphdr iph;
    bpf_skb_load_bytes(skb, ETH_HLEN, &iph, sizeof(iph));

    if (iph.protocol != IPPROTO_TCP) {
    	return 0;
    }

    __u32 ip_hlen = iph.ihl << 2;

    struct tcphdr tcph;

    bpf_skb_load_bytes(skb, ETH_HLEN + sizeof(iph), &tcph, sizeof(tcph));

    __u32 tcp_hlen = tcph.doff << 2;

    __u8 filter_port_key = 0;
    __u16 *port = bpf_map_lookup_elem(&filter_helper, &filter_port_key);
    if (!port) return 0;

    if (bpf_ntohs(tcph.dest) != *port) {
    	return 0;
    }

    __u32 payload_offset = ETH_HLEN + ip_hlen + tcp_hlen;

    __u8 header[5];
    bpf_skb_load_bytes(skb, payload_offset, &header, sizeof(header));

    if (header[4] != 3) {
    	return 0;
    }

    __u32 my_pkt_len = header[0] | header[1] << 8 | header[2] << 16;

    struct event evt;
	__builtin_memset(&evt, 0, sizeof(event));
	evt.saddr = iph.saddr;
	evt.sport = tcph.source,
	evt.daddr = iph.daddr,
	evt.dport = tcph.dest,
	evt.seq = header[3];
	evt.pkt_len = my_pkt_len;
	evt.payload_offset = payload_offset;
	evt.class_id = 0;

    __u8 evt_key = 0;

    bpf_map_update_elem(&my_pkt, &evt_key, &evt, BPF_ANY);

    struct event evt;
    __u8 b;
    int i = 0;
    __u32 buf_idx = 0;
    #pragma unroll
    for (i = 0; i < 200; i++) {
	    if (buf_idx == my_pkt_len - 1) {
    		bpf_ringbuf_output(&my_pkt_evt, &evt, sizeof(evt), 0);
		    return -1;
	    }
    	if (bpf_skb_load_bytes(skb, payload_offset+5+i, &b, sizeof(b)) == 0) {
        	if (bpf_map_update_elem(&buf, &buf_idx, &b, BPF_ANY) == 0) {
        		buf_idx++;
        	}
        }
    }

    bpf_tail_call(skb, &jmp_table, 0);

    return -1;
}

SEC("socket_1")
int sql_filter_1(struct __sk_buff *skb) {
    __u8 evt_key = 0;

    struct event *evt = bpf_map_lookup_elem(&my_pkt, &evt_key);

    if (!evt) return -1;

    __u8 b;
    __u32 buf_idx = 200;
    #pragma unroll
    for (int i = 200; i < 2048; i++) {
        if (buf_idx == *my_pkt_len - 1) {
            bpf_ringbuf_output(&my_pkt_evt, evt, sizeof(*evt), 0);
            return -1;
        }

    	if (bpf_skb_load_bytes(skb, *payload_offset+5+i, &b, sizeof(b)) == 0) {
		    if (bpf_map_update_elem(&buf, &buf_idx, &b, BPF_ANY) == 0) {
			    buf_idx++;
		    }
	    }
    }

    return -1;
}

char LICENSE[] SEC("license") = "GPL";