package regionquery

import (
	"io"
	"os"
	"sync"
	"testing"
)

var rootRegionData = []byte("世界\tWorld")
var rootRegionConf = &Conf{
	{Key: "continent", Width: 1},
	{Key: "country", Width: 2},
	{Key: "province", Width: 2},
	{Key: "city", Width: 2},
	{Key: "district", Width: 2},
}

func getFileReader(file string) io.Reader {
	reader, _ := os.Open(file)
	return reader
}

func newRegion() *Region {
	return NewRegion(rootRegionData, rootRegionConf, nil, nil)
}

func TestRegion_Load(t *testing.T) {
	rgn := newRegion()

	// test load data file
	df := "testdata/test_10.txt"
	err := rgn.Load(getFileReader(df))
	if err != nil {
		t.Fatalf("load exists data file failed: %s.", err)
	}

	// test reload data file
	df = "testdata/test_100.txt"
	err = rgn.ReLoad(getFileReader(df))
	if err != nil {
		t.Fatalf("reload data file failed: %s.", err)
	}

	// test load not exists data file
	df = "testdata/not_exists.data"
	err = rgn.Load(getFileReader(df))
	if err == nil {
		t.Fatalf("load not exists data file must be failed.")
	}
}

var findCases = []struct {
	code string
	err  error
	data []byte
}{
	{"1", nil, []byte("亚洲\tAsia")},
	{"101", nil, []byte("中国")},
	{"10101", nil, []byte("北京")},
	{"1010101", nil, []byte("北京")},
	{"101010101", nil, []byte("东城")},
	{"101010116", nil, []byte("延庆")},
	{"101020116", nil, []byte("蓟县")},
	{"101020117", ErrorRegionNotFound, []byte("")},
}

func TestRegion_Find(t *testing.T) {
	rgn := newRegion()
	df := "testdata/test_100.txt"
	err := rgn.Load(getFileReader(df))
	if err != nil {
		t.Fatalf("load exists data file failed: %s.", err)
	}

	for _, c := range findCases {
		r, err := rgn.Find(c.code)
		if err != c.err {
			t.Errorf("find data for %s expected error: %v, got: %v.", c.code, c.err, err)
		}

		if err == nil && string(r.Data) != string(c.data) {
			t.Errorf("find data for %s failed. expected: %s, got: %s.", c.code, string(c.data), string(r.Data))
		}
	}
}

func TestRegion_Parallel_Find(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup

	rgn := newRegion()
	df := "testdata/test_100.txt"
	err := rgn.Load(getFileReader(df))
	if err != nil {
		t.Fatalf("load exists data file failed: %s.", err)
	}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for _, c := range findCases {
				r, err := rgn.Find(c.code)
				if err != c.err {
					t.Errorf("find data for %s expected error: %v, got: %v.", c.code, c.err, err)
				}

				if err == nil && string(r.Data) != string(c.data) {
					t.Errorf("find data for %s failed. expected: %s, got: %s.", c.code, string(c.data), string(r.Data))
				}
			}
		}()
	}

	wg.Wait()
}

func TestRegion_dismantleCode(t *testing.T) {
	var cases = []struct {
		code   string
		err    error
		pieces []string
	}{
		{"1", nil, []string{"1"}},
		{"101", nil, []string{"1", "01"}},
		{"10101", nil, []string{"1", "01", "01"}},
		{"1010101", nil, []string{"1", "01", "01", "01"}},
		{"101010101", nil, []string{"1", "01", "01", "01", "01"}},
		{"101010116", nil, []string{"1", "01", "01", "01", "16"}},
		{"101020116", nil, []string{"1", "01", "02", "01", "16"}},
		{"1010201161", ErrorRegionCodeIncorrect, []string{}},
		{"10102011", ErrorRegionCodeIncorrect, []string{}},
	}

	rgn := newRegion()
	for _, c := range cases {
		pieces, err := rgn.dismantleCode(c.code)
		if err != c.err {
			t.Errorf("dismantle code for %s expected error: %v, got: %v.", c.code, c.err, err)
		}

		if err == nil && len(pieces) != len(c.pieces) {
			t.Errorf("dismantle code for %s failed. expected %v, got %v.", c.code, c.pieces, pieces)
		}
	}
}

func BenchmarkRegion_Load(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rgn := newRegion()
		err := rgn.Load(getFileReader("testdata/china_level_5.txt"))
		if err != nil {
			b.Fatalf("load exists data file failed: %s.", err)
		}
	}
}

func BenchmarkRegion_Find(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	rgn := newRegion()
	err := rgn.Load(getFileReader("testdata/china_level_5.txt"))
	if err != nil {
		b.Fatalf("load exists data file failed: %s.", err)
	}

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		rgn.Find("101260503")
		b.StopTimer()
	}
}
