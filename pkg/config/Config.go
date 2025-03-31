package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

/*
Configuration Object that holds all the necessary config information
*/
type Configuration struct {
	postgresHost     string
	postgresPort     string
	postgresDatabase string
	postgresUser     string
	postgresPassword string

	relayAddress string
	relayPort    string
	relayName    string

	RelayWriteWait      time.Duration
	RelayPongWait       time.Duration
	RelayPingWait       time.Duration
	RelayMaxMessageSize int64

	NIP11Software    string
	NIP11Description string
	NIP11Version     string
	NIP11Pubkey      string
	NIP11Contact     string
}

/*
Init initializes the config parameters
*/
func (conf *Configuration) Init(envFile string) (*Configuration, error) {
	err := godotenv.Load(envFile)
	if err != nil {
		return conf, err
	}

	conf.postgresHost = os.Getenv("POSTGRES_HOST")
	conf.postgresPort = os.Getenv("POSTGRES_PORT")
	conf.postgresDatabase = os.Getenv("POSTGRES_DB")
	conf.postgresUser = os.Getenv("POSTGRES_USER")
	conf.postgresPassword = os.Getenv("POSTGRES_PASSWORD")

	conf.relayAddress = os.Getenv("RELAY_ADDRESS")
	conf.relayPort = os.Getenv("RELAY_PORT")

	conf.RelayWriteWait = 10 * time.Second
	conf.RelayPongWait = 60 * time.Second
	conf.RelayPingWait = conf.RelayPongWait / 2
	conf.RelayMaxMessageSize, _ = strconv.ParseInt(os.Getenv("RELAY_MAX_MESSAGE_SIZE"), 10, 64)

	conf.NIP11Software = os.Getenv("NIP11_SOFTWARE")
	conf.NIP11Description = os.Getenv("NIP11_DESCRIPTION")
	conf.NIP11Version = os.Getenv("NIP11_VERSION")
	conf.NIP11Contact = os.Getenv("NIP11_CONTACT")
	conf.NIP11Pubkey = os.Getenv("NIP11_PUBKEY")

	return conf, nil
}

/*
RelayName returns the name of the relay from the env file
*/
func (conf *Configuration) RelayName() string {
	return conf.relayName
}

/*
GetRelayAddress returns the assembled address of the server in form adddress:port
*/
func (conf *Configuration) GetRelayAddress() string {
	return fmt.Sprintf("%v:%v", conf.relayAddress, conf.relayPort)
}

/*
GetDatabaseConnectionString Construct the database connection string
*/
func (conf *Configuration) GetDatabaseConnectionString() string {
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=UTC",
		conf.postgresHost,
		conf.postgresUser,
		conf.postgresPassword,
		conf.postgresDatabase,
		conf.postgresPort)
}

/*
Config Export the Config object
*/
var Config, _ = (&(Configuration{})).Init(".env")
