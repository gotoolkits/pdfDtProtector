package pdf

import (
	"errors"
	"os"
	"strconv"

	"github.com/gotoolkits/pdfDtProtector/index"
	pdfreader "github.com/ledongthuc/pdf"
	pdfwriter "github.com/signintech/gopdf"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gographics/imagick.v3/imagick"
)

type Image struct {
	FilePath string
	Width    uint
	Height   uint
}

type MaskBlock struct {
	Cols uint
	Rows uint
}

type MaskPix struct {
	StartPixs PixPoint
	Len       float64
}

type DataProtector interface {
	ReadPdfContent(f *os.File, r *pdfreader.Reader) (string, error)
	ConvertPdfToImages(filePath string, resolution float64, compressionQuality uint) (imgs []Image, err error)
	DataProtect(imges []Image, msk MaskBlock, pixs []PixPoint) (maskImgs []Image, err error)
	WritePdf(imgesPath []string) (err error)
	Destory()
}

type PdfDtProtector struct {
	datas     index.ContentCacheTable
	idxTable  index.CharacterIndexTable
	f         *os.File
	r         *pdfreader.Reader
	PWidth    uint
	PHeight   uint
	PfontSize float64
}

func init() {
	imagick.Initialize()
}

func NewPdfDtProtector() PdfDtProtector {
	return PdfDtProtector{datas: index.NewContentCacheable(),
		idxTable: index.NewCharacterIndexTable()}
}

func (pdp *PdfDtProtector) Destory() {
	if pdp.f != nil {
		pdp.f.Close()
	}
	imagick.Terminate()
}

func (pdp *PdfDtProtector) ConvertPdfToImages(resolution float64, compressionQuality uint) (imgs []Image, err error) {
	if pdp.f == nil {
		return nil, errors.New("Not found openning pdf file.")
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.SetResolution(resolution, resolution); err != nil {
		log.Error("Setting resolution failed!")
		return nil, err
	}

	if err := mw.ReadImageFile(pdp.f); err != nil {
		log.Errorf("Read image file failed: %s", err.Error())
		return nil, err
	}

	pages := int(mw.GetNumberImages())
	path := ""

	if pages < 1 {
		log.Errorf("Read PDF page is 0.")
		return nil, errors.New("Read PDF page is 0.")
	}

	pWidth := mw.GetImageWidth()
	pHeight := mw.GetImageHeight()

	pdp.PHeight = pHeight
	pdp.PWidth = pWidth

	for i := 0; i < pages; i++ {
		mw.SetIteratorIndex(i) // This being the page offset

		mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_REMOVE)
		mw.MergeImageLayers(imagick.IMAGE_LAYER_FLATTEN)

		mw.SetImageFormat("jpg")
		mw.SetImageCompression(imagick.COMPRESSION_JPEG)
		mw.SetImageCompressionQuality(compressionQuality)

		mw.ThumbnailImage(pWidth, pHeight)

		path = "./" + pdp.f.Name() + "_" + strconv.Itoa(i+1) + "page" + ".jpg"
		err = mw.WriteImage(path)
		if err != nil {
			return nil, err
		}

		jpg := Image{FilePath: path,
			Width:  pWidth,
			Height: pHeight,
		}

		imgs = append(imgs, jpg)
	}

	return
}

func (pdp *PdfDtProtector) DataProtect(imges []Image, maskRows uint, pixsLocation []Location) (maskImgs []Image, err error) {
	if len(imges) < 1 {
		return
	}

	mskPixs := pixlocToPixbyStart(pixsLocation)
	log.Info("Pixlococation To Pix by startPt.")
	for i, img := range imges {
		f, err := os.Open(img.FilePath)
		if err != nil {
			f.Close()
			return nil, err
		}

		page := i + 1
		maskimg, err := drawMosaic(page, f, maskRows, mskPixs)
		log.Info("Drawing Mosaic Page: ", page)
		if err != nil {
			return nil, err
		}
		maskImgs = append(maskImgs, maskimg)
	}

	return
}

func pixlocToPixbyStart(pixsLocation []Location) (mskPixs map[int][]MaskPix) {
	mskPixs = make(map[int][]MaskPix)

	for _, loc := range pixsLocation {
		if _, ok := mskPixs[loc.Page]; !ok {
			mskPixs[loc.Page] = []MaskPix{}
		}

		mskPixs[loc.Page] = append(mskPixs[loc.Page],
			MaskPix{StartPixs: loc.Start,
				Len: loc.Length})
	}
	return
}

func drawMosaic(page int, img *os.File, maskRows uint, mskPixs map[int][]MaskPix) (maskImg Image, err error) {
	if len(mskPixs) < 1 {
		return Image{}, errors.New("Pixs args can't NULL.")
	}

	pw := imagick.NewPixelWand()
	pw.SetColor("black")

	dw := imagick.NewDrawingWand()
	dw.SetFillColor(pw)

	mw := imagick.NewMagickWand()

	lw := imagick.NewMagickWand()
	log.Info("read strating.")
	if err = lw.ReadImageFile(img); err != nil {
		return
	}
	log.Info("read finish.")

	log.Info("Protecting page: ", page)

	for _, mskPix := range mskPixs[page] {

		if err = mw.NewImage(uint(mskPix.Len), maskRows, pw); err != nil {
			return
		}

		if err = mw.DrawImage(dw); err != nil {
			return
		}

		if err = lw.CompositeImage(mw, imagick.COMPOSITE_OP_SRC_IN, false, int(mskPix.StartPixs.X), int(mskPix.StartPixs.Y)); err != nil {
			return
		}
	}

	imgName := img.Name() + "_" + "mask.jpg"
	log.Info("write image strating.")
	if err = lw.WriteImage(imgName); err != nil {
		return
	}
	log.Info("write image finished.")

	pWidth := lw.GetImageWidth()
	pHeight := lw.GetImageHeight()
	maskImg = Image{FilePath: imgName,
		Width:  pWidth,
		Height: pHeight,
	}
	return
}

func (pdp *PdfDtProtector) WritePdf(imges []Image) (err error) {
	length := len(imges)
	if length < 1 {
		return
	}

	pdfw := pdfwriter.GoPdf{}
	rect := pdfwriter.Rect{W: 595.28, H: 841.89} //595.28, 841.89 = A4

	pdfw.Start(pdfwriter.Config{PageSize: rect})

	for _, img := range imges {
		pdfw.AddPage()
		pdfw.Image(img.FilePath, 0, 0, &rect) //print image

	}

	pdfw.WritePdf(imges[0].FilePath + ".pdf")
	return
}
