package tests

import (
	"testing"

	"github.com/mohae/deepcopy"
)

func BenchmarkCopySliceSlice(b *testing.B) {
	list := []string{"main", "access_combined", "/var/log/httpd/access_log", "$host$", `$ip$ $clientip$ - - [$ts$] "GET /product.screen?product_id=HolyGouda&JSESSIONID=SD3SL1FF7ADFF8 HTTP 1.1" $status$ $timetaken$ "http://shop.buttercupgames.com/cart.do?action=view&itemId=HolyGouda" "$useragent$" $size$`}
	listlist := make([][]string, 1000)
	for i := 0; i < 1000; i++ {
		listlist = append(listlist, list)
	}
	for n := 0; n < b.N; n++ {
		_ = copylist(listlist)
	}
}

func copylist(src [][]string) (dst [][]string) {
	dst = make([][]string, len(src))
	for i := range src {
		dst[i] = make([]string, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

func BenchmarkCopySliceSliceByte(b *testing.B) {
	list := [][]byte{[]byte("main"), []byte("access_combined"), []byte("/var/log/httpd/access_log"), []byte("$host$"), []byte(`$ip$ $clientip$ - - [$ts$] "GET /product.screen?product_id=HolyGouda&JSESSIONID=SD3SL1FF7ADFF8 HTTP 1.1" $status$ $timetaken$ "http://shop.buttercupgames.com/cart.do?action=view&itemId=HolyGouda" "$useragent$" $size$`)}
	listlist := make([][][]byte, 1000)
	for i := 0; i < 1000; i++ {
		listlist = append(listlist, list)
	}
	for n := 0; n < b.N; n++ {
		_ = copybytelist(listlist)
	}
}

func copybytelist(src [][][]byte) (dst [][][]byte) {
	dst = make([][][]byte, len(src))
	for i := range src {
		dst[i] = make([][]byte, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

func BenchmarkCopySliceMap(b *testing.B) {
	m := map[string]string{"index": "main",
		"sourcetype": "access_combined",
		"source":     "/var/log/httpd/access_log",
		"host":       "$host$",
		"_raw":       `$ip$ $clientip$ - - [$ts$] "GET /product.screen?product_id=HolyGouda&JSESSIONID=SD3SL1FF7ADFF8 HTTP 1.1" $status$ $timetaken$ "http://shop.buttercupgames.com/cart.do?action=view&itemId=HolyGouda" "$useragent$" $size$`,
	}
	var listmap []map[string]string
	for i := 0; i < 1000; i++ {
		listmap = append(listmap, copyevent(m))
	}
	for n := 0; n < b.N; n++ {
		for _, lm := range listmap {
			_ = copyevent(lm)
		}
	}
}

func copyevent(src map[string]string) (dst map[string]string) {
	dst = make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func BenchmarkCopyDeepCopySliceSlice(b *testing.B) {
	list := []string{"main", "access_combined", "/var/log/httpd/access_log", "$host$", `$ip$ $clientip$ - - [$ts$] "GET /product.screen?product_id=HolyGouda&JSESSIONID=SD3SL1FF7ADFF8 HTTP 1.1" $status$ $timetaken$ "http://shop.buttercupgames.com/cart.do?action=view&itemId=HolyGouda" "$useragent$" $size$`}
	listlist := make([][]string, 1000)
	for i := 0; i < 1000; i++ {
		listlist = append(listlist, list)
	}
	for n := 0; n < b.N; n++ {
		_ = deepcopy.Copy(listlist)
	}
}

func BenchmarkCopyDeepCopySliceMap(b *testing.B) {
	m := map[string]string{"index": "main",
		"sourcetype": "access_combined",
		"source":     "/var/log/httpd/access_log",
		"host":       "$host$",
		"_raw":       `$ip$ $clientip$ - - [$ts$] "GET /product.screen?product_id=HolyGouda&JSESSIONID=SD3SL1FF7ADFF8 HTTP 1.1" $status$ $timetaken$ "http://shop.buttercupgames.com/cart.do?action=view&itemId=HolyGouda" "$useragent$" $size$`,
	}
	var listmap []map[string]string
	for i := 0; i < 1000; i++ {
		listmap = append(listmap, copyevent(m))
	}
	for n := 0; n < b.N; n++ {
		_ = deepcopy.Copy(listmap)
	}
}
