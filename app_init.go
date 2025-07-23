package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/aarzilli/nucular"
)

const (
	MATRIX_LOG_PREFIX   = "[MATRIX]"
	MATRIX_LOGGER_FLAGS = log.Ltime | log.Lmsgprefix
	MATRIX_DATA_FOLDER  = ".matrix"
	MATRIX_LOGS_FOLDER  = "logs"
	MATRIX_CONFIG_FILE  = "config.toml"
	MATRIX_GUI_SCALE    = 2.0

	MATRIX_POPUP_FLAGS = nucular.WindowScalable | nucular.WindowMovable | nucular.WindowClosable
)

type initResult struct {
	logger *log.Logger
	config matrixConfig
}

func initAppResources() (result *initResult, err error) {
	result = &initResult{}

	wasSuccessful := false

	defer func() {
		if result.logger != nil && wasSuccessful {
		}
	}()

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	matrixDataPath := filepath.Join(home, MATRIX_DATA_FOLDER)

	matrixLogsPath := filepath.Join(matrixDataPath, MATRIX_LOGS_FOLDER)

	err = os.MkdirAll(matrixLogsPath, 0o777|os.ModeDir)
	if err != nil {
		return result, err
	}

	curTime := time.Now().Format(time.DateTime)

	workingLogPath := filepath.Join(matrixLogsPath, strings.ReplaceAll(fmt.Sprintf("%s.log", strings.ReplaceAll(curTime, " ", "_")), ":", "-"))

	file, err := os.Create(workingLogPath)
	if err != nil {
		return result, err
	}

	logFileHandle = file

	logger := log.New(logFileHandle, MATRIX_LOG_PREFIX+" ", MATRIX_LOGGER_FLAGS)
	result.logger = logger

	logger.Println("Beginning of application resource initialization")
	logger.Println("Matrix data folder created at path:", matrixDataPath)
	logger.Println("Matrix logs folder created at path:", matrixLogsPath)
	logger.Println("Logger linked to file at path:", workingLogPath)

	matrixConfigPath := filepath.Join(matrixDataPath, MATRIX_CONFIG_FILE)

	if _, err = os.Stat(matrixConfigPath); errors.Is(err, os.ErrNotExist) {
		logger.Println("Config file not found, creating...")

		var data []byte
		data, err = toml.Marshal(&defaultMatrixConfig)
		if err != nil {
			return result, err
		}

		logger.Println("")

		err = os.WriteFile(matrixConfigPath, data, 0o666)
		if err != nil {
			return
		}

		logger.Println("Config file created at path:", matrixConfigPath)
	} else if err != nil {
		return
	}

	logger.Println("Loading config from path:", matrixConfigPath)

	_, err = toml.DecodeFile(matrixConfigPath, &result.config)
	if err != nil {
		return
	}

	logger.Println("Config loaded")

	wasSuccessful = true

	logger.Println("Application resources initialized")

	return result, nil
}
