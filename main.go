package main

import (
	"bufio"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/google/gousb"
	"github.com/shtirlic/postrint/printer"
	"golang.org/x/image/font"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	fontFile := "B612Mono-Regular.ttf"

	fontBytes, err := ioutil.ReadFile(fontFile)
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
	fontSize := 5.0
	fontSpacing := 1.5

	rgba := image.NewGray(image.Rect(0, 0, 390, 200))
	draw.Draw(rgba, rgba.Bounds(), bg, image.Point{}, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(203)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	text := []string{"A/C ID   DATE   GMT   FLTN     CITY PAIR",
		".VQ-BBM  12FEB  1155  SDM890   UTFF ULLI"}
	//text = []string{"1811-7625-6209"}

	pt := freetype.Pt(10, 10+int(c.PointToFixed(fontSize)>>6))
	for _, s := range text {
		_, err = c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += c.PointToFixed(fontSize * fontSpacing)
		//fmt.Println((pt.Y.Ceil()))
	}

	outFile, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	cutImg := rgba.SubImage(image.Rect(0, 0, 384, int(pt.Y.Ceil())))
	defer outFile.Close()
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

	prt, ctx, err := printer.NewUsbPrinter(gousb.NewContext(), 0x28e9, 0x0289, 0x01)
	defer ctx.Close()

	if err != nil {
		os.Exit(1)
	}
	prt.PrintImage(cutImg)
	prt.FeedLines(2)
}
