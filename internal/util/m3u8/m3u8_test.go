package m3u8_test

import (
	"fmt"
	"log"
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/transfer"
	"video-downloader-go/internal/util/m3u8"
)

func TestTsTransferInit(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	config.Load("/Users/ambitious/Desktop/学习/Go/projects/video-downloader-go/config/config.yml")
	fmt.Println(transfer.Instance(""))
}

func TestReadTsUrls(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	headers := map[string]string{}
	url := "https://apd-vlive.apdcdn.tc.qq.com/defaultts.tc.qq.com/B_tRCdt2L6hl1ezG-aht1_p264FX2g4lSJ8vpBLy4ShDviX-0x9w95_rx7NVLaVkg3/svp_50112/fHIXvYesr8QrPXsCjJ1lSnBscoDoNQMbDWSOSjKfwfkHSXo2ErfZlPoGcDRkHOnLj3Tqz98eseYnD-CVfNZQChihBULS2NAOPTdrKgCLkNV68aaPAm62SN2_rdqSMHz4VuPJxBtWV20Suri1hZa1dNb2RD0kfkPrG3wtBkVjG_LiaWliiU9WCtSJQ-1kdkacVGLHCnXyNkI5lgiPfNRAHMqSvkI19YhEoTG4zkdFOxahbEqflwZPRA/gzc_1000102_0b53zuafgaaax4apbskt7js4btodkpaaav2a.f322016.ts.m3u8?ver=4"
	metas, err := m3u8.ReadTsUrls(url, headers)
	if err != nil {
		t.Fatal(err)
	}
	for _, meta := range metas {
		fmt.Println(*meta)
	}
}

func TestCheckM3U8(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	headers := map[string]string{
		"Referer": "https://mgtv.com",
	}
	url := "https://pcvideotx.titan.mgtv.com/c1/2023/11/10_0/821A43EA5936B8B17C54A19EA08C65B8_20231110_1_1_1203_mp4/8B2A2F03FE9B4FC510357EE622C46265.m3u8?arange=0&pm=b5WNERoDc2s0USn1a8oQJjGaDgRUqOpjYFOfh3cdQMceZfcBWTg5HOO_aUNlnn_o4qVVO69NAUDuiT2JJC54g11jhTK_YmquHviZy4A0n~dGksf_5fkjH4jENQtLgwKnApeTzOJ4YMXbkEq3U68zWNRb7PtjBUgYZNN3QJa2BJC8MNSOe~E1PpmMpj_b7rElXZ1uDWJ8QW~AAdWWIhnhjqvOBqoBeQKZvMSYqgd3YCzb56L1Y5FCnDfHIK1lTRdzq2~Sj0MZnsgbVZrfVi5y4Tjm5ajq_PgABeM7Uda9SlcdGtR8QI_5IK4G7SRhKx5uK68ni4Qq4RN8yJB5gF7ksHwqh8EbaCTnC3jddciVnuICvUQ8NatijLApiFcFyNOMl5yDPK8L0L5CHyYU935P3WFzeQFXaqgTQlZY0MvdSch12ESiN23vfao9_vyG0Gnjny0_jdNMT2AfvkGY&mr=JaoWKkQ2JcrSoGAt3TfzT0I13LZq_7yVzeQ5nwFtcQHzQPEGgACyOBdJUs9qBAsQazptYEsaUZmWRmh_fP02MixeTH8jjzJ1RFHOkgW3bEs7T2_YTVPpAJk9KsRCR71kj14u779MB7~tzF~YZK7JjCtvUcvs6kKHWwNbYuziQmjYSXyR6Sl1L~T00zYS29KzVyOMqfWVMpoYX~WThZ3jlitb9rzWegG1VP9IMKr1sCyfN42fgNwlNI8on9IpSF50CuEa17mRpJHBCBzR0Ne6pRb8M8H213VxwejmYtlHUHuZgH8kXrQwWimAp99MGgqIO7hfa0VW8S~~87ZByrCBxIJuwOLhsDxZcHVfDK0rUD10QsbXcxfadjzlLF9kNDX5NwCmoRjVFRgSAjX6X9pTcoudAQZ3H3aGcpbfaKjVYbtLUbFIs18OPu~qRQhPjjnljqkNiunNXF1IDoj3jor7xTbyQguWPWUxmfe52MBvDnEww5eB8C9jUQqGhj1_gk5hwSVd9NxUX3PX_EfGBfXzfkjslgQ_MfzHI7rRbAHMCaa8EsCYLG96SBrXa_7w5xowZYTwfANBzXYbS7Byzyd41Wlx5mNb8qUT_08AEQJgYxYczHHgRO9f1_S~RYuqz9cLyXG9Avs~1oliPE8yMmT1Og--&uid=null&scid=25015&cpno=6i06rp&ruid=573e14c3de6143d2&sh=1"
	res := m3u8.CheckM3U8(url, headers)
	if !res {
		t.Fail()
	}
}

// 测试解析 EXT-X-MAP 头
func TestResolveXMap(t *testing.T) {
	line := `#EXT-X-MAP:URI="F0D8437719A36EBD4048C8534EC67035_0_0_0_fmp4.mp4?arange=0&pm=8a5FBb0qCpf~NUGyCxwPC7gKlAjZ3ToeO0hmkUwQPOV_qKadzWLybjh9w5lQaGE86GfEnC5Pqv70cacMEukhUnyDA0O8zWtCueUbMfZryI8GHrK2oQZyHgpo1JlWY7zfbjVfzgAqrpdEzkC4DP_KeQiGLSSVu8wejY462GU0QJ3tQRI6dj_~S8OIO2drR~e5yxksqgntxfPmwGMct0MIaTDTcefFUiFMNM2wYn~pz4eiDAlpr~vqB2BoSreoVHw0VBzNZxsBFmCtocm1o5m0iqfzLipiiokhANUVgFWb28waNNPWYwbKEWr4F6wnLyAuLh7TVm85qwGIcEeRz5IWjKsNZxtq86T9TZzybvJ18QXqH3IgagEcVKN3hXX_lWLbUpyojgEEn1_ipWLokFMtfVfjmGG_qVkN&mr=V0AY~7bxWVVgbvsSQ100takUSjINTtL7BVb7FZ3LXJSldCvVEeeN3VEl9CXIJxdM1P3Nikv0tsEt8HLnd8xYJ6M6pULaMkIpZH5ZPGO3SCTkmHIQqyVTghAf0vV8uufq5yVyw2deo6JTJh7WFptP73jqwt5VOE1Sgcjfc0BD5YUiDU91QhfTuZ5SOtpEFaFKyW_KbZrnKrDWJRT~Fp8LuMNqMMBy14JO6nLWP~s7NbXPZsXfVwe40AnvQ~BxLUrMsYAnqdVTtCi6USlu4Q4QYzNmehsJhlv~WFbArM4poEzjFk0f8qSkran54vENtlX3QUv3mpC9MqjrIUNeREpNfldNHkagACFM9U3Pq4FD7ar0Q1jMxuEYgTZwnsCxaysM_pDQ9tVGcv~1tPHR0jKgJMGcH2TKf5H0XF_dxj_bBzPpmpmPKQn0wUbWkF8qyqfmBr5aas3_967UoFdiRAxZP1HlAkD5hseRMVfT0U95Hn4PMtjsBB4D9ofd4vRUjabysB6975pP3SXtIge9dpXIHqh1VDJB7rQLJuSrvAvLY9jEHVWqbaijeGOYSRldOjR9tS5rSSoFTATbOg9~4fAftiH8~Jxsi8lP44xsHhHLL4m4~FmhGR0LmlTxNxhdyH572KJ1Ro~FUEmjcuDh7jhBKQ--&uid=null&scid=25021&cpno=6i06rp&ruid=e01d0afeb25942bc&sh=1&ftc=webO1&sftc=v6.7.46ds1_vtpVOD"`
	hi, err := m3u8.ResolveXMap(line)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println(hi)
}
