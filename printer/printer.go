package printer

import (
	"bytes"
	"fmt"
	"github.com/google/gousb"
	"image"
	"image/color"
	"log"
)

const (
	ESC byte = 0x1b
	GS  byte = 0x1d
	NUL byte = 0x00
	SOH byte = 0x01
	ETX byte = 0x03
	EOT byte = 0x04
	LF  byte = 0x0A
	HT  byte = 0x09
)

var (
	CmdInit         = []byte{ESC, '@'}
	CmdFeed         = []byte{ESC, 'd'}
	CmdCut          = []byte{GS, 'V', 'A'}
	CmdImage        = []byte{GS, 'v', '0'}
	CmdImageBitMode = []byte{ESC, 0x2A, 33}
)

type Printer struct {
	*gousb.Device
	vid            int
	pid            int
	outEndpointNum int
	//*gousb.Interface
	//*gousb.OutEndpoint
}

func NewUsbPrinter(ctx *gousb.Context, vid int, pid int, outEndpointNum int) (prt *Printer, ctxOut *gousb.Context, err error) {

	dev, err := ctx.OpenDeviceWithVIDPID(gousb.ID(vid), gousb.ID(pid))
	if err != nil {
		log.Printf("Could not open a device: %v", err)
		return
	}

	prt, ctxOut, err = &Printer{
		Device:         dev,
		vid:            vid,
		pid:            pid,
		outEndpointNum: outEndpointNum,
	}, ctx, nil
	return
}

func (p *Printer) Write(data []byte) (*Printer, error) {
	intf, done, err := p.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", p, err)
	}
	defer done()

	ep, err := intf.OutEndpoint(0x01)
	if err != nil {
		log.Printf("%s.OutEndpoint(0x01): %v", intf, err)
		return p, err
	}

	numBytes, err := ep.Write(data)

	if numBytes != len(data) {
		log.Printf("%s.Write([%d]): only %d bytes written, returned error is %v", ep, len(data), numBytes, err)
		return p, err
	}
	fmt.Printf("%d bytes of data: %v successfully sent.\r\n", len(data), data)

	return p, err
}

func (p *Printer) FeedLines(count int) {
	buff := bytes.NewBuffer(CmdFeed)
	buff.WriteByte(0x1)
	for count > 0 {
		p.Write(buff.Bytes())
		//time.Sleep(time.Second*1)
		count -= 1
	}
}

func (p *Printer) CutPaper() {
	p.Write(CmdCut)
}

func (p *Printer) PrintImage(imgSource image.Image) {
	img := invertImage(imgSource)
	sz := img.Bounds().Size()

	maxWidth := 384
	threshold := 16
	// lines are packed in bits
	imageWidth := sz.X
	if imageWidth > maxWidth {
		// truncate if image is too large
		imageWidth = maxWidth
	}

	bytesWidth := imageWidth / 8
	if imageWidth%8 != 0 {
		bytesWidth += 1
	}

	data := make([]byte, bytesWidth*sz.Y)

	for y := 0; y < sz.Y; y++ {
		for x := 0; x < imageWidth; x++ {
			if int(img.GrayAt(x, y).Y) >= threshold {
				// position in data is: line_start + x / 8
				// line_start is y * bytesWidth
				// then 8 bits per byte
				data[y*bytesWidth+x/8] |= 0x80 >> uint(x%8)
			}
		}
	}
	fmt.Println(data)
	p.Raster(imageWidth, sz.Y, bytesWidth, data)
}

func intLowHigh(inpNumber int, outBytes int) (outp []byte) {

	maxInput := (256 << (uint((outBytes * 8)) - 1))

	if outBytes < 1 || outBytes > 4 {
		log.Println("Can only output 1-4 bytes")
	}
	if inpNumber < 0 || inpNumber > maxInput {
		log.Println("Number too large. Can only output up to " + string(maxInput) + " in" + string(outBytes) + "byes")
	}
	for i := 0; i < outBytes; i++ {
		inpNumberByte := byte(inpNumber % 256)
		outp = append(outp, inpNumberByte)
		inpNumber = inpNumber / 256
	}
	return
}

func (p *Printer) Raster(width, height, lineWidth int, imgBw []byte) {
	densityByte := byte(0)
	header := []byte{
		0x1D, 0x76, 0x30}
	header = append(header, densityByte)
	width = (width + 7) >> 3
	header = append(header, intLowHigh(width, 2)...)
	header = append(header, intLowHigh(height, 2)...)

	fullImage := append(header, imgBw...)

	p.Write(fullImage)
}

func invertImage(img image.Image) (inverted *image.Gray) {
	bounds := img.Bounds()
	inverted = image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, _ := img.At(x, y).RGBA()
			r = 255 - r>>8
			inverted.SetGray(x, y, color.Gray{uint8(r)})
		}
	}
	return
}
