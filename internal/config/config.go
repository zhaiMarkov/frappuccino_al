package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	APIPort    string
}

// Прочитать файл env и проверить данные для будущей подключения а так же работы базы данных
func LoadConfig() (*Config, error) {
	envMap, err := ParseEnvFile(".env")
	if err != nil {
		return nil, err
	}

	dbHost, exist := envMap["DB_HOST"]
	if !exist {
		return nil, fmt.Errorf("the DB_HOST value is not set in the environment variables")
	}

	dbPort, exist := envMap["DB_PORT"]
	if !exist {
		return nil, fmt.Errorf("the DB_PORT value is not set in the environment variables")
	}

	dbUser, exist := envMap["DB_USER"]
	if !exist {
		return nil, fmt.Errorf("the DB_USER value is not set in the environment variables")
	}

	dbPassword, exist := envMap["DB_PASSWORD"]
	if !exist {
		return nil, fmt.Errorf("the DB_PASSWORD value is not set in the environment variables")
	}

	dbName, exist := envMap["DB_NAME"]
	if !exist {
		return nil, fmt.Errorf("the DB_NAME value is not set in the environment variables")
	}

	apiPort, exist := envMap["API_PORT"]
	if !exist {
		return nil, fmt.Errorf("the API_PORT value is not set in the environment variables")
	}

	return &Config{
		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,
		APIPort:    apiPort,
	}, nil
}

// Прочитать файл и запарсить данные с него
func ParseEnvFile(filename string) (map[string]string, error) {
	envMap := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		slog.Error("Error file opening", "file name", filename, "error", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("wrong format line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		envMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envMap, nil
}
