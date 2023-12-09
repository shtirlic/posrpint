package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/google/gousb"
	"golang.org/x/image/font"

	"github.com/shtirlic/posprint/printer"
)

type AcarsMessage struct {
	OpenDate  string
	PrintDate string
	Reg       string
	Message   string
}

func wordWrap(text string, lineWidth int) (wrapped string) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}
	wrapped = words[0]
	spaceLeft := lineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}
	return
}

func main() {
	fontFile := "B612Mono-Regular.ttf"

	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	fg, bg := image.Black, image.White
	fontSize := 5.2
	fontSpacing := 1.4

	rgba := image.NewGray(image.Rect(0, 0, 386, 2000))
	draw.Draw(rgba, rgba.Bounds(), bg, image.Point{}, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(203)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	// text := []string{"A/C ID   DATE   GMT   FLTN     CITY PAIR",
	// ".VQ-BBM  12FEB  1155  SDM890   UTFF ULLI"}
	// //text = []string{"1811-7625-6209"}

	aM := AcarsMessage{
		OpenDate:  "16/03/22   09:00:55   OPEN",
		PrintDate: "16/03/22 09:01:15",
		Reg:       "D-ABNW",
		Message:   "EBBR DEP ATIS S 0850Z   ULLI 272030Z 00000MPS 4500 0600NE PRFG BR SCT025 06/05 Q1031 R10R/090060 TEMPO 0200 FG VV002",
	}
	dat, err := os.ReadFile("acars.tpl")
	t, err := template.New("").Parse(string(dat))
	var tpl bytes.Buffer
	err = t.Execute(&tpl, aM)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(tpl.String())

	var text []string
	var str []string
	var strTpl string
	// str:= word_wrap(tpl.String(),40)
	for _, s := range strings.Split(tpl.String(), "\n") {
		str = append(str, wordWrap(s, 40))
	}
	strTpl = strings.Join(str, "\n")

	text = append(text, strings.Split(strTpl, "\n")...)
	// sc := bufio.NewScanner(os.Stdin)
	// for sc.Scan() {
	//	str:= word_wrap(sc.Text(),40)
	//	strArray:=strings.Split(str,"\n")
	//	text = append(text, strArray...)
	// }
	fmt.Println(text)
	pt := freetype.Pt(3, 0+int(c.PointToFixed(fontSize)>>6))
	for _, s := range text {
		_, err = c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += c.PointToFixed(fontSize * fontSpacing)
		fmt.Println(pt.Y.Ceil())
	}

	outFile, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	cutImg := rgba.SubImage(image.Rect(0, 0, 386, pt.Y.Ceil()))
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(outFile)
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, cutImg)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")

	prt, ctx, err := printer.NewUsbPrinter(gousb.NewContext(), 0x28e9, 0x0289, 0x81)
	defer func(ctx *gousb.Context) {
		err := ctx.Close()
		if err != nil {
			log.Println(err)
		}
	}(ctx)
	if err != nil {
		os.Exit(1)
	}
	prt.PrintImage(cutImg)
	prt.FeedLines(3)
}
