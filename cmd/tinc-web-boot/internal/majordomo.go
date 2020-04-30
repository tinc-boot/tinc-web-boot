package internal

import "encoding/hex"

//go:generate msgp
//msgp:tuple Share

type Share struct {
	Port      uint16
	Subnet    string
	Addresses [][4]byte
	Network   string
	Code      string
}

func (z *Share) ToHex() string {
	data, err := z.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(data)
}

func (z *Share) FromHex(txt string) error {
	data, err := hex.DecodeString(txt)
	if err != nil {
		return err
	}
	_, err = z.UnmarshalMsg(data)
	return err
}
