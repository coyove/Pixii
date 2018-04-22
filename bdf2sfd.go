package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const pixel = 64
const D = 2
const descent = D * pixel
const ascent = 12*pixel - descent
const bitmask = 25

func main() {
	buf, _ := ioutil.ReadFile("Pixii-12.bdf")
	lines := strings.Split(string(buf), "\n")

	numglyph := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "CHARS ") {
			numglyph, _ = strconv.Atoi(line[6:])
			break
		}
	}

	var sfdheader = fmt.Sprintf(`SplineFontDB: 1.0
FontName: Pixii
FullName: Pixii
FamilyName: Pixii
Weight: Medium
Comments: Created by Coyove, inspired by unifont
Comments: Wen Quan Yi Bitmap Song
Comments: ZhongYi Electronic
Version: 1.00
ItalicAngle: 0
UnderlinePosition: -144
UnderlineWidth: 40
Ascent: %d
Descent: %d
NeedsXUIDChange: 1
XUID: [1021 140 1293607838 5610107]
FSType: 0
PfmFamily: 33
TTFWeight: 500
TTFWidth: 5
Panose: 2 0 6 4 0 0 0 0 0 0
LineGap: 721
VLineGap: 0
OS2WinAscent: 0
OS2WinAOffset: 1
OS2WinDescent: 0
OS2WinDOffset: 1
HheadAscent: 0
HheadAOffset: 1
HheadDescent: 0
HheadDOffset: 1
ScriptLang: 1
 1 latn 1 dflt 
Encoding: UnicodeBmp
UnicodeInterp: none
DisplaySize: -24
AntiAlias: 1
FitToEm: 1
WinInfo: 0 50 22
TeXData: 1 0 0 346030 173015 115343 0 1048576 115343 783286 444596 497025 792723 393216 433062 380633 303038 157286 324010 404750 52429 2506097 1059062 262144
BeginChars: 65536 %d
`, ascent, descent, numglyph)

	f, _ := os.Create("Pixii-outline.sfd")
	f.WriteString(sfdheader)

	i := 0

	type pair struct {
		unicode  int
		bitmap   [][]byte
		charname string
		width    int
	}

	table := make([]*pair, 0)
	for i < len(lines) {
		if strings.HasPrefix(lines[i], "STARTCHAR ") {
			uni, _ := strconv.Atoi(lines[i+1][9:])
			name := lines[i][10:]
			i = i + 2

			w, dw, offy := 0, 0, 0

			for lines[i] != "BITMAP" {
				if strings.HasPrefix(lines[i], "BBX") {
					bb := strings.Split(lines[i], " ")
					w, _ = strconv.Atoi(bb[1])
					dw, _ = strconv.Atoi(bb[3])

					h, _ := strconv.Atoi(bb[2])
					dh, _ := strconv.Atoi(bb[4])

					offy = 12 - D - dh - h
				}
				i = i + 1
			}

			i = i + 1
			bitmap := [][]byte{}
			for y := 0; y < offy; y++ {
				bitmap = append(bitmap, nil)
			}

			for lines[i] != "ENDCHAR" {

				bits := make([]byte, 0, 12)
				for x := 0; x < dw; x++ {
					bits = append(bits, 0)
				}

				for ii := range lines[i] {
					b, _ := strconv.ParseInt(lines[i][ii:ii+1], 16, 64)
					for bi := uint64(60); bi < 64; bi++ {
						bits = append(bits, byte(uint64(b)<<bi>>63))
					}
				}

				bits = bits[:dw+w]
				bitmap = append(bitmap, bits)
				i = i + 1
			}

			p := &pair{
				unicode:  uni,
				bitmap:   bitmap,
				charname: name}

			if w+dw <= 6 {
				if uni > 0x3000 {
					p.width = 12
				} else {
					p.width = 6
				}
			} else {
				p.width = 12
			}
			table = append(table, p)
		} else {
			i++
		}
	}

	for i, p := range table {
		set := ""
		bitmap := p.bitmap

		for i := 0; i < 12; i++ {
			if i >= len(bitmap) {
				break
			}
			line := bitmap[i]
			if line == nil {
				continue
			}

			for ix, b := range line {
				if b == 0 {
					continue
				}

				x0 := ix * pixel
				x1 := x0 + pixel
				y := 12 - D - i
				y0 := y * pixel
				y1 := y0 - pixel
				set += fmt.Sprintf(`%d %d m %d
 %d %d l %d
 %d %d l %d
 %d %d l %d
 %d %d l %d
`, x0, y0, bitmask,
					x0, y1, bitmask,
					x1, y1, bitmask,
					x1, y0, bitmask,
					x0, y0, bitmask)
			}
		}

		f.WriteString(fmt.Sprintf(`StartChar: %s
Encoding: %d %d %d
Width: %d
Flags: HW
TeX: 0 0 0 0
Fore
SplineSet
%s
EndSplineSet
EndChar
`, p.charname, p.unicode, p.unicode, i, p.width*pixel, set))

	}

	f.WriteString(`EndChars
EndSplineFont`)

	f.Close()
}
