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
		FlagLogLevel:             "info",
	}
	parseFlagsServer(&conf)

	return &conf
}

func parseFlagsServer(c *ConfigServer) {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.FlagDatabaseDSN, "d", "postgres://user:password@localhost:5464/go-ya-gophermart", "DATABASE_DSN string")
	flag.StringVar(&c.FlagAccrualSystemAddress, "r", "http://localhost:8081/", "ACCRUAL_SYSTEM_ADDRESS string")
	//flag.StringVar(&c.FlagLogLevel, "l", "info", "log level")

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		c.FlagRunAddr = envRunAddr
	}
	if envDatabaseDSN := os.Getenv("DATABASE_URI"); envDatabaseDSN != "" {
		c.FlagDatabaseDSN = envDatabaseDSN
	}
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		c.FlagAccrualSystemAddress = envAccrualSystemAddress + "/"
	}
	//if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
	//	c.FlagLogLevel = envLogLevel
	//}

}
