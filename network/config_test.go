package network

import "testing"

func Test_parse(t *testing.T) {
	node := Node{
		Name:      "TEST",
		Subnet:    "1.2.3.4/32",
		Address:   []Address{{Host: "127.0.0.1", Port: 321}, {Host: "127.0.0.1", Port: 1223}},
		PublicKey: "---\nXXX\n---",
	}
	text, err := node.MarshalText()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(text))

	var cp Node
	err = cp.UnmarshalText(text)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", cp)
}
