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
#include "headers/bpf_helpers.h"
#include "headers/bpf_endian.h"

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

struct event_key {
	 __u8 seq;
    __u16 sport;
    __u16 dport;
    __u32 saddr;
    __u32 daddr;
};

struct event_value {
	__u32 payload_offset;
	__u32 my_pkt_len;
	__u32 class_id;
};

struct bpf_map_def SEC("maps") my_pkt = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(struct event_key),
	.value_size = sizeof(struct event_value),
	.max_entries = 2048,
};

// used by tail call
struct bpf_map_def SEC("maps") tmp_pkt_evt = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u8),
    .value_size = sizeof(struct event_key),
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

    struct event_key evt_key;
	__builtin_memset(&evt_key, 0, sizeof(struct event_key));
	evt_key.saddr = iph.saddr;
	evt_key.sport = tcph.source,
	evt_key.daddr = iph.daddr,
	evt_key.dport = tcph.dest,
	evt_key.seq = header[3];


	struct event_value evt_value;
	__builtin_memset(&evt_value, 0, sizeof(struct event_value));
	evt_value.payload_offset = payload_offset;
	evt_value.my_pkt_len = my_pkt_len;
	evt_value.class_id = 0;

    bpf_map_update_elem(&my_pkt, &evt_key, &evt_value, BPF_ANY);

    __u8 tmp_key = 0;
    bpf_map_update_elem(&tmp_pkt_evt, &tmp_key, &evt_key, BPF_ANY);

    __u8 b;
    int i = 0;
    __u32 buf_idx = 0;
    #pragma unroll
    for (i = 0; i < 200; i++) {
	    if (buf_idx == my_pkt_len - 1) {
    		bpf_ringbuf_output(&my_pkt_evt, &evt_key, sizeof(struct event_key), 0);
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
    __u8 tmp_key = 0;

    struct event_key *tmp_evt_key = bpf_map_lookup_elem(&tmp_pkt_evt, &tmp_key);

    if (!tmp_key) return -1;

    struct event_value *tmp_value = bpf_map_lookup_elem(&tmp_pkt_evt, &tmp_evt_key);

    if (!tmp_value) return -1;

    __u32 payload_offset = tmp_value->payload_offset;
    __u32 my_pkt_len = tmp_value->my_pkt_len;
    __u8 b;
    __u32 buf_idx = 200;
    #pragma unroll
    for (int i = 200; i < 2048; i++) {
        if (buf_idx == my_pkt_len - 1) {
            bpf_ringbuf_output(&my_pkt_evt, tmp_value, sizeof(*tmp_value), 0);
            return -1;
        }

    	if (bpf_skb_load_bytes(skb, payload_offset+5+i, &b, sizeof(b)) == 0) {
		    if (bpf_map_update_elem(&buf, &buf_idx, &b, BPF_ANY) == 0) {
			    buf_idx++;
		    }
	    }
    }

    return -1;
}

char LICENSE[] SEC("license") = "GPL";