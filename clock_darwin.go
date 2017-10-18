package gontpd

func (s *Service) setOffset(p *peer) (err error) {
	Info.Printf("set offset from :%s offset=%s", p.addr)
	return
}

func systemPrecision() int8 {
	return 1
}
