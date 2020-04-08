package network

import "testing"

func Test_parse(t *testing.T) {
	node := Node{
		Name:      "TEST",
		Subnet:    "1.2.3.4/32",
		Address:   []Address{{Host: "127.0.0.1", Port: 321}, {Host: "127.0.0.1", Port: 1223}},
		PublicKey: "---\nXXX\n---",
	}
	text, err := node.Build()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(text))

	var cp Node
	err = cp.Parse(text)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", cp)

	cfg := Config{
		Name:      "XYZ",
		Port:      123,
		Interface: "tinc0",
		AutoStart: true,
		ConnectTo: []string{"Alfa", "Beta"},
	}

	text, err = cfg.Build()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(text))

	cfg = Config{}

	err = cfg.Parse(text)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", cfg)
}
