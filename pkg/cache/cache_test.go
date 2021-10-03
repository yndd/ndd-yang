package cache

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
)

func TestGetNotificationFromUpdate(t *testing.T) {
	target := "dev1"
	origin := "test"
	x := map[string]interface{}{"admin-state": "enable", "rir-name": "rfc1918"}
	b, _ := json.Marshal(x)
	tests := []struct {
		inp *gnmi.Update
		exp interface{}
	}{
		{
			inp: &gnmi.Update{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "ipam"},
						{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "10.0.0.0/8", "network-instance": "default"}},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: b}},
			},
			//exp: "ipam",
		},
	}
	c := NewCache([]string{target})
	for _, tt := range tests {
		n, err := c.GetNotificationFromUpdate(target, origin, tt.inp)
		if err != nil {
			t.Errorf("GetNotificationFromUpdate: %v\n", err)
		}
		for _, u := range n.GetUpdate() {
			fmt.Printf("Update: %v\n", u)
		}
	}
}

func TestGetJson(t *testing.T) {
	target := "dev1"
	origin := "test"
	tests := []struct {
		inp *gnmi.Path
		exp interface{}
	}{
		{
			inp: &gnmi.Path{
				Origin: origin,
				Elem:   []*gnmi.PathElem{},
			},
			//exp: "ipam",
		},
		{
			inp: &gnmi.Path{
				Origin: origin,
				Elem: []*gnmi.PathElem{
					{Name: "ipam"},
				},
			},
			//exp: "aggregate",
		},

		{
			inp: &gnmi.Path{
				Origin: origin,
				Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "10.0.0.0/8", "network-instance": "default"}},
				},
			},
			//exp: map[string]interface {}{"Prefix":"10.0.0.0/8", "admin-state":"enable", "network-instance":"default", "rir-name":"rfc1918", "tenant":"default"},
		},
	}

	n := &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix: &gnmi.Path{
			Target: target,
			Origin: origin,
			//Elem:   []*gnmi.PathElem{{Name: "a"}, {Name: "b", Key: map[string]string{"key": "value"}}},
		},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "10.0.0.0/8", "network-instance": "default"}},
					{Name: "admin-state"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "enable"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "10.0.0.0/8", "network-instance": "default"}},
					{Name: "rir-name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "10.0.0.0/8", "network-instance": "default"}},
					{Name: "tenant"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "10.0.0.0/8", "network-instance": "default"}},
					{Name: "network-instance"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "10.0.0.0/8", "network-instance": "default"}},
					{Name: "Prefix"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "10.0.0.0/8"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.64.0.0/16", "network-instance": "default"}},
					{Name: "admin-state"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "enable"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.64.0.0/16", "network-instance": "default"}},
					{Name: "rir-name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.64.0.0/16", "network-instance": "default"}},
					{Name: "tenant"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.64.0.0/16", "network-instance": "default"}},
					{Name: "network-instance"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.64.0.0/16", "network-instance": "default"}},
					{Name: "Prefix"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "100.64.0.0/16"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.112.0.0/16", "network-instance": "default"}},
					{Name: "admin-state"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "enable"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.112.0.0/16", "network-instance": "default"}},
					{Name: "rir-name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.112.0.0/16", "network-instance": "default"}},
					{Name: "tenant"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.112.0.0/16", "network-instance": "default"}},
					{Name: "network-instance"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "100.112.0.0/16", "network-instance": "default"}},
					{Name: "Prefix"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "100.112.0.0/16"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "3100::/16", "network-instance": "default"}},
					{Name: "admin-state"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "enable"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "3100::/16", "network-instance": "default"}},
					{Name: "rir-name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "3100::/16", "network-instance": "default"}},
					{Name: "tenant"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "3100::/16", "network-instance": "default"}},
					{Name: "network-instance"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "3100::/16", "network-instance": "default"}},
					{Name: "Prefix"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "3100::/16"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "rir", Key: map[string]string{"name": "rfc1918"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "rir", Key: map[string]string{"name": "rfc6598"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc6598"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "rir", Key: map[string]string{"name": "ula"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "ula"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "rir", Key: map[string]string{"name": "ripe"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "ripe"}},
			},
		},
	}

	c := NewCache([]string{target})
	if err := c.GnmiUpdate(target, n); err != nil {
		t.Errorf("GnmiUpdate: %v\n", err)
	}

	for _, tt := range tests {
		/*
			pp := path.ToStrings(tt.inp, true)
			fmt.Printf("pp %v\n", pp)
			c.GetCache().Query(target, pp, func(_ []string, _ *ctree.Leaf, val interface{}) error {
				fmt.Printf("val %v\n", val)
				return nil
			})
		*/

		d, err := c.GetJson(target, tt.inp)
		if err != nil {
			t.Errorf("GetConfig: %v\n", err)
		}
		//fmt.Println(d)
		jsonString, _ := json.MarshalIndent(d, "", "\t")
		fmt.Println(string(jsonString))

		fmt.Printf("out1: %#v\n", d)
		fmt.Printf("out2: %#v\n", tt.exp)

		//if !reflect.DeepEqual(d, tt.exp) {
		//	t.Errorf("sortedVals(%v):\n got  %v\n want %v\n", tt.inp, d, tt.exp)
		//}
	}
}
