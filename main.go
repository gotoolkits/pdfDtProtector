package main

import (
	"github.com/gotoolkits/pdfDtProtector/config"
	"github.com/gotoolkits/pdfDtProtector/pdf"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	var filename string

	if len(os.Args) > 1 {
		filename = os.Args[1]
	} else {
		log.Fatal("please input pdf file name and path.")
	}

	if err := config.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	cfg := config.GlobleConfigs

	pdp := pdf.NewPdfDtProtector()
	defer pdp.Destory()

	if err := pdp.PdfOpen(filename); err != nil {
		log.Fatal("pdf file open failed!", err)
	}

	ruleMatchStr := []pdf.SensiteiveData{}
	for i, regx := range cfg.RuleCfg.RegxRule {
		ok, match, err := pdp.IsIncludeSensiteiveData(regx)
		if err != nil {
			log.Errorln(err)
		}

		if ok {
			ruleMatchStr = append(ruleMatchStr, match...)
		}

		log.Infoln("Rule ", i, " matched string ", len(match))
	}

	if len(ruleMatchStr) < 1 {
		log.Fatal("No Found Sensiteive Data in pdf.")
	}

	PtMap, _ := pdp.FindSensitiveData(ruleMatchStr)
	imgs, _ := pdp.ConvertPdfToImages(cfg.SettingCfg.ImagePpi, cfg.SettingCfg.CompressionQuality)

	pixList := pdf.TransImagePtToPix(PtMap, pdp.PHeight, cfg.SettingCfg.ImagePpi, pdp.PfontSize)
	log.Info("Transing image Pt to Pix...")

	mskRows := cfg.SettingCfg.MaskRows
	imgsWithMask := []pdf.Image{}
	var err error

	if imgsWithMask, err = pdp.DataProtect(imgs, mskRows, pixList); err != nil {
		log.Error(err)
	}
	log.Info("Data Protected: ", len(pixList))

	if err = pdp.WritePdf(imgsWithMask); err != nil {
		log.Fatalln(err)
	}

}
