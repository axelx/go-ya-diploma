package config

import (
	"flag"
	"fmt"
	"os"
)

type ConfigServer struct {
	FlagRunAddr              string
	FlagDatabaseDSN          string
	FlagAccrualSystemAddress string
	FlagLogLevel             string
}

func (c *ConfigServer) String() string {
	return fmt.Sprintf("FlagRunAddr: %s, FlagDatabaseDSN: %v, FlagAccrualSystemAddress: %s, FlagLogLevel: %s",
		c.FlagRunAddr, c.FlagDatabaseDSN, c.FlagAccrualSystemAddress, c.FlagLogLevel)
}

func NewConfigServer() *ConfigServer {
	conf := ConfigServer{
		FlagRunAddr:              "",
		FlagDatabaseDSN:          "",
		FlagAccrualSystemAddress: "",
		FlagLogLevel:             "",
	}
	parseFlagsServer(&conf)

	return &conf
}

func parseFlagsServer(c *ConfigServer) {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.FlagDatabaseDSN, "d", "", "DATABASE_DSN string")
	flag.StringVar(&c.FlagAccrualSystemAddress, "r", "localhost:8081", "ACCRUAL_SYSTEM_ADDRESS string")
	flag.StringVar(&c.FlagLogLevel, "l", "info", "log level")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.FlagRunAddr = envRunAddr
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		c.FlagDatabaseDSN = envDatabaseDSN
	}
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		c.FlagLogLevel = envAccrualSystemAddress
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.FlagLogLevel = envLogLevel
	}

}
