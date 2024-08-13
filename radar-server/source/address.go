package source

import "fmt"

type Address struct {
	A1   uint8
	A2   uint8
	A3   uint8
	A4   uint8
	Port uint16
}

func NewAddress(a1 uint8, a2 uint8, a3 uint8, a4 uint8, port uint16) Address {
	return Address{A1: a1, A2: a2, A3: a3, A4: a4, Port: port}
}

func (a Address) IsAddressMatch(address2 Address) bool {
	if a.A1 == 0 && a.A2 == 0 && a.A3 == 0 && a.A4 == 0 {
		return true
	}
	return a.A1 == address2.A1 && a.A2 == address2.A2 && a.A3 == address2.A3 && a.A4 == address2.A4
}

func (a Address) IsPortMatch(address2 Address) bool {
	if a.Port == 0 {
		return true
	}
	return a.Port == address2.Port || a.Port == 0 || address2.Port == 0
}

func (a Address) IsMatch(address2 Address) bool {
	return a.IsAddressMatch(address2) && a.IsPortMatch(address2)
}

func (a Address) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", a.A1, a.A2, a.A3, a.A4, a.Port)
}
