package regionquery

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

const (
	regionFieldCount = 2
)

var (
	ErrorRegionNotFound      = errors.New("region not found")
	ErrorRegionCodeIncorrect = errors.New("region code incorrect")
)

// TODO map并发读写
type Region struct {
	Data []byte
	Conf *Conf

	Sup *Region
	Sub map[string]*Region
}

func NewRegion(data []byte, conf *Conf, sup *Region, sub map[string]*Region) *Region {
	if sub == nil {
		sub = map[string]*Region{}
	}
	return &Region{
		Data: data,
		Conf: conf,
		Sup:  sup,
		Sub:  sub,
	}
}

func (rgn *Region) Load(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		item := strings.SplitN(line, "\t", regionFieldCount)
		if len(item) != regionFieldCount {
			continue
		}

		code, data := item[0], []byte(item[1])
		pieces, err := rgn.dismantleCode(code)
		if err != nil {
			continue
		}

		nRgn := rgn
		for _, p := range pieces {
			subRgn, ok := nRgn.Sub[p]
			if !ok {
				subRgn = NewRegion(data, nil, nRgn, nil)
				nRgn.Sub[p] = subRgn
			}
			nRgn = subRgn
		}
		if string(nRgn.Data) != string(data) {
			nRgn.Data = data
		}
	}

	return scanner.Err()
}

func (rgn *Region) ReLoad(r io.Reader) error {
	nRgn := NewRegion(rgn.Data, rgn.Conf, nil, nil)
	err := nRgn.Load(r)
	if err != nil {
		return err
	}

	*rgn = *nRgn
	return nil
}

func (rgn *Region) Find(code string) (*Region, error) {
	pieces, err := rgn.dismantleCode(code)
	if err != nil {
		return nil, err
	}

	nRgn := rgn
	for _, p := range pieces {
		subRgn, ok := nRgn.Sub[p]
		if !ok {
			return nil, ErrorRegionNotFound
		}
		nRgn = subRgn
	}
	return nRgn, nil
}

func (rgn *Region) dismantleCode(code string) ([]string, error) {
	cl, nw := len(code), 0
	conf := *rgn.Conf

	pieces := []string{}
	for _, cnf := range conf {
		tnw := nw + cnf.Width
		if tnw > cl {
			break
		}

		pieces = append(pieces, code[nw:tnw])
		nw = tnw
	}

	if cl == 0 || cl != nw {
		return nil, ErrorRegionCodeIncorrect
	}
	return pieces, nil
}
