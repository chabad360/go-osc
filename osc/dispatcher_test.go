package osc

import (
	"net"
	"testing"
)

func TestDispatcher_AddMethodFunc(t *testing.T) {
	type args struct {
		addr   string
		method MethodFunc
	}
	tests := []struct {
		name    string
		methods map[string]Method
		args    args
		wantErr bool
	}{
		{"valid", nil, args{"/address/test", func(_ *Message) {}}, false},
		{"invalid", nil, args{"/address*/test", func(_ *Message) {}}, true},
		{"already_exists", map[string]Method{"/address/test": MethodFunc(func(_ *Message) {})}, args{"/address/test", func(_ *Message) {}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dispatcher{
				methods: tt.methods,
			}
			if err := d.AddMethodFunc(tt.args.addr, tt.args.method); (err != nil) != tt.wantErr {
				t.Errorf("AddMethodFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var testDispatcher = &Dispatcher{
	methods: map[string]Method{
		"/osc": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 1
			}
		}(),
		"/os": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 2
			}
		}(),
		"/osv": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 4
			}
		}(),
		"/osabc": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 8
			}
		}(),

		"/osc123": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 16
			}
		}(),
		"/osc1b3": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 32
			}
		}(),
		"/oscz": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 64
			}
		}(),
		"/osc/z": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 128
			}
		}(),
		"/osc/23f": func() MethodFunc {
			return func(msg *Message) {
				msg.Arguments[0] = msg.Arguments[0].(int) + 256
			}
		}(),
	},
}

func TestDispatcher_Dispatch(t *testing.T) { // TODO: somehow test bundles
	type args struct {
		packet Packet
		a      net.Addr
	}
	tests := []struct {
		name   string
		args   args
		expect int
	}{
		{"single", args{NewMessage("/osc", 0), nil}, 1},
		{"c_or_not", args{NewMessage("/os{c,}", 0), nil}, 3},
		{"single_any", args{NewMessage("/os{?,}", 0), nil}, 7},
		{"single_must", args{NewMessage("/os{c,v}", 0), nil}, 5},
		{"match_in_part", args{NewMessage("/osc{?,}z", 0), nil}, 64},
		{"match_multiple_parts", args{NewMessage("/osc/?", 0), nil}, 128},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDispatcher.Dispatch(tt.args.packet, tt.args.a)
			//time.Sleep(time.Second)
			p := tt.args.packet.(*Message)
			if p.Arguments[0].(int) != tt.expect {
				t.Errorf("Dispatch() got = %v, expect %v", p.Arguments[0].(int), tt.expect)
			}
		})
	}
}
