package pdf

import (
	"container/list"
	"errors"
	"github.com/gotoolkits/pdfDtProtector/config"
	pdfreader "github.com/ledongthuc/pdf"
	log "github.com/sirupsen/logrus"
	"regexp"
)

type PixPoint struct {
	X float64
	Y float64
}

type Location struct {
	Start  PixPoint
	End    PixPoint
	Length float64
	Page   int
}

type Text struct {
	txt  pdfreader.Text
	Page int
}

type SensiteiveData struct {
	Datas []string
	Page  int
}

func (pdp *PdfDtProtector) FindSensitiveData(datas []SensiteiveData) (ppmap map[string][]Location, err error) {
	if err = pdp.ReadAndCachePdfContent(); err != nil {
		return
	}

	ppmap, err = pdp.searchIndexTableGetPixPoint(datas)
	if err != nil {
		log.Error(err)
		return
	}

	return
}

func (pdp *PdfDtProtector) IsIncludeSensiteiveData(regxExpr string) (include bool, matchs []SensiteiveData, err error) {
	pdf, err := pdp.GetPdfPlainText()
	if err != nil {
		return false, nil, err
	}

	for page, content := range pdf {
		rgx, _ := regexp.Compile(regxExpr)
		found := rgx.FindAllString(content, -1)

		if len(found) < 1 {
			return false, nil, nil
		}

		matchs = append(matchs, SensiteiveData{
			Datas: found,
			Page:  page,
		})

	}

	return true, matchs, nil
}

func (pdp *PdfDtProtector) searchIndexTableGetPixPoint(matchs []SensiteiveData) (map[string][]Location, error) {
	if len(matchs) < 1 {
		return nil, errors.New("match string is null!")
	}

	ppmap := make(map[string][]Location)

	for _, SensiteiveData := range matchs {
		for _, secureStr := range SensiteiveData.Datas {

			ptLocs, err := pdp.matchCacheByCharacter(secureStr)
			if err != nil {
				return nil, err
			}

			if len(ptLocs) < 1 {
				return nil, errors.New("Not Match search string in Cache tables!")
			}
			ppmap[secureStr] = ptLocs

		}

	}
	return ppmap, nil
}

func (pdp *PdfDtProtector) matchCacheByCharacter(match string) (ptLocs []Location, err error) {
	if len(pdp.idxTable.IndexTable) < 1 || pdp.datas.Cache.Len() < 1 {
		return ptLocs, errors.New("CacheTable and IndexTable is null!")
	}

	length := len(match)
	fristb := string(match[0])

	if elems, ok := pdp.idxTable.IndexTable[fristb]; ok {
		for _, el := range elems {
			txts := combinByFirestItem(length, el)

			if compareStrAndSlice(match, txts) {
				startPt := PixPoint{X: txts[0].txt.X, Y: txts[0].txt.Y}
				endPt := PixPoint{X: txts[length-1].txt.X, Y: txts[length-1].txt.Y}
				page := txts[0].Page

				ptl := Location{
					Start: startPt,
					End:   endPt,
					Page:  page,
				}
				ptLocs = append(ptLocs, ptl)
			}
		}
	}
	return
}

func TransImagePtToPix(ppmap map[string][]Location, imageHeight uint, ppi float64, fontSize float64) (pixLocs []Location) {
	if len(ppmap) < 1 {
		return
	}

	for _, ptLocs := range ppmap {
		for _, ptloc := range ptLocs {
			pixloc := Location{
				Start: PixPoint{
					X: getPixX(ptloc.Start.X, ppi),
					Y: getPixY(ptloc.Start.Y, imageHeight, ppi, fontSize),
				},

				End: PixPoint{
					X: getPixX(ptloc.End.X, ppi),
					Y: getPixY(ptloc.End.Y, imageHeight, ppi, fontSize),
				},
				Length: getPixX(ptloc.End.X, ppi) - getPixX(ptloc.Start.X, ppi) + calcPtToPx(fontSize, ppi),
				Page:   ptloc.Page,
			}
			pixLocs = append(pixLocs, pixloc)
		}
	}

	return
}

func combinByFirestItem(size int, first *list.Element) (cacheItems []Text) {
	elem := first
	for idx := 0; idx < size; idx++ {
		value := elem.Value.(Text)
		cacheItems = append(cacheItems, value)
		if elem.Next() == nil {
			return
		}
		elem = elem.Next()
	}

	return
}

func compareStrAndSlice(str string, cacheItems []Text) bool {
	if len(str) != len(cacheItems) {
		return false
	}

	var cacheStr = ""
	for _, citem := range cacheItems {
		cacheStr = cacheStr + citem.txt.S
	}
	return (str == cacheStr)
}

func (pdp *PdfDtProtector) ReadAndCachePdfContent() (err error) {
	if pdp.r == nil {
		return errors.New("pdf Reader error.")
	}

	totalPage := pdp.r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := pdp.r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		texts := p.Content().Text
		pdp.PfontSize = texts[0].FontSize
		for _, text := range texts {
			pTxt := Text{txt: text, Page: pageIndex}
			txtItem := pdp.datas.Cache.PushBack(pTxt)
			pdp.idxTable.ItemPush(text.S, txtItem)
		}
	}
	return
}

func (pdp *PdfDtProtector) PdfOpen(filePath string) (err error) {
	f, r, err := pdfreader.Open(filePath)
	//	defer f.Close()

	if err != nil {
		return
	}
	pdp.f = f
	pdp.r = r
	return
}

func (pdp *PdfDtProtector) GetPdfPlainText() ([]string, error) {
	var texts = []string{}

	if pdp.r == nil {
		return nil, errors.New("pdf Reader error.")
	}

	if pdp.f == nil {
		return nil, errors.New("Not found openning pdf file.")
	}

	totalPage := pdp.r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := pdp.r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		b, err := p.GetPlainText(nil)
		if err != nil {
			return nil, err
		}

		texts = append(texts, b)
	}

	return texts, nil
}

func getPixX(ptX float64, ppi float64) (pX float64) {
	fPix := calcPtToPx(ptX, ppi)
	return fPix
}

func getPixY(ptY float64, imageHeight uint, ppi float64, fontSize float64) (pY float64) {
	fPix := calcPtToPx(ptY, ppi)
	fz := calcPtToPx(fontSize, ppi)
	offset := config.GlobleConfigs.SettingCfg.Offset
	return float64(imageHeight) - fPix - fz/2.5 - float64(offset)
}

func calcPtToPx(pt float64, ppi float64) (pix float64) {
	return float64(pt / 72 * ppi)
}
