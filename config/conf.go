package config

import (
	"errors"
	"github.com/spf13/viper"
)

type Configs struct {
	SettingCfg Settings `json:"settings"`
	RuleCfg    Rules    `json:"rules"`
}

type Settings struct {
	ImagePpi           float64 `json:"imagePPI"`
	CompressionQuality uint    `json:"compressionQuality"`
	MaskRows           uint    `json:"maskRows"`
	Offset             uint    `json:"offset"`
}

type Rules struct {
	RegxRule []string `json:"regxRule"`
}

var (
	GlobleConfigs Configs = Configs{}
)

func LoadConfig() (err error) {

	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/pdfDtProtector/")

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	GlobleConfigs.SettingCfg.ImagePpi = viper.GetFloat64("settings.imagePPI")
	GlobleConfigs.SettingCfg.CompressionQuality = uint(viper.GetInt("settings.compressionQuality"))
	GlobleConfigs.SettingCfg.MaskRows = uint(viper.GetInt("settings.maskRows"))
	GlobleConfigs.SettingCfg.Offset = uint(viper.GetInt("settings.offset"))
	GlobleConfigs.RuleCfg.RegxRule = viper.GetStringSlice("rules.regxRule")

	if !checkVaild() {
		return errors.New("Check config file and setting value failed !")
	}

	return

}

func checkVaild() bool {
	if GlobleConfigs.SettingCfg.ImagePpi < 0 ||
		GlobleConfigs.SettingCfg.CompressionQuality < 0 ||
		GlobleConfigs.SettingCfg.MaskRows < 0 ||
		len(GlobleConfigs.RuleCfg.RegxRule) < 1 {
		return false
	}
	return true
}
