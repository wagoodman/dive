package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

func configureCi() (bool, *viper.Viper, error) {
	isCiFromEnv, _ := strconv.ParseBool(os.Getenv("CI"))
	isCi = isCi || isCiFromEnv

	if isCi {
		ciConfig.SetConfigType("yaml")

		if _, err := os.Stat(ciConfigFile); !os.IsNotExist(err) {
			fmt.Printf("  Using CI config: %s\n", ciConfigFile)

			fileBytes, err := os.ReadFile(ciConfigFile)
			if err != nil {
				return isCi, nil, err
			}

			err = ciConfig.ReadConfig(bytes.NewBuffer(fileBytes))
			if err != nil {
				return isCi, nil, err
			}
		} else {
			fmt.Println("  Using default CI config")
		}
	}

	return isCi, ciConfig, nil
}
