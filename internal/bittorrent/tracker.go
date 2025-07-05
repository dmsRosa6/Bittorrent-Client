package bittorrent


type Tracker struct{
	Address string
}

func (t *Tracker) New(address string) Tracker {
	return Tracker{Address: address}
}
func (t *Tracker) GetAddress() string {
	return t.Address;
}

func (t *Tracker) SetAddress(newAddress string) {
	t.Address = newAddress
}