// library to parse pdfdoc
// author: prr. azul software
// based on azulpdfold/pdfLib/pdfLib.go
//
// library pdf files in go
// author: prr
// date: 21 June 2023
// copyright 2023 prr azul software
//

package azulParseLib

import (
	"os"
	"fmt"
	"log"
	"bytes"
	"strconv"
//	"strings"
//	"io"
//	"compress/zlib"

//    util "github.com/prr123/utility/utilLib"
)

const (
	letter =iota
	landscape
	A5
	A4
	A3
	A2
	A1
)

const (
	Catalog = iota
	Pages
	Page
	Contents
	Font
	FontDescriptor
	Data
)

type ParsePdf struct {
	PdfFilnam string `yaml:"pdf file"`
	Majver int `yaml:"major"`
	Minver int `yaml:"minor"`
	fil *os.File
	buf *[]byte
//	filSize int64
//	filNam string
//	objSize int
	NumTrailer int `yaml:"numTrailer"`
	NumObj int `yaml:"numObj"`
	PageCount int `yaml:"pages:"`
	InfoId int
	RootId int
	PagesId int
	xrefPos int
	StartXrefPos int `yaml:"startxref"`
//	trailerPos int
//	xrefPos2 int
//	startXrefPos2 int
//	trailerPos2 int
//	objStart int
	PageObjList *[]int
	FontObjList *[]int
//	fCount int
//	gStateIds *[]int
//	gCount int
//	xObjIds *[]int
//	xObjCount int
//	xrefList *[]xrefObj
//	pageList *[]pgObj
//	objList *[]pdfObj
	ObjList *[]pdfObj
//	fontList *[]fontObj
//	gStateList *[]gStateObj
//	fonts *[]objRef
//	gStates *[]objRef
//	xObjs *[]objRef
	mediabox *[4]float32
//	test bool
//	verb bool
	Dbg bool `yaml:"debug"`
}

type pdfObj struct {
	BufPos int
	EndPos int
	objId int
	objTyp int
	objTypStr string
	dictMap *objDict
	dict bool
	array bool
	simple bool
	parent int
	stream bool
	streamSt int
	streamEnd int
}

type dictVal struct {
	valTyp int
	valStr string
}

type objDict map[string]*dictVal

type xrefObj struct {
	bufPos int
	startobj int
	numObjs int
}

type pgObj struct {
	id int
	pageNum int
	mediabox *[4]float32
	contentId int
	parentId int
	fontId int
	gStateId int
	xObjId int
	fonts *[]objRef
	gStates *[]objRef
	xObjs *[]objRef
}

type objRef struct {
	Id int
	Nam string
}

type fontObj struct {
	id int
	fontDesc *FontDesc
	fontDescId int
	subtyp string
	name string
	base string
	encode string
	desc int
	fchar int
	lchar int
	widths int
	widthList *[]int
}

type FontDesc struct {
	id int
	fname string
	flags int
	italic int
	ascent int
	descent int
	capheight int
	avgwidth	int
	maxwidth	int
	fontweight	int
	Xheight int
	fontBox [4]int
	fileId int
}

type gStateObj struct {
	BM string

}

type resourceList struct {
	fonts *[]objRef
	gStates *[]objRef
	xObjs *[]objRef
}

type pdfDoc struct {

	pageType int

}

func InitPdfParseLib(pdfFilnam string)(info *ParsePdf, err error) {
	pdf := ParsePdf{
		PdfFilnam: pdfFilnam,
		Dbg: true,
	}
    pdfFil, err := os.Open(pdfFilnam)
    if err != nil {
        return nil, fmt.Errorf("os.Open %v\n", err)
    }

    filinfo, err := pdfFil.Stat()
    if err != nil {
        return nil, fmt.Errorf("parseFil.Stat: %v\n", err)
    }

	buf := make([]byte, filinfo.Size())

	_, err = pdfFil.Read(buf)
	if err !=nil {return nil, fmt.Errorf("pdfFil.REad: %v\n", err)}

	pdf.buf = &buf

//    pdf.fil = parseFil
//    pdf.filSize = filinfo.Size()
//    pdf.filNam = pdfFilnam

	// maybe allocote more
//	fontIds := make([]int, 10)
//	pdf.fontIds = &fontIds

//	gStateIds := make([]int, 10)
//	pdf.gStateIds = &gStateIds

	return &pdf, nil
}

func (pdf *ParsePdf) SavePdf (filnam string)(err error) {

	return nil
}

func (pdf *ParsePdf) parseTopTwoLines()(err error) {

	buf := *pdf.buf

	maxPos := 100
	if len(buf) < maxPos {maxPos = len(buf)}

	//read top line
	endFl := -1
	for i:=0; i<maxPos; i++ {
		if (buf[i] == '\n') || (buf[i] == '\r'){
			endFl = i
			break
		}
	}

	if endFl == -1 {return fmt.Errorf("no eol in first line!")}

	idx := bytes.Index(buf[:5],[]byte("%PDF-"))
	if idx == -1 {return fmt.Errorf("first line %s string is not \"%%PDF-\"!", string(buf[:5]))}

	verStr := string(buf[5:endFl])

	majver:= 99
	minver:= 99
	_, err = fmt.Sscanf(verStr, "%d.%d", &majver, &minver)
	if err != nil {return fmt.Errorf("cannot parse pdf version: %v!",err)}

	if pdf.Dbg {fmt.Printf("dbg: pdf version: %d.%d\n", majver, minver)}

	if majver > 2 {return fmt.Errorf("invalid pdf version %d", majver)}

	pdf.Majver = majver
	pdf.Minver = minver

	// second line
	startSl := -1
	for i:=endFl+1; i<maxPos; i++ {
		if buf[i] == '%' { startSl = i}
	}
	if startSl == -1 {return fmt.Errorf("no % in second line!")}

	endSl := -1
	for i:=startSl; i<maxPos; i++ {
		if buf[i] == '\n'{
			endSl = i
			break
		}
	}

	if endSl == -1 {return fmt.Errorf("no eol in second line!")}

	endPSl := endSl
	if buf[endSl-1] == '\r' {endPSl--}

	dif := endPSl - startSl

	if dif > 5 && pdf.Dbg {
		fmt.Printf("dif > 5: %d\n", dif)
		for i:=startSl; i< endPSl; i++ {
			fmt.Printf("[%d]:%d/%q ", i, buf[i], buf[i])
		}
		fmt.Printf("\n2 line [%d:%d]: %s\n", startSl, endSl, string(buf[startSl:endSl]))
	}

	if dif <5 || dif> 10 {return fmt.Errorf(" dif: %d no 4 chars after percent chars %d:%d!", dif, startSl, endSl)}

	for i:=startSl+1; i<startSl+5; i++ {
		if !(buf[i] > 120) {return fmt.Errorf("char %q not valid in second top line!", buf[i])}
	}

	return nil
}

func (pdf *ParsePdf) parseLast3Lines()(xrefVal int, err error) {

	buf := *pdf.buf

	llEnd := len(buf) -1
	xrefVal = -1

	// now we have the real last line which should contain %%EOF
	slEnd := -1

	// fix if buf[llEnd] == '\n\'
	if (buf[llEnd] == '\n') {llEnd--}
	if (buf[llEnd] == '\r') {llEnd--}

	idx := bytes.Index(buf[llEnd-10:llEnd+1], []byte("%%EOF"))
	if idx == -1 {return -1, fmt.Errorf("last line %s: cannot find \"%%EOF\"!", string(buf[slEnd+1:llEnd+1]))}
	llStart := llEnd-10+idx

	idx = bytes.Index(buf[llStart - 50:llStart], []byte("startxref"))
	if idx == -1 {return -1, fmt.Errorf("last line: cannot find \"startxref\"!")}

	pdf.StartXrefPos = llStart-50 + idx

//fmt.Printf("startxref: %s\n", string(buf[pdf.startxref:tlEnd]))
	tlinEnd := -1
	tlinSt := pdf.StartXrefPos + len("startxref")
	for i:=tlinSt; i<tlinSt + 20; i++ {
		if buf[i]=='\n' {
			tlinEnd = i
			break
		}
	}
	if tlinEnd <0 {return -1, fmt.Errorf("no eol found in startxref!")}

	slinEnd := -1
	for i:=tlinEnd+1; i<tlinEnd+ 20; i++ {
		if buf[i]=='\n' {
			slinEnd = i
			break
		}
	}
	if slinEnd <0 {return -1, fmt.Errorf("no eol found in line after startxref!")}
	if buf[slinEnd-1] == '\r' {slinEnd--}

	_, err = fmt.Sscanf(string(buf[tlinEnd+1:slinEnd+1]),"%d", &xrefVal)
	if err != nil {return -1, fmt.Errorf("cannot parse xref: %v!", err)}
//	pdf.xrefPos = xrefVal
	return xrefVal, nil
}

func (pdf *ParsePdf) findNextObj(start int)(objSt, objEnd int, err error) {

	buf := *pdf.buf
	objSt = -1

	idx := bytes.Index(buf[start: start+200], []byte("obj"))
	if idx <0 {return -1, -1, fmt.Errorf("no obj found!")}

	objtSt := start + idx
//fmt.Printf("objtSt[%d]: %s\n", objtSt, string(buf[objtSt: objtSt+3]))

	for i:= objtSt; i>0; i-- {
		if buf[i] == '\n' {
			objSt = i+1
			break
		}
	}
	if objSt == -1 {return -1, -1, fmt.Errorf("no eol before obj found!")}
//fmt.Printf("objSt[%d]: %s\n", objSt, string(buf[objSt: objtSt+3]))


	idx = bytes.Index(buf[objtSt+3: objtSt+200], []byte("endobj"))
	if idx <0 {return -1, -1, fmt.Errorf("no endobj found!")}

	objEnd = objtSt + 3 + idx +6
//fmt.Printf("obSt[%d]: %s\n", objSt, string(buf[objSt: objEnd]))

	return objSt, objEnd, nil
}

// method parses trailer
func (pdf *ParsePdf) parseTrailer(stPos int)(prev int, err error) {

	prev = -1
	if pdf.StartXrefPos < 1 {return -1, fmt.Errorf("no valid startxref")}

	// find trailer keyword in first line
	txtslic, nextPos, err := pdf.readLine(stPos)
	if err != nil {return -1, fmt.Errorf("trailer first line no eol")}

	tidx:= bytes.Index(txtslic, []byte("trailer"))
	if tidx == -1 {return -1, fmt.Errorf("key word trailer not found!")}

	// get trailer dict  line

	txtslic, _, err = pdf.readLine(nextPos)

	dictCont, err := pdf.parseDict(txtslic)
	if err != nil {return -1, fmt.Errorf("parseDict trailer: %v", err)}

	fmt.Printf("dbg -- trailer: %s\n", string(dictCont))

	trailerDict, err := pdf.parseDictCont(dictCont)
	if err != nil {return -1, fmt.Errorf("parseDictCont: %v", err)}

	fmt.Println("trailerDict:")
	for k,v := range trailerDict {
		fmt.Printf("key: %s val: %s\n", k, v.valStr)
	}
	xrefObj, ok:= trailerDict["Prev"]
	if ok {
//		fmt.Printf("xrefObj: %v\n", xrefObj)
		pos, err := strconv.Atoi(xrefObj.valStr)
		if err != nil {return -1, fmt.Errorf("Atoi: Key Prev has no int value", err)}
		prev = pos
	}

	return prev, nil
}


// function finds start and end of double brackets
func (pdf *ParsePdf) parseDblBracket(pSt int, pLen int)(bst int, bend int, err error) {

	buf := *pdf.buf
	bst = -1
	nestLev :=0

//	fmt.Printf("dbg start dblBr -- %d %d %s\n", pSt, pSt + pLen, string(buf[pSt: pSt+pLen+1]))
	for i:=pSt; i< pSt+pLen; i++ {
		if buf[i] == '<' {
			if buf[i+1] == '<' {
				if nestLev == 0 {
					bst = i+2
				}
				nestLev++
			}
		}
		if buf[i] == '>' {
			if buf[i+1] == '>' {
				if nestLev==1 {
					bend = i-1
				}
				nestLev--
			}
		}
	}

//	fmt.Printf("dbg -- %d %d %s\n", bst, bend, string(buf[bst:bend+1]))
	if bst == -1 {return -1, -1, fmt.Errorf("no << found!")}
	if bend == -1 {return -1, -1, fmt.Errorf("no >> found!")}
	if nestLev != 0{return bst, -1, fmt.Errorf("nestLev: %d!", nestLev)}
	return bst, bend, nil
}

// method that parses dict
func (pdf *ParsePdf) parseDict(dictBuf []byte)(out []byte, err error) {

//	fmt.Printf("dbg start dblBr -- %d %d %s\n", pSt, pSt + pLen, string(dictBuf[pSt: pSt+pLen+1]))
	bst := -1
	for i:=0; i< len(dictBuf); i++ {
		if dictBuf[i] == '<' {
			if dictBuf[i+1] == '<' {
				bst = i+2
				break
			}
		}
	}
	if bst == -1 {return out, fmt.Errorf("no << found!")}

	bend:= -1
	for i:=len(dictBuf)-1; i>0; i-- {
		if dictBuf[i] == '>' {
			if dictBuf[i-1] == '>' {
				bend = i-1
				break
			}
		}
	}

	if bend == -1 {return out, fmt.Errorf("no >> found!")}
//	fmt.Printf("dbg -- %d %d %s\n", bst, bend, string(dictBuf[bst:bend+1]))
	out = dictBuf[bst:bend]
	return out, nil
}


func (pdf *ParsePdf) parseTrailerDict(key string, dict []byte, Type int)(objId int, val int, err error) {

	//find key
	keyByt := []byte("/" + key)


//fmt.Printf("trailer start: %d end: %d %s\n", Start, pdf.startxref, string(buf[Start:pdf.startxref]))

//fmt.Printf("key: %s\nstr [%d:%d]: %s\n", string(keyByt), Start, linEnd, string(buf[Start:(Start + 40)]))

	keyIdx := bytes.Index(dict, keyByt)

	if keyIdx == -1 {return -1, -1, fmt.Errorf("cannot find keyword %s", key)}

	valSt:= keyIdx+len(keyByt)
	rootEnd := -1
	for i:=valSt; i< len(dict); i++ {
		switch dict[i] {
		case '/','\n','\r','>':
			rootEnd = i
		default:
		}
		if rootEnd > 0 {break}
	}
	if rootEnd == -1 {rootEnd = len(dict)}
//{return -1, -1, fmt.Errorf("cannot find end delimiter after key %s", key)}

	keyObjStr := string(dict[valSt:rootEnd])

fmt.Printf("dbg -- parse Trailer: Key %s: value: %s\n", key, keyObjStr)
	val = -1

	switch Type {
	case 1:
		_, err = fmt.Sscanf(keyObjStr, " %d %d R", &objId, &val)
		if err != nil {return -1,-1, fmt.Errorf("cannot parse obj ref after keyword %s! %v", key, err)}
	case 2:
		_, err = fmt.Sscanf(keyObjStr, " %d", &objId)
		if err != nil {return -1, -1, fmt.Errorf("cannot parse obj val after keyword %s! %v", key, err)}
	default:
		return -1, -1, fmt.Errorf("invalid obj value type!")
	}
	return objId, val, nil
}


/*
func (pdf *InfoPdf) getKVStr(instr string)(outstr string, err error) {

	stPos := -1
	endPos := 0

	for i:=0; i< len(instr); i++ {
		if instr[i] == '<' {
			if instr[i+1] == '<' {
				stPos = i+2
				break
			}
		}
	}

	if stPos == -1 {return "", fmt.Errorf("no open double bracket!")}

	for i:=len(instr)-1; i> stPos; i-- {
		if instr[i] == '>' {
			if instr[i-1] == '>' {
				endPos = i-1
				break
			}
		}
	}

	if endPos == 0 {return "", fmt.Errorf("no closing brackets!")}

	outstr = instr[stPos: endPos] + "\n"
	return outstr, nil
}

func (pdf *InfoPdf) getKvMap(instr string)(kvMap map[string]string , err error) {

//fmt.Printf("******* start getKvMap\n")
//fmt.Println(instr)
//fmt.Printf("******* end getKvMap\n")
	ist := 0
	icount := 0
	linStr := ""
	key := ""
	val := ""
	valStr := ""
	kvMap = make(map[string]string)

	for i:=0; i< len(instr); i++ {
		if instr[i] != '\n' { continue}

		linStr = instr[ist:i]
		icount++
//fmt.Printf("linStr %d: %s\n", icount, linStr)
			_, err = fmt.Sscanf(linStr, "/%s %s", &key, &val)
			if err != nil {return kvMap, fmt.Errorf("parse error in line %d %s %v", icount, linStr, err)}
			// 2 : first letter is / second is ws
//fmt.Printf("key: %s val: %s %q\n", key, val, val[0])

			switch val[0] {
			case '/':
				valStr = linStr[(len(key)+2):]
				ist = i+1

//fmt.Printf("    /valStr: %s\n", valStr)

			case '<':
				remSt := ist + len(key) +1
				remStr := string(instr[remSt:])
//fmt.Printf("remStr: %s %d\n", remStr, remSt)
				tvalStr, errStr := pdf.getKVStr(remStr)
				if errStr != nil {return kvMap, fmt.Errorf("parse error in line %s %v", remStr, errStr)}
//fmt.Printf("   <valStr: %s\n", valStr)
				ist = remSt + len(tvalStr) + 5
				i = ist + 6
				tvalByt := []byte(tvalStr)
				for j:=0; j< len(tvalByt); j++ {
					if tvalByt[j] == '\n' {tvalByt[j] = ' '}
				}
				valStr = string(tvalByt)
			default:
				valStr = linStr[(len(key)+2):]
				ist = i+1
//fmt.Printf("    def valStr: %q %s\n", val[0], valStr)

			}


			kvMap[key] = valStr

	} // i



	if ist == 0 {return kvMap, fmt.Errorf("no eol found!")}

	return kvMap, nil
}
*/

/*
func (pdf *ParsePdf) decodeStream(objId int)(streamSl *[]byte, err error) {

//fmt.Printf("******* getstream instr\n")
//fmt.Println(instr)
//fmt.Printf("******* end ***\n")

	buf := *pdf.buf
	obj := (*pdf.ObjList)[objId]

	if obj.streamSt < 0 {return nil, fmt.Errorf("no stream start found!")}

	streamLen := obj.streamEnd -obj.streamSt
	if streamLen < 1 {return nil, fmt.Errorf("no stream found!")}

	if pdf.Dbg {fmt.Printf("**** stream [length: %d] ****\n", streamLen)}

	stbuf := buf[obj.streamSt:obj.streamEnd]

	bytStream := bytes.NewReader(stbuf)

	streamBuf := new(bytes.Buffer)

	bytR, err := zlib.NewReader(bytStream)
	if err != nil {return nil, fmt.Errorf("stream deflate error: %v", err)}

//	_ = copy(stream, bytR)
	io.Copy(streamBuf, bytR)

	bytR.Close()

	stream := streamBuf.Bytes()
	return &stream, nil
}
*/
// method parse the xref section
// line 1 xref
// line 2 [first Obj] [numObj]
// lines 3 to 3 + numObj: objpos
func (pdf *ParsePdf) parseXref(stPos int)(trailStartPos int, err error) {

//	buf := *pdf.buf

	objStart := 0
	objNum := 0

	if pdf.StartXrefPos <1 { return -1, fmt.Errorf("not a valid xref supplied!")}
//	if pdf.trailerPos <1 { return fmt.Errorf("not a valid trailer supplied!")}

	// line 1: xref
	// line 2: get first Obj and number f objects
	// line 3: get first Obj

	// line 1: xref
	txtslic, nextPos, err := pdf.readLine(stPos)
	if err != nil {return -1, fmt.Errorf("readLine for xref")}

	idx := bytes.Index(txtslic, []byte("xref"))
	if idx == -1 {return -1, fmt.Errorf("readLine cannot parse \"xref\": %s",string(txtslic))}

	// line 2: get first Obj and number of objects
	txtslic, nextPos, err = pdf.readLine(nextPos)
	if err != nil {return -1, fmt.Errorf("readLine for start Obj and num Obj: %v", err)}

	_, err = fmt.Sscanf(string(txtslic), "%d %d", &objStart, &objNum)
	if err != nil {return -1, fmt.Errorf("readLine cannot parse 2nd line objstart objnum: %v", err)}
	pdf.xrefPos = 1
	if objStart > 0 {pdf.xrefPos =-1}
	// line 3+: get pos of each Obj
	if pdf.Dbg {fmt.Printf("dbg -- objStart: %d objNum: %d\n", objStart, objNum)}
	if 	pdf.ObjList == nil {
		objlist := make([]pdfObj, (objNum + objStart))
		pdf.ObjList = &objlist
		pdf.NumObj = objNum + objStart
	}

	val2:=-1
	endStr := ""
	objPos := 0
	for i := 0; i< objNum; i++ {
		txtslic, nextPos, err = pdf.readLine(nextPos)
		if err != nil {return -1, fmt.Errorf("readLine for obj %d: %v", i, err)}

		_, err = fmt.Sscanf(string(txtslic), "%d %d %s", &objPos, &val2, &endStr)
		if err != nil {	return -1, fmt.Errorf("parsing Obj %d: %v", i, err)}
		(*pdf.ObjList)[objStart + i].BufPos = objPos
	}

	if pdf.Dbg {
		fmt.Printf("dbg: **** objlist[%d] ******\n", pdf.NumObj)
    	for i := 0; i< pdf.NumObj; i++ {
			obj := (*pdf.ObjList)[i]
			fmt.Printf("%d: %d\n", i, obj.BufPos)
		}
	}
	return nextPos, nil
}

/*
func (pdf *InfoPdf) parseContent(instr string, pgNum int)(outstr string, err error) {

fmt.Println("**** Content ***")
	outstr = fmt.Sprintf("**** KV ****\n")

	objStr, err := pdf.getKVStr(instr)
	if err != nil {
		outstr += fmt.Sprintf("// getVkStr error: %v\n", err)
		return outstr, fmt.Errorf("getVkStr error: %v", err)
	}


//fmt.Println("***** objStr parsePage")
//fmt.Println(objStr)
//fmt.Println("***** end objstr")

	kvm, err := pdf.getKvMap(objStr)
	if err != nil {
		outstr += fmt.Sprintf("getVkMap error: %v", err)
		return outstr, fmt.Errorf("getVkMap error: %v", err)
	}


	for key, val := range kvm {
		outstr += fmt.Sprintf("key: %s value: %s\n", key, val)
		fmt.Printf("key: %s value: %s\n", key, val)
	}

fmt.Println("*** end Content kv ")

	streamStr, err := pdf.getStream(instr)
	if err != nil {
		outstr += streamStr + fmt.Sprintf("stream deflate error: %v\n", err)
		return outstr, fmt.Errorf("stream deflate error: %v", err)
	}
	if len(streamStr) == 0 {
		outstr += "no stream\n"
		return outstr, nil
	}

//	outstr += streamStr

fmt.Printf("stream length: %d\n", len(streamStr))
	outstr += fmt.Sprintf("**** stream [length: %d] ****\n", len(streamStr))

	stbuf := []byte(streamStr)

	bytStream := bytes.NewReader(stbuf)

	bytR, err := zlib.NewReader(bytStream)
	if err != nil {
		outstr += fmt.Sprintf("stream deflate error: %v\n", err)
		return outstr, fmt.Errorf("stream deflate error: %v", err)
	}
	nbuf := new(strings.Builder)
	_, err = io.Copy(nbuf, bytR)
	if err != nil {
		outstr += fmt.Sprintf("stream copy error: %v\n", err)
		return outstr, fmt.Errorf("stream copy error: %v", err)
	}

	bytR.Close()

	outstr += nbuf.String()

fmt.Printf("stream:\n%s\n****\n", nbuf.String())
fmt.Println("***** end streamstr")

	return outstr, nil
}

*/


// method that decodes pdf file to text
// parsing sequence
// Step 1: top 2 lines
// Step 1a: check whether file is linearized
//			if so
// Step 2: last 3 lines
// Step 3: trailer section
// Step 4: xref section
// Step 5: objs
// Step 6: parseInfo
// Step 6: parseRoot
// Step 7: parsePages
// Step 8: for each page: parsePage
// Step 9: for each page: parseContent
// Step 10: parse each Font object
// Step 11: parse each related FontDescriptor object
// Step 12: parse each gstate object
// Step 13: parse each xobject object
//

/*
	txtFil, err := os.Create(txtfil)
	if err != nil {return fmt.Errorf("error creating textFile %s: %v\n", txtfil, err);}
	defer txtFil.Close()
	log.Printf("created text file\n")

	pdf.txtFil = txtFil

	buf := make([]byte,pdf.filSize)
	pdf.buf = &buf

	_, err = (pdf.fil).Read(buf)
	if err != nil {return fmt.Errorf("error Read: %v", err)}
	(pdf.fil).Close()
	log.Printf("read pdf file\n")/
*/

func (pdf *ParsePdf) ParsePdfDoc()(err error) {

	buf:= *pdf.buf
	err = pdf.parseTopTwoLines()
	if err != nil {return fmt.Errorf("parseTopTwoLines: %v", err)}
	log.Printf("success parsing top two lines!\n")
//fmt.Printf("**** first two lines ***\n%s\n",outstr)

	// last 3 lines:
	// startxref
	// [byte pos]
	// %%EOF
	xrefPos, err := pdf.parseLast3Lines()
	if err != nil {return fmt.Errorf("parseLast3Lines: %v", err)}
	log.Printf("parsed last3Line xref: %d\n", xrefPos)

	log.Printf("startXrefPos: %d\n", xrefPos)

	// parse xref tables and trailers
	numXref := 1

	for i:=0; i< numXref; i++ {

		trailPos, err := pdf.parseXref(xrefPos)
		if err != nil {return fmt.Errorf("parseXref: %v", err)}

		log.Printf("trailer start Pos: %d\n", trailPos)

		xrefPos, err = pdf.parseTrailer(trailPos)
		if err != nil {return fmt.Errorf("parseTrailer: %v", err)}

		if xrefPos == -1 { break}

		numXref++
		log.Printf("xrefpos[%d]: %d\n", numXref, xrefPos)
	}

	log.Printf("trailers: %d\n", numXref)
	pdf.NumTrailer = numXref

	// peek ahead to see whether doc is linearized
	firstObjSt, firstObjEnd, err := pdf.findNextObj(10)
	if err != nil {return fmt.Errorf("findNextObj: %v", err)}

	log.Printf("first obj[%d:%d]\n",firstObjSt, firstObjEnd)

	dblStart, dblEnd, err := pdf.parseDblBracket(firstObjSt, firstObjEnd - firstObjSt)
	if err != nil {return fmt.Errorf("findDblBracket: %v", err)}

	log.Printf("bracket[%d:%d]: %s\n", dblStart, dblEnd,string(buf[dblStart:dblEnd+1]))

		fObjCont, err := pdf.parseDict(buf[firstObjSt:firstObjEnd])
		if err != nil {return fmt.Errorf("parseDict first Obj: %v", err)}
	fmt.Printf("dbg -- first obj: %s\n", string(fObjCont))

		fObjDict, err := pdf.parseDictCont(fObjCont)
		if err != nil {return fmt.Errorf("parseDictCont first Obj: %v", err)}

		fmt.Printf("********* dbg -- first obj keys: %d\n", len(fObjDict))
		for k, v := range fObjDict {
			fmt.Printf("  key: %-15s val: %s Typ: %d\n",k, v.valStr, v.valTyp)
		}

	docLin, ok := fObjDict["Linearized"]
	if ok {
		fmt.Printf("doc linearized: %s\n", docLin.valStr)
	}

	// create a list of objects
	log.Printf("****** parsing object list *****\n")
	err = pdf.parseObjList()
	if err != nil {return fmt.Errorf("parseObjList: %v", err)}


	// loop through the object list and create a dict obj for each pdf object
	log.Printf("parsing object dict\n")
	dictSlic := []byte{}
	for i:=1; i<pdf.NumObj; i++ {

		// get the dict slice for each object
		obj:= (*pdf.ObjList)[i]
		if obj.stream {
			dictSlic = buf[obj.BufPos:obj.streamSt]
		} else {
			dictSlic = buf[obj.BufPos:obj.EndPos]
		}
//	fmt.Printf("dbg -- obj[%d]: %s\n", i, string(dictSlic))
		
		// parse the dictSlice
		dictCont, err := pdf.parseDict(dictSlic)
		if err != nil {return fmt.Errorf("parseDict Obj[%d]: %v", i, err)}
	fmt.Printf("dbg -- obj[%d]: %s\n", i, string(dictCont))

		dictObj, err := pdf.parseDictCont(dictCont)
		if err != nil {return fmt.Errorf("parseDictCont Obj[%d]: %v", i, err)}

		(*pdf.ObjList)[i].dictMap = &dictObj

		fmt.Printf("********* dbg -- obj[%d]: keys: %d\n", i, len(dictObj))
		for k, v := range dictObj {
			fmt.Printf("  key: %-15s val: %s Typ: %d\n",k, v.valStr, v.valTyp)
		}

		// check whether dictMap contants the key Type
		keyVal, ok := dictObj["Type"]
		if ok {
			fmt.Printf("Obj[%d]: %d %s\n", i, keyVal.valTyp, keyVal.valStr)
			(*pdf.ObjList)[i].objTyp = keyVal.valTyp
			(*pdf.ObjList)[i].objTypStr = keyVal.valStr
		}

//		keyValStr, err := pdf.parseKeyType(i, )
//		if err != nil {return fmt.Errorf("parseDictMap Obj[%d]: %v", i, err)}
	}

	pdf.PrintObjList()

	return nil
}


// method that parses the type key
func (pdf *ParsePdf) parseKeyType(i int, valstr string)(err error) {

	switch valstr {
	case "Catalog":
		pdf.RootId = i

	case "Page":


	case "Pages":
		pdf.PagesId = i

	case "Font":


	case "FontDescriptor":

	default:

	}

	return nil
}

// method that parses a dictionary
func (pdf *ParsePdf) parseDictCont(dictObj []byte)(dictMap objDict, err error) {

	// states
	// 0: find key start
	// 1: find key end
	// 2: find val start
	// 3: find value end
	//	  reset to state to 0

	state := 0
	keySt := -1  //key start position
	keyEnd := -1 //key end position
	valSt := -1  //value start position
	valEnd := -1 //value end position
	objTyp := -1 // vslue type
	dictNest := 0
	arrayNest :=0
	roundNest := 0
//xx
	dictMap = make(map[string]*dictVal)

	for i:=0; i< len(dictObj); i++ {
		switch state {
		case 0:
			if dictObj[i] == '/' {
				keySt = i+1
				state = 1
			}
		case 1:
			if dictObj[i] == ' ' {
				keyEnd = i
				valSt = i+1
				state = 2
				objTyp = 2
				break
			}
			if dictObj[i] == '/' {
				keyEnd = i
				state = 2
				valSt = i+1
				objTyp = 1
				break
			}

			if dictObj[i] == '[' {
//				keySt = i+1
				keyEnd = i
				state = 5
				valSt = i
				objTyp = 5
				arrayNest++
				break
			}
			if dictObj[i] == '(' {
//				keySt = i+1
				keyEnd = i
				state = 6
				valSt = i
				objTyp = 6
				roundNest++
				break
			}

			if dictObj[i] == '<' {
				if dictObj[i+1] == '<' {
					keyEnd = i
					state = 4
					valSt = i+2
					objTyp = 4
					dictNest++
					break
				}
				return dictMap, fmt.Errorf("unviable dict val after key: %s", string(dictObj[keySt:keyEnd+1]))
			}

		case 2:
			if dictObj[i] == '/' {
				valEnd = i
				keystr := string(dictObj[keySt:keyEnd])
				valstr := string(dictObj[valSt:valEnd])
//fmt.Printf("dbg -- key: %s val: %s Type: %d\n", keystr, valstr, objTyp)

				dictMap[keystr] = &dictVal{
					valTyp: objTyp,
					valStr: valstr,
				}

				keySt = i+1
				keyEnd = -1
				valSt = -1
				valEnd = -1
				state = 1
				break
			}
		// dict
		case 4:
			if dictObj[i] == '<' {
				if dictObj[i+1] == '<' {
					dictNest++
				}
			}
			if dictObj[i] == '>' {
				if dictObj[i+1] == '>' {
					dictNest--
				}
			}
			if dictNest == 0 {
				valEnd = i
				keystr := string(dictObj[keySt:keyEnd])
				valstr := string(dictObj[valSt:valEnd])
//fmt.Printf("dbg -- key: %s val: %s Type: %d\n", keystr, valstr, objTyp)

				dictMap[keystr] = &dictVal{
					valTyp: objTyp,
					valStr: valstr,
				}

				keySt = -1
				keyEnd = -1
				valSt = -1
				valEnd = -1
				state = 0
			}

		case 5:
			if dictObj[i] == '[' {
					arrayNest++
				}
			if dictObj[i] == ']' {
					arrayNest--
				}

			if arrayNest == 0 {
				valEnd = i
				keystr := string(dictObj[keySt:keyEnd])
				valstr := string(dictObj[valSt+1:valEnd])
				dictMap[keystr] = &dictVal{
					valTyp: objTyp,
					valStr: valstr,
				}
				state = 0
				keySt = -1
				keyEnd = -1
				valSt = -1
				valEnd = -1
				state = 0
			}
		case 6:
			if dictObj[i] == ')' {
					roundNest--
				}

			if roundNest == 0 {
				valEnd = i
				keystr := string(dictObj[keySt:keyEnd])
				valstr := string(dictObj[valSt+1:valEnd])
				dictMap[keystr] = &dictVal{
					valTyp: objTyp,
					valStr: valstr,
				}
				state = 0
				keySt = -1
				keyEnd = -1
				valSt = -1
				valEnd = -1
				state = 0
			}

		default:
			return dictMap, fmt.Errorf("unviable state: %d", state)
		}
	}

	if state == 2 {
		valEnd = len(dictObj)
		keystr := string(dictObj[keySt:keyEnd])
		valstr := string(dictObj[valSt:valEnd])

//fmt.Printf("dbg -- key: %s val: %s Type: %d\n", keystr, valstr, objTyp)

		dictMap[keystr] = &dictVal{
			valTyp: objTyp,
			valStr: valstr,
		}
	}


	return dictMap, nil
}


/*
func (pdf *InfoPdf) parseObjLin(linSl []byte, istate int )(newstate int, err error) {

	var pdfobj pdfObj
	var objList []pdfObj

//fmt.Println("line: ",string(linSl))
	if (pdf.rdObjList) != nil {
		objList = *pdf.rdObjList
	}

	cObjIdx := len(objList) -1

	nxtLin := false
	for is:= 0; is<5; is++ {

		switch istate {
		case 0:
		// obj begin
			id:=0
			_, scerr := fmt.Sscanf(string(linSl), "%d 0 obj", &id)
			if scerr != nil {return istate, fmt.Errorf("obj start scan %s error: %v", string(linSl), scerr)}
//			pdfobj.objId = id
			objList = append(objList, pdfobj)
			istate++
			nxtLin = true
		case 1:
		// obj type
			idx := bytes.Index(linSl, []byte("/Type"))
			if idx == -1 {
				objList[cObjIdx].typ = Data
			} else {
				tpos :=0
				for i:= 5; i< len(linSl);i++ {
					if linSl[i] == '/' {
						tpos = i
						break
					}
				}
				if tpos ==0 {return 0, fmt.Errorf("no type property found in %s", string(linSl))}
				tepos := len(linSl)-1
				for i:= tpos + 1; i< len(linSl);i++ {
					if linSl[i] == '/' || linSl[i] == ' ' {
						tepos = i
						break
					}
				}
				objTyp := string(linSl[tpos+1: tepos])
fmt.Printf("obj %d id %d type %s\n", cObjIdx, objList[cObjIdx].bufPos, objTyp)
			}
			istate++
			nxtLin = true
		case 2:
		// stream
			istate = 4
			nxtLin = false
		case 3:
		// endstream

		case 4:
		// end obj
			idx := bytes.Index(linSl, []byte("endobj"))
//return 0, fmt.Errorf("no endobj found in %s", string(linSl))}

			nxtLin = true
			if idx> -1 {istate = 0}
		default:
			return istate, fmt.Errorf("invalid istate: %d", istate)
		}
		if nxtLin {break}
	} // is
	pdf.rdObjList = &objList
	newstate = istate
	return newstate, nil
}

func (pdf *InfoPdf) parseRoot()(err error) {

	buf := *pdf.buf

	if pdf.rootId > pdf.numObj {return fmt.Errorf("invalid rootId!")}
	if pdf.rootId ==0 {return fmt.Errorf("rootId is 0!")}

	obj := (*pdf.objList)[pdf.rootId]

	objByt := buf[obj.contSt:obj.contEnd]
	objId, err := pdf.parseObjRef("/Pages",objByt)
	if err != nil {return fmt.Errorf("Root obj: parsing name \"/Pages\" error: %v!", err)}

	pdf.pagesId = objId

	outstr, err := pdf.parseName("Title", objByt)
	if err != nil {}
//return fmt.Errorf("Root obj: parsing name \"/Title\" error: %v!", err)}

fmt.Printf("Title: %s\n", outstr)
	return nil
}


func (pdf *InfoPdf) parseKeyText(key string, obj pdfObj)(outstr string, err error) {

	buf:= *pdf.buf
	objByt := buf[obj.contSt:obj.contEnd]
//fmt.Printf("found key: %s in %s\n", key, string(objByt))

	keyByt := []byte("/" + key)
	ipos := bytes.Index(objByt, keyByt)
	if ipos == -1 {return fmt.Sprintf("no /%s",key), fmt.Errorf("could not find keyword \"%s\"!", key)}

	valSt:= obj.contSt + ipos + len(keyByt) + 1
	rootEnd := -1
	for i:=valSt; i< obj.contEnd; i++ {
		switch buf[i] {
		case '/','\n','\r':
			rootEnd = i
			break
		default:
		}
	}

	if rootEnd == -1 {return fmt.Sprintf("/%s no value", key), fmt.Errorf("cannot find end delimiter after key %s", key)}

	outstr = string(buf[valSt:rootEnd])

	return outstr, nil
}

//page
func (pdf *InfoPdf) parsePage(iPage int)(err error) {

	var pgobj pgObj
//	buf := *pdf.buf
	buf := *pdf.buf

	pgobj.pageNum = iPage + 1

	//determine the object id for page iPage
	pgobjId := (*pdf.pageIds)[iPage]
	txtFil := pdf.txtFil

	outstr := fmt.Sprintf("***** Page %d: id %d *******\n", iPage+1, pgobjId)
	fmt.Printf(outstr)
	txtFil.WriteString(outstr)

	// get obj for page[iPage]
	obj := (*pdf.objList)[pgobjId]

//fmt.Printf("testing page %d string:\n%s\n", iPage+1, string(buf[obj.start: obj.end]))

	objByt := buf[obj.contSt:obj.contEnd]
	objId, err := pdf.parseObjRef("/Contents",objByt)
	if err != nil {return fmt.Errorf("parse \"Contents\" error: %v!", err)}

	outstr = fmt.Sprintf("Contents Obj: %d\n", objId)
	fmt.Printf(outstr)
	txtFil.WriteString(outstr)

	pgobj.contentId = objId

	mbox, err := pdf.parseMbox(obj)
	if err!= nil {
		pdf.txtFil.WriteString("no Name \"/MediaBox\" found!\n")
		fmt.Println("no Name \"/MediaBox\" found!")
	} else {
		pgobj.mediabox = mbox
		outstr = fmt.Sprintf("MediaBox: %.1f %.1f %.1f %.1f\n", mbox[0], mbox[1], mbox[2], mbox[3])
		txtFil.WriteString(outstr)
		fmt.Println(outstr)
	}

	reslist, err := pdf.parseResources(obj)
//fmt.Printf("page %d reslist: %v\n", iPage, reslist)
	if err != nil {
		outstr := fmt.Sprintf("parsing error \"/Resources\": %v!\n", err)
		pdf.txtFil.WriteString(outstr)
		fmt.Println(outstr)
	}

	if reslist != nil {
		if reslist.fonts != nil {pgobj.fonts = reslist.fonts}
		if reslist.gStates != nil {pgobj.gStates = reslist.gStates}
		if reslist.xObjs != nil {pgobj.xObjs = reslist.xObjs}
	}

//fmt.Printf("page %d: %v\n", iPage, pgobj)
	(*pdf.pageList)[iPage] = pgobj

	return nil
}


func (pdf *InfoPdf) parsePageContent(iPage int)(err error) {

	pgobj := (*pdf.pageList)[iPage]

	contId := pgobj.contentId

fmt.Printf("content obj: %d\n", contId)

	obj := (*pdf.objList)[contId]

	buf := *pdf.buf
//fmt.Printf("cont obj [%d:%d]:\n%s\n", obj.contSt, obj.contEnd, string(buf[obj.contSt:obj.contEnd]))

	txtFil := pdf.txtFil

	outstr := fmt.Sprintf("***** Content Page %d: id %d *******\n", iPage+1, contId)
	fmt.Printf(outstr)
	txtFil.WriteString(outstr)

	//Filter
	dictByt := buf[obj.contSt+2: obj.contEnd-2]
	key:="/Filter"

	filtStr, err := pdf.parseName(key, dictByt)
	if err != nil {return fmt.Errorf("parseName error parsing value of %s: %v",key, err)}
//fmt.Printf("key %s: val %s\n", key, filtStr)
	if filtStr != "FlateDecode" {
		outstr = fmt.Sprintf("Filter %s not implemented!\n", filtStr)
		fmt.Printf(outstr)
		txtFil.WriteString(outstr)
	}

	key = "/Length"
	streamLen, err := pdf.parseInt(key, dictByt)
	if err != nil {return fmt.Errorf("parseInt error: parsing value of %s: %v",key, err)}
//fmt.Printf("key %s: %d\n", key, streamLen)
	altStreamLen := obj.streamEnd - obj.streamSt
	if streamLen != altStreamLen {
		outstr = fmt.Sprintf("stream length inconsistent! Obj: %d Calc: %d\n", streamLen, altStreamLen)
		fmt.Printf(outstr)
		txtFil.WriteString(outstr)
	}

	// decode stream

	stream, err := pdf.decodeStream(contId)
	if err != nil {return fmt.Errorf("decodeStream: %v")}

fmt.Println("stream: ", string(*stream))

	txtFil.WriteString(string(*stream))

	return nil
}

func (pdf *InfoPdf) parseFont(objId int)(err error) {

	obj := (*pdf.objList)[objId]

	buf := *pdf.buf
	objByt := buf[obj.contSt:obj.contEnd]

fmt.Printf("font obj [%d:%d]:\n%s\n", obj.contSt, obj.contEnd, string(objByt))

	txtFil := pdf.txtFil

	outstr := fmt.Sprintf("********* Font: id %d *******\n", objId)
	fmt.Printf(outstr)
	txtFil.WriteString(outstr)

	key := "/Subtype"
	valStr, err := pdf.parseName(key, objByt)
	if err != nil {return fmt.Errorf("%s: parseName: %v", key, err)}

fmt.Printf("key: %s val: %s\n",key ,valStr)

	key ="/BaseFont"
	valStr, err = pdf.parseName(key, objByt)
	if err != nil {return fmt.Errorf("%s: parseName %v",key, err)}

fmt.Printf("key: %s val: %s\n", key, valStr)

	key = "/FontDescriptor"
	robjId, err := pdf.parseObjRef(key, objByt)
	if err != nil {return fmt.Errorf("%s: parseObjRef: %v",key ,err)}

fmt.Printf("key: %s val: %d\n", key, robjId)

	return nil
}
*/
/*
func (pdf *InfoPdf) parsePages()(err error) {

	if pdf.pagesId > pdf.numObj {return fmt.Errorf("invalid pagesId!")}
	if pdf.pagesId ==0 {return fmt.Errorf("pagesId is 0!")}

	obj := (*pdf.objList)[pdf.pagesId]

//fmt.Printf("pages:\n%s\n", string(buf[obj.start: obj.end]))

	err = pdf.parseKids(obj)
	if err!= nil {return fmt.Errorf("parseKids: %v", err)}

	if pdf.verb {
		fmt.Printf("pages: pageCount: %d\n", pdf.pageCount)
		for i:=0; i< pdf.pageCount; i++ {
			fmt.Printf("page: %d objId: %d\n", i+1, (*pdf.pageIds)[i])
		}
	}
	pageList := make([]pgObj, pdf.pageCount)

	mbox, err := pdf.parseMbox(obj)
	if err!= nil {
		pdf.txtFil.WriteString("no Name \"/MediaBox\" found!\n")
		fmt.Println("no Name \"/MediaBox\" found!")
	}
	pdf.mediabox = mbox

	reslist, err := pdf.parseResources(obj)
	if err!= nil {
		pdf.txtFil.WriteString("no Name \"/Resources\" found!\n")
		fmt.Println("no Name \"/Resources\" found!")
	}

//fmt.Printf("resList: %v\n", reslist)

	if reslist != nil {
		if reslist.fonts != nil {pdf.fonts = reslist.fonts}
		if reslist.gStates != nil {pdf.gStates = reslist.gStates}
		if reslist.xObjs != nil {pdf.gStates = reslist.xObjs}
	}
	pdf.pageList = &pageList
	return nil
}
*/

/*
func (pdf *InfoPdf) parseResources(obj pdfObj)(resList *resourceList, err error) {

	var reslist resourceList

	buf := *pdf.buf
	objByt := buf[obj.contSt:obj.contEnd]
//fmt.Printf("Resources: %s\n",string(objByt))

	dictBytPt, err := pdf.parseDict("/Resources",objByt)
	if err != nil {return nil, fmt.Errorf("parseDict: %v", err)}
	dictByt := *dictBytPt
fmt.Printf("*** Resources dict ***\n%s\n", string(dictByt))
/*
	idx := bytes.Index(objByt, []byte("/Resources"))
	if idx == -1 {
		if pdf.verb {fmt.Printf("parseResource: cannot find keyword \"/Resources\"!\n")}
		return nil, nil
	}

	// either indirect or a dictionary
	valst := obj.contSt + idx + len("/Resources")
	objByt = buf[valst: obj.contEnd]
//fmt.Printf("Resources valstr [%d:%d]: %s\n",valst, obj.contEnd, string(objByt))
//ddd

	dictSt := bytes.Index(objByt, []byte("<<"))
//fmt.Printf("dictSt: %d\n", dictSt)

	if dictSt == -1 {
//fmt.Println("Resources: indirect obj")
		valend := -1
		for i:= valst; i< obj.contEnd; i++ {
			if buf[i] == 'R' {
				valend = i+1
				break
			}
		}
		if valend == -1 {return nil, fmt.Errorf("cannot find R for indirect obj of \"/Resources\"")}
		inObjStr := string(buf[valst:valend])

//fmt.Printf("ind obj: %s\n", inObjStr)

		objId :=0
		rev := 0
		_, err = fmt.Sscanf(inObjStr,"%d %d R", &objId, &rev)
		if err != nil{return nil, fmt.Errorf("cannot parse %s as indirect obj of \"/Resources\": %v", inObjStr, err)}

		fmt.Printf("Resource Id: %d\n", objId)
//todo find resource string
		return nil, nil
	}

	if pdf.verb {fmt.Println("**** Resources: dictionary *****")}

	resByt := buf[valst: obj.contEnd -2]
//fmt.Printf("Resources valstr [%d:%d]: %s\n",valst, obj.contEnd-2, string(resByt))

	// short cut to be fixed by parsing nesting levels
	dictEnd := bytes.LastIndex(resByt, []byte(">>"))
	if dictEnd == -1 {return nil, fmt.Errorf("no end brackets for dict!")}

	tByt := buf[valst: valst+dictEnd-2]
	tEnd := bytes.LastIndex(tByt, []byte(">>"))
	tSt := bytes.LastIndex(tByt, []byte("<<"))
	if tEnd < tSt {dictEnd = tEnd}

	if dictEnd == -1 {return nil, fmt.Errorf("no end brackets for dict!")}

	dictSt += valst +2
	dictEnd += valst
	dictByt := buf[dictSt:dictEnd]
//fmt.Printf("Resource Dict [%d: %d]:\n%s\n", dictSt, dictEnd, string(dictByt))
//fmt.Println()

	// find Font
	if pdf.verb {fmt.Println("**** Font: dictionary *****")}
	fidx := bytes.Index(dictByt, []byte("/Font"))
	if fidx == -1 {
		if pdf.verb {fmt.Println("parseResources: no keyword \"/Font\"!")}
		reslist.fonts = nil
	} else {
//rrr
		objrefList, objId, err := pdf.parseIObjRefList("Font", &dictByt)
		if err != nil {return nil, fmt.Errorf("cannot get objList for \"/Font\": %v!", err)}
		reslist.fonts = objrefList
		if pdf.verb {fmt.Printf("fonts [%d]: %v\n",objId ,objrefList)}
	}
fmt.Printf("reslist: %v\n", reslist)

	// ExtGState
	if pdf.verb {fmt.Println("**** ExtGstate: dictionary *****")}
	gidx := bytes.Index(dictByt, []byte("/ExtGState"))
	if gidx == -1 {
		if pdf.verb {fmt.Println("parseResources: no keyword \"/ExGState\"!")}
		reslist.gStates = nil
	} else {
		objrefList, objId, err := pdf.parseIObjRefList("ExtGState", &dictByt)
		if err != nil {return nil, fmt.Errorf("cannot get objList for \"/ExtGState\": %v!", err)}
		reslist.gStates = objrefList
		if pdf.verb {fmt.Printf("ExtGState [%d]: %v\n",objId ,objrefList)}
	}

fmt.Printf("reslist: %v\n", reslist)


	// find XObject
	if pdf.verb {fmt.Println("**** XObject: dictionary *****")}
	xidx := bytes.Index(dictByt, []byte("/XObject"))
	if xidx == -1 {
		if pdf.verb {fmt.Println("parseResources: no keyword \"/XObject\"!")}
		reslist.xObjs = nil
	} else {
		objrefList, objId, err := pdf.parseIObjRefList("XObject", &dictByt)
		if err != nil {return nil, fmt.Errorf("cannot get objList for \"/XObject\": %v!", err)}
		reslist.xObjs = objrefList
		if pdf.verb {fmt.Printf("XObject [%d]: %v\n",objId ,objrefList)}
	}

fmt.Printf("reslist: %v\n", reslist)

	// ProcSet
fmt.Println("\n**** ProcSet: array *****")

	pidx := bytes.Index(dictByt, []byte("/ProcSet"))
	if pidx == -1 {return nil, fmt.Errorf("cannot find keyword \"/ProcSet\"")}


	pvalst := pidx + len("/ProcSet")
	pByt := dictByt[pvalst:]
//fmt.Printf("ProcSet valstr [%d:%d]: %s\n",pvalst, dictEnd, string(pByt))

	parrSt := bytes.Index(pByt, []byte("["))
//fmt.Printf("font dictSt: %d\n", fdictSt)

	parrEnd := bytes.Index(pByt, []byte("]"))
	if parrEnd == -1 {return nil, fmt.Errorf("no end bracket for ProcSet array!")}

	parrSt += pvalst +1
	parrEnd += pvalst
	parrByt := pByt[parrSt:parrEnd]
fmt.Printf("ProcSet Array [%d: %d]: %s\n", parrSt, parrEnd, string(parrByt))

	return &reslist, nil
}
*/
/*
func (pdf *InfoPdf) parseIObjRefList(keyname string, dictbyt *[]byte)(objlist *[]objRef, objId int, err error) {

	var keyDictByt []byte
	dictByt := *dictbyt
	objId = -1

//fmt.Printf("****  parsing dictionary for %s *****\n", keyname)
//fmt.Printf("dict:\n%s\n", string(dictByt))

	keyByt := []byte("/" + keyname)
	fidx := bytes.Index(dictByt, []byte(keyByt))
	if fidx == -1 {return nil, -1, fmt.Errorf("cannot find keyword \"/%s\"",keyname)}

	dictEnd := len(dictByt)
	fvalst := fidx + len(keyByt)

	valByt := dictByt[fvalst: dictEnd]
//fmt.Printf("font valstr [%d:%d]: %s\n",fvalst, dictEnd, string(valByt))

	fdictSt := bytes.Index(valByt, []byte("<<"))
//fmt.Printf("font dictSt: %d\n", fdictSt)

	if fdictSt == -1 {
		if pdf.verb {fmt.Printf("%s: indirect obj\n", keyname)}
		fvalend := -1
		for i:= 0; i< len(valByt); i++ {
			if valByt[i] == 'R' {
				fvalend = i+1
				break
			}
		}
		if fvalend == -1 {return nil, -1, fmt.Errorf("cannot find R for indirect obj of \"/%s\"", keyname)}
		inObjStr := string(valByt[:fvalend])

//fmt.Printf("ind obj: %s\n", inObjStr)

		rev := 0
		_, err = fmt.Sscanf(inObjStr,"%d %d R", &objId, &rev)
		if err != nil{return nil, -1, fmt.Errorf("cannot parse %s as indirect obj of \"/%s\": %v", inObjStr, keyname, err)}

		if pdf.verb {fmt.Printf("%s indirect Obj Id: %d\n", keyname, objId)}

		objSl, err := pdf.getObjCont(objId)
		if err != nil{return nil, -1, fmt.Errorf("cannot get content of obj %d: %v", objId, err)}

		valByt = *objSl
		fdictSt = bytes.Index(valByt, []byte("<<"))

	}

fmt.Printf("%s: valstr [%d:]: %s\n",keyname, fdictSt, string(valByt))

	fdictEnd := bytes.Index(valByt, []byte(">>"))
	if fdictEnd == -1 {return nil, -1, fmt.Errorf("no end brackets for dict of %s!", keyname)}

	keyDictByt = valByt[fdictSt +2 :fdictEnd]


//fmt.Printf("%s key Dict: %s\n", keyname, string(keyDictByt))

	objList, err := pdf.parseIrefCol(&keyDictByt)
	if err != nil {return nil, objId, fmt.Errorf("%s parsing ref objs error: %v", keyname, err)}

	return objList, objId, nil
}
*/
/*
func (pdf *InfoPdf) parseIrefCol(inbuf *[]byte)(refList *[]objRef, err error) {

	var objref objRef
	var reflist []objRef

	buf := *inbuf

	val := 0
	objId := -1

	refCount := 0
	istate := 0

	objEnd := -1
	namSt:= -1
	namEnd := -1

	for i:= 0; i< len(buf); i++ {

		switch istate {
		case 0:
		// look for start of obj name
			if buf[i] == '/' {
				namSt = i+1
				istate = 1
			}
		case 1:
		// look for end of obj name
			if buf[i] == ' ' {
				namEnd = i
				istate = 2
			}

		case 2:
		// look for end of obj reference
			if buf[i] == 'R' {
				objEnd = i+1
				refCount++
//fmt.Printf(" inobjref: \"%s\"\n", string(buf[namEnd:objEnd]))
				_, errsc := fmt.Sscanf(string(buf[namEnd:objEnd])," %d %d R", &objId, &val)
				if errsc != nil {return nil, fmt.Errorf("parse obj ref error of obj %d: %v", refCount, errsc)}

				if namEnd< namSt {return nil, fmt.Errorf("parse obj name error of obj %d!", refCount)}
				objref.Id = objId
				objref.Nam = string(buf[namSt:namEnd])
				reflist = append(reflist, objref)
				istate = 0
			}
		default:
		}
	}


	return &reflist, nil
}
*/

/*
func (pdf *InfoPdf) getObjCont(objId int)(objSlice *[]byte, err error) {

	if objId > pdf.numObj {return nil, fmt.Errorf("objIs %d is not valid!", objId)}

	buf := *pdf.buf
	obj := (*pdf.objList)[objId]

	objByt := buf[obj.contSt:obj.contEnd]

	return &objByt, nil
}

func (pdf *InfoPdf) parseKids(obj pdfObj)(err error) {

	buf := *pdf.buf
	objByt := buf[obj.contSt:obj.contEnd]
//fmt.Printf("Kids obj: %s\n",string(objByt))

	idx := bytes.Index(objByt, []byte("/Kids"))
	if idx == -1 {return fmt.Errorf("cannot find keyword \"/Kids\"")}

	opPar := -1
	opEnd := -1

//fmt.Printf("brack start: %s\n", string(buf[(obj.contSt +5 + idx):obj.contEnd]))

	fini := false
	for i:= obj.contSt +idx + 5; i< obj.contEnd; i++ {
		switch buf[i] {
		case '[':
			opPar = i + 1
		case ']':
			opEnd = i
			fini = true
		case '\n', '\r', '/':
			fini = true
		default:
		}
		if fini {break}
	}

	if opEnd <= opPar {fmt.Errorf("no matching square brackets in seq!")}

//fmt.Printf("brack [%d: %d]: \"%s\"\n", opPar, opEnd, string(buf[opPar:opEnd]))

	arBuf := buf[opPar:opEnd]
	pgList, err := pdf.parseIrefArray(&arBuf)
	if err != nil {return fmt.Errorf("parseArray: %v", err)}

	pdf.pageCount = len(*pgList)
	pdf.pageIds = pgList

//	fmt.Printf("Objects: %v Count: %d\n",pgList, len(*pgList))
	return nil
}
*/
/*
func (pdf *InfoPdf) parseIrefArray(inbuf *[]byte)(idList *[]int, err error) {

	var pg []int

	buf := *inbuf

	val := 0
	pgId := -1

	st := 0
	refCount := 0
	iref := -1

	for i:= 0; i< len(buf); i++ {
		if buf[i] == 'R' {
			iref = i
			refCount++
			_, errsc := fmt.Sscanf(string(buf[st:iref+1]),"%d %d R", &pgId, &val)
			if errsc != nil {return nil, fmt.Errorf("scan error obj %d: %v", refCount, errsc)}
			pg = append(pg, pgId)
			st = i+1
		}
	}
	return &pg, nil
}

//aa
func (pdf *InfoPdf) parseMbox(obj pdfObj)(mBox *[4]float32, err error) {

	var mbox [4]float32

	buf := *pdf.buf
	objByt := buf[obj.contSt:obj.contEnd]
//fmt.Printf("Mbox: %s\n",string(objByt))

	idx := bytes.Index(objByt, []byte("/MediaBox"))
	if idx == -1 {return nil, fmt.Errorf("cannot find keyword \"/MediaBox\"")}

	opPar := -1
	opEnd := -1

//fmt.Printf("brack start: %s\n", string(buf[(obj.contSt +5 + idx):obj.contEnd]))

	fini := false
	for i:= obj.contSt +idx + 5; i< obj.contEnd; i++ {
		switch buf[i] {
		case '[':
			opPar = i + 1
		case ']':
			opEnd = i
			fini = true
		case '\n', '\r', '/':
			fini = true
		default:
		}
		if fini {break}
	}

	if opEnd <= opPar {return nil, fmt.Errorf("no matching square brackets in seq!")}

//fmt.Printf("brack [%d: %d]: %s\n", opPar, opEnd, string(buf[opPar:opEnd]))

	// parse references
	_, errsc := fmt.Sscanf(string(buf[opPar:opEnd]),"%f %f %f %f", &mbox[0], &mbox[1], &mbox[2], &mbox[3])
	if errsc != nil {return nil, fmt.Errorf("scan error mbox: %v", errsc)}

//fmt.Printf("mbox: %v\n",mbox)
	return &mbox, nil
}
*/
/*
func (pdf *InfoPdf) parseInt(keyword string, objByt []byte)(num int, err error) {

	var indObj pdfObj

	buf := *pdf.buf

	keyByt := []byte(keyword)
	idx := bytes.Index(objByt, keyByt)
	if idx == -1 {return -1, fmt.Errorf("cannot find keyword \"%s\"", string(keyByt))}

	opSt := -1
	opEnd := -1

	valst := idx+len(keyByt)
	valByt := objByt[valst:]

//fmt.Printf("valstr: %s\n", string(valByt))

	// whether indirect obj
	inObjId := parseIndObjRef(objByt[valst:])
	// todo make sure inObjIs is valid

	if inObjId > -1 {
		indObj = (*pdf.objList)[inObjId]
		valByt = buf[indObj.contSt: indObj.contEnd]
	}

//fmt.Printf("parse num obj valByt: %s\n", string(valByt))

	endByt := []byte{'\n', '\r', '/', ' '}

	istate := 0
	for i:= 0; i< len(valByt); i++ {
		switch istate {
		case 0:
			if util.IsNumeric(valByt[i]) {opSt = i;istate =1}

		case 1:
			if isEnding(valByt[i], endByt) {opEnd= i; istate =2}

		default:
		}
		if istate == 2 {break}
	}

	if istate == 0 {return -1, fmt.Errorf("no number found!")}
	if istate == 1 {opEnd = len(valByt)}

	valBuf := valByt[opSt:opEnd]

//fmt.Printf("key /%s val[%d: %d]: \"%s\"\n", keyword, opSt, opEnd, string(valBuf))

	_, err = fmt.Sscanf(string(valBuf), "%d", &num)
	if err != nil {return -1, fmt.Errorf("cannot parse num: %v", err)}

	return num, nil
}

func (pdf *InfoPdf) parseFloat(keyword string, objByt []byte)(fnum float32, err error) {

	var indObj pdfObj

	buf:= *pdf.buf
	keyByt := []byte(keyword)

	idx := bytes.Index(objByt, keyByt)
	if idx == -1 {return -1.0, fmt.Errorf("cannot find keyword \"%s\"", string(keyByt))}

	opSt := -1
	opEnd := -1
	valst := idx+len(keyByt)
	valByt := objByt[valst:]

//fmt.Printf("valstr: %s\n", string(valByt))

	// whether indirect obj
	inObjId := parseIndObjRef(objByt[valst:])
	// todo make sure inObjIs is valid

	if inObjId > -1 {
		indObj = (*pdf.objList)[inObjId]
		valByt = buf[indObj.contSt: indObj.contEnd]
	}

//fmt.Printf("parse float obj str: %s\n", string(valByt))

	endByt := []byte{'\n', '\r', '/', ' '}

	istate := 0
	for i:= 0; i< len(valByt); i++ {
		switch istate {
		case 0:
			if util.IsNumeric(valByt[i]) {opSt = i;istate =1}

		case 1:
			if isEnding(valByt[i], endByt) {opEnd= i; istate =2}

		default:
		}
		if istate == 2 {break}
	}

	if istate == 0 {return -1, fmt.Errorf("no number found!")}
	if istate == 1 {opEnd = len(valByt)}

	valBuf := valByt[opSt:opEnd]


//fmt.Printf("key %s val[%d: %d]: \"%s\"\n", keyword, opSt, opEnd, string(valBuf))

	_, err = fmt.Sscanf(string(valBuf), "%f", &fnum)
	if err != nil {return -1, fmt.Errorf("cannot parse fnum: %v", err)}

	return fnum, nil
}

func (pdf *InfoPdf) parseString(keyword string, objByt []byte)(outstr string, err error) {

	var indObj pdfObj
	buf := *pdf.buf

	keyByt := []byte(keyword)

	idx := bytes.Index(objByt, keyByt)
	if idx == -1 {return "", fmt.Errorf("cannot find keyword \"%s\"", string(keyByt))}

	opSt := -1
	opEnd := -1
	valst := idx+len(keyByt)
	valByt := objByt[valst:]

fmt.Printf("valstr: %s\n", string(valByt))

	// whether indirect obj
	inObjId := parseIndObjRef(objByt[valst:])
	// todo make sure inObjIs is valid

	if inObjId > -1 {
		indObj = (*pdf.objList)[inObjId]
		valByt = buf[indObj.contSt: indObj.contEnd]
	}

	endByt := []byte{'\n', '\r', '/'}

	istate := 0
	for i:= 0; i< len(valByt); i++ {
		switch istate {
		case 0:
			if valByt[i] == '(' {opSt = i;istate =1}

		case 1:
			if valByt[i] == ')' {opEnd = i;istate =2}

			if isEnding(valByt[i], endByt) {opEnd= i; istate =3}

		default:
		}
		if istate > 1 {break}
	}

	switch istate {
	case 0:
		return "", fmt.Errorf("no open '(' found!")
	case 1, 3:
		return "", fmt.Errorf("no open ')' found!")
	case 2:
		if opEnd -1 < opSt +1 {return "", fmt.Errorf("inverted string [%d:%d]",opSt+1, opEnd-1)}

	default:
		return "", fmt.Errorf("unknown istate %d!", istate)
	}

	valBuf := valByt[opSt+1:opEnd -1]

fmt.Printf("key /%s val[%d: %d]: \"%s\"\n", keyword, opSt, opEnd, string(valBuf))

	_, err = fmt.Sscanf(string(valBuf), "%s", &outstr)
	if err != nil {return "", fmt.Errorf("cannot parse string: %v", err)}

	return outstr, nil
}

func (pdf *InfoPdf) parseName(keyword string, objByt []byte)(outstr string, err error) {

//	var indObj pdfObj
//	buf := *pdf.buf

	keyByt := []byte(keyword)

	idx := bytes.Index(objByt, keyByt)
	if idx == -1 {return "", fmt.Errorf("cannot find keyword \"%s\"", string(keyByt))}

	opSt := -1
	opEnd := -1
	valst := idx+len(keyByt)
	valByt := objByt[valst:]

//fmt.Printf("valstr: %s\n", string(valByt))

	endByt := []byte{'\n', '\r', '/'}
	istate := 0
	for i:= 0; i< len(valByt); i++ {
		switch istate {
		case 0:
			if valByt[i] == '/' {opSt = i;istate =1}

		case 1:
			if isEnding(valByt[i], endByt) {opEnd= i; istate =2}

		default:
		}
		if istate > 1 {break}
	}

	switch istate {
	case 0:
		return "", fmt.Errorf("no open '/' found!")
	case 1:
		opEnd = len(valByt)
	case 2:
		if opEnd -1 < opSt +1 {return "", fmt.Errorf("inverted string [%d:%d]",opSt+1, opEnd-1)}

	default:
		return "", fmt.Errorf("unknown istate %d!", istate)
	}

	valBuf := valByt[opSt:opEnd]

//fmt.Printf("key %s val[%d: %d]: \"%s\"\n", keyword, opSt, opEnd, string(valBuf))

	_, err = fmt.Sscanf(string(valBuf), "/%s", &outstr)
	if err != nil {return "", fmt.Errorf("cannot parse string: %v", err)}

	return outstr, nil
}
*/



/*
func (pdf *InfoPdf) parseObjRef(keyword string, objByt []byte) (objId int, err error) {
// function parses ValByt to find object reference
// if no obj id found return obj Id = -1

	keyByt := []byte(keyword)

	idx := bytes.Index(objByt, keyByt)
	if idx == -1 {return -3, fmt.Errorf("cannot find keyword \"%s\"", string(keyByt))}

	valst := idx+len(keyByt)
	valByt := objByt[valst:]

	valEnd := -1
	for i:=0; i< len(valByt); i++ {
		if valByt[i] == 'R' {
			valEnd = i
			break
		}
	}
	if valEnd == -1 {return -1, fmt.Errorf("cannot find char \"R\"!")}

	ref :=0
	_, err = fmt.Sscanf(string(valByt[:valEnd+1]),"%d %d R", &objId, &ref)
	if err != nil {return -2, fmt.Errorf("cannot parse Obj Ref: %v!",err)}

	return objId, nil
}


func parseIndObjRef(valByt []byte) (objId int) {
// function parses ValByt to find object reference
// if no obj id found return obj Id = -1
	valEnd := -1
	for i:=0; i< len(valByt); i++ {
		if valByt[i] == 'R' {
			valEnd = i
			break
		}
	}
	if valEnd == -1 {return -1}

	ref :=0
	_, err:= fmt.Sscanf(string(valByt[:valEnd+1]),"%d %d R", &objId, &ref)
	if err != nil {return -2}
	return objId
}
*/
/*
func isEnding (b byte, ending []byte)(end bool) {

	idx := -1
	for i:=0; i<len(ending); i++ {
		if ending[i] == b {
			idx = i
			break
		}
	}
	if idx == -1 {return false}
	return true
}

func (pdf *InfoPdf) findKeyWord(key string, obj pdfObj)(ipos int) {

	buf:= *pdf.buf
	objByt := buf[obj.contSt:obj.contEnd]
//fmt.Printf("find key: %s in %s\n", key, string(objByt))

	keyByt := []byte("/" + key)
	ipos = bytes.Index(objByt, keyByt)

	return ipos
}
*/

// method that parses an object to determine dict start/end stream start/end and object type
func (pdf *ParsePdf) parseObjList()(err error) {

	for i:= 1; i< pdf.NumObj; i++ {
		err = pdf.parseObj(i)
		if err != nil {return fmt.Errorf("obj[%d] %v",i, err)}
	}
	return nil
}

func (pdf *ParsePdf) parseObj(objIdx int)(err error) {

	buf := *pdf.buf

	obj := (*pdf.ObjList)[objIdx]
	stPos := obj.BufPos
//fmt.Printf("***Obj: %d: %d *****\n", objIdx, stPos)

	txtslic, _, err := pdf.readLine(stPos)
	if err != nil { return fmt.Errorf("readLine: %v", err)}
//fmt.Printf("dbg next-- obj[%d]: st %d len %d\n", objIdx, stPos, len(txtslic))
//fmt.Printf("dbg next-- \"%s\"\n", string(txtslic[:8]))

	objId, err := pdf.parseObjHead(txtslic)
	if err != nil {return fmt.Errorf("parseObjHead: %v", err)}
	obj.objId = objId

	// find endobj
	idx := bytes.Index(buf[stPos:], []byte("endobj"))
	if idx == -1 {return fmt.Errorf("no endobj found")}
	endPos := stPos + idx
	obj.EndPos = endPos
	// find stream
	idxStream := bytes.Index(buf[stPos: endPos], []byte("stream"))
	if idxStream == -1 {
		obj.stream = false
		(*pdf.ObjList)[objIdx] = obj
		return nil
	}

	streamSt := stPos + idxStream
	obj.streamSt = streamSt

	idx = bytes.Index(buf[streamSt: endPos], []byte("endstream"))
	if idx == -1 {
		obj.streamEnd = -1
		(*pdf.ObjList)[objIdx] = obj
		return fmt.Errorf("no streamend")
	 }

	obj.streamEnd = obj.streamSt + idx
	obj.stream = true

	(*pdf.ObjList)[objIdx] = obj
	return nil
}


func (pdf *ParsePdf) parseObjHead(txt []byte) (objId int, err error){

	val := -1
	txtStr := string(txt)
	_, err = fmt.Sscanf(txtStr, "%d %d obj", &objId, &val)
	if err != nil {return -1, fmt.Errorf("Scan: %v", err)}
	if val != 0 {return -1, fmt.Errorf("objhead val != 0: %d!", val)}

	return objId, nil
}


//rr

func (pdf *ParsePdf) readLine(stPos int)(out []byte, nextPos int, err error) {

	buf := *pdf.buf


	maxPos := stPos + 3000
	if len(buf) < maxPos {maxPos = len(buf)}
//	if pdf.Dbg {fmt.Printf("\nreadLine [%d:%d]:\n%s\n***\n", stPos, maxPos, string(buf[stPos:maxPos]))}

	endPos := -1
	nextPos = -1
	for i:=stPos; i < maxPos; i++ {
//if pdf.Dbg {fmt.Printf("i: %d char: %q\n",i, buf[i])}
		if buf[i] == '\r' {
			endPos = i
			nextPos = i+1
			if buf[i+1] == '\n' {nextPos++}
			break
		}
	}

	if endPos == -1 {endPos = maxPos}

	out = buf[stPos:endPos]
	return out, nextPos, nil
}


func getObjTypStr(ityp int)(s string) {
	switch ityp {
		case 1:
			s = "Title"
		case 2:
			s = "Catalog"
		case 3:
			s = "Pages"
		case 4:
			s = "Page"
		case 5:
			s = "Font"
		case 6:
			s = "Font Desc"
		case 7:
			s = "Data"
		case 8:
			s= "ca"
		default:
			s= "unknown"
	}
	return s
}

/*
func (pdf *InfoPdf) PrintPage (iPage int) {
		pgobj := (*pdf.pageList)[iPage]
		fmt.Printf("****************** Page %d *********************\n", iPage + 1)
		fmt.Printf("Page Number:     %d\n", pgobj.pageNum)
		fmt.Printf("Contents Obj Id: %d\n", pgobj.contentId)
		fmt.Printf("Media Box: ")
		if pgobj.mediabox == nil {
			fmt.Printf("no\n")
		} else {
			mbox := pgobj.mediabox
			for i:=0; i< 4; i++ {fmt.Printf(" %.1f", mbox[i])}
			fmt.Printf("\n")
		}
		fmt.Printf("Resources:\n")
		if pgobj.fonts == nil {
			fmt.Println("-- no Fonts")
		} else {
			fmt.Println("-- Font Ids:")
			for i:=0; i< len(*pgobj.fonts); i++ {
				fmt.Printf("   %s %d\n", (*pgobj.fonts)[i].Nam, (*pgobj.fonts)[i].Id)
			}
		}
		if pgobj.gStates == nil {
			fmt.Println("-- no ExtGstates")
		} else {
			fmt.Println("-- ExtGstate Ids:")
			for i:=0; i< len(*pgobj.gStates); i++ {
				fmt.Printf("   %s %d\n", (*pgobj.gStates)[i].Nam, (*pgobj.gStates)[i].Id)
			}
		}
		if pgobj.xObjs == nil {
			fmt.Println("-- no XObjects")
		} else {
			fmt.Println("-- XObject Ids:")
			for i:=0; i< len(*pgobj.xObjs); i++ {
				fmt.Printf("   %s %d\n", (*pgobj.xObjs)[i].Nam, (*pgobj.xObjs)[i].Id)
			}
		}
		fmt.Println("**********************************************")

}
*/


func (pdf *ParsePdf) PrintPdfDocStruct() {

	fmt.Println("\n******************** Info Pdf **********************\n")
	fmt.Printf("File Name: %s\n", pdf.PdfFilnam)
//	fmt.Printf("File Size: %d\n", pdf.filSize)
	fmt.Println()

	fmt.Printf("pdf version: %d.%d\n",pdf.Majver, pdf.Minver)

	fmt.Printf("obj count: %d\n", pdf.NumObj -1 )

	fmt.Printf("****** obj list *******\n")
	for i:=1; i< pdf.NumObj; i++ {
		obj := (*pdf.ObjList)[i]
		fmt.Printf("  obj[%d] Id: %d Typ: %d TypStr: %s\n", i, obj.objId, obj.objTyp, obj.objTypStr)
	}
	fmt.Printf("**** end obj list *****\n")

/*
	fmt.Printf("Page Count: %3d\n", pdf.pageCount)
	if pdf.mediabox == nil {
		fmt.Printf("no MediaBox\n")
	} else {
		fmt.Printf("MediaBox:    ")
		for i:=0; i< 4; i++ {
			fmt.Printf(" %.2f", (*pdf.mediabox)[i])
		}
		fmt.Println()
	}
	fmt.Printf("Font Count: %d\n", pdf.fCount)
	fmt.Printf("Font Obj Ids:\n")
	for i:=0; i< pdf.fCount; i++ {
		fmt.Printf("%d\n", (*pdf.fontIds)[i])
	}

	if pdf.gStates == nil {
		fmt.Println("-- no ExtGstates")
	} else {
		fmt.Println("-- ExtGstate Ids:")
		for i:=0; i< len(*pdf.gStates); i++ {
			fmt.Printf("   %s %d\n", (*pdf.gStates)[i].Nam, (*pdf.gStates)[i].Id)
		}
	}

	if pdf.xObjs == nil {
		fmt.Println("-- no xObjs")
	} else {
		fmt.Println("-- xObj Ids:")
		for i:=0; i< len(*pdf.xObjs); i++ {
			fmt.Printf("   %s %d\n", (*pdf.xObjs)[i].Nam, (*pdf.xObjs)[i].Id)
		}
	}


	fmt.Println()
	fmt.Println()
	fmt.Printf("Info Id:    %5d\n", pdf.infoId)
	fmt.Printf("Root Id:    %5d\n", pdf.rootId)
	fmt.Printf("Pages Id:   %5d\n", pdf.pagesId)


	fmt.Printf("Xref Loc:      %5d\n", pdf.xrefPos)
	fmt.Printf("trailer Loc:   %5d\n", pdf.trailerPos)
	fmt.Printf("startxref Loc: %5d\n", pdf.startXrefPos)

	fmt.Println()
	fmt.Printf("*********************** xref Obj List [%3d] *********************\n", pdf.numObj)
	fmt.Printf("Objects: %5d    First Object Start Pos: %2d\n", pdf.numObj, pdf.objStart)
	fmt.Println("*****************************************************************")

	if pdf.objList == nil {
		fmt.Println("objlist is nil!")
		return
	}

	fmt.Println("                             Content      Stream")
	fmt.Println("Obj   Id type start  end   Start  End  Start  End  Length Type")
	for i:= 0; i< len(*pdf.objList); i++ {
		obj := (*pdf.objList)[i]
		fmt.Printf("%3d: %3d  %2d  %5d %5d %5d %5d %5d %5d %5d   %-15s\n",
		i, obj.bufPos, obj.typ, obj.start, obj.end, obj.contSt, obj.contEnd, obj.streamSt, obj.streamEnd, obj.streamEnd - obj.streamSt, obj.typstr)
	}
	fmt.Println()

	fmt.Println("*********** sequential Obj List *********************")
	if pdf.objList == nil {
		fmt.Println("rdObjlist is nil!")
		return
	}

	fmt.Println("Obj seq  Id start  type")
	for i:= 0; i< len(*pdf.rdObjList); i++ {
		obj := (*pdf.rdObjList)[i]
		fmt.Printf("%3d   %5d\n", i, obj.bufPos)
	}
	fmt.Println("************************************************")
	fmt.Println()

	for ipg:=0; ipg< pdf.pageCount; ipg++ {
		pdf.PrintPage(ipg)
	}

	for i:=0; i< pdf.fCount; i++ {
		objId := (*pdf.fontIds)[i]
		fmt.Printf("*************** Font Obj %d ******************\n", objId)

	}

	if pdf.fonts == nil {
		fmt.Println("-- no Fonts")
	} else {
		fmt.Println("-- Font Ids:")
		for i:=0; i< len(*pdf.fonts); i++ {
				fmt.Printf("   %s %d\n", (*pdf.fonts)[i].Nam, (*pdf.fonts)[i].Id)
		}
	}
*/
	return
}

func (pdf *ParsePdf) PrintObjList (){

	fmt.Printf("dbg: **** objlist[%d] ******\n", pdf.NumObj)
	fmt.Printf("obj start end id  objType\n")
    for i := 0; i< pdf.NumObj; i++ {
		obj := (*pdf.ObjList)[i]
		fmt.Printf("%d: %d %d %d %d %t %d %d\n", i, obj.BufPos, obj.EndPos, obj.objId, obj.objTyp, obj.stream, obj.streamSt, obj.streamEnd)
	}
	fmt.Printf("dbg: **** end objlist 1 ******\n")
	fmt.Printf("*** obj id type ***\n")
    for i := 0; i< pdf.NumObj; i++ {
		obj := (*pdf.ObjList)[i]
		fmt.Printf("%d: %d %d %s\n", i, obj.objId, obj.objTyp, obj.objTypStr)
	}
	fmt.Printf("dbg: **** end objlist ******\n")
}
