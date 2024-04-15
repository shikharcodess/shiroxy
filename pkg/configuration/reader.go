package configuration

import (
	"errors"
	"os"
	"shiroxy/pkg/loader"
	"shiroxy/pkg/models"

	"github.com/spf13/viper"
)

func ConfigReader(configUrl string) (*models.Config, error) {

	loaderController := loader.ProgressLoaderPayload{
		Close:          make(chan bool),
		Failed:         false,
		Content:        `-\|/`,
		MainContent:    "* Loading Configuration: ",
		SuccessMessage: "Success 🎉",
		FailedMessage:  "Failed 😔",
	}
	go loader.ProgressLoader(&loaderController)

	var config models.Config
	viper.SetConfigName("shiroxy")
	viper.SetConfigType("json")

	file, err := os.Open(configUrl)
	if err != nil {
		loaderController.CloseLoader(true, "", 1000)
		return nil, errors.New("Error opening file:" + err.Error())
	}
	defer file.Close()

	if err := viper.ReadConfig(file); err != nil {
		loaderController.CloseLoader(true, "", 1000)
		return nil, errors.New("Error reading config file:" + err.Error())
	}

	if err := viper.Unmarshal(&config); err != nil {
		loaderController.CloseLoader(true, "", 1000)
		return nil, errors.New("Unable to unmarshal into struct:" + err.Error())
	}

	loaderController.CloseLoader(false, "", 1000)
	return &config, nil
}
