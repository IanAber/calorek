package main

import (
	"database/sql"
	"encoding/binary"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"go.bug.st/serial"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"
)

var (
	WebPort          string
	databaseServer   string
	databasePort     string
	databaseName     string
	databaseLogin    string
	databasePassword string
	webFiles         string
	pDB              *sql.DB
	Params           ParamsType
	logStatement     *sql.Stmt
	s                serial.Port
)

func ToTemperature(buffer []byte) float64 {
	return (float64(binary.BigEndian.Uint16(buffer)) / 16.0) - 55.0
}

func connectToDatabase() (*sql.DB, *sql.Stmt, error) {
	if pDB != nil {
		_ = pDB.Close()
		pDB = nil
	}
	// Set the time zone to Local to correctly record times
	var sConnectionString = databaseLogin + ":" + databasePassword + "@tcp(" + databaseServer + ":" + databasePort + ")/" + databaseName + "?loc=Local"

	db, err := sql.Open("mysql", sConnectionString)
	if err != nil {
		return nil, nil, err
	}
	err = db.Ping()
	if err != nil {
		_ = db.Close()
		pDB = nil
		return nil, nil, err
	}
	logStatement, err := db.Prepare("INSERT INTO heatpump.values(dischargePressure, suctionPressure, sourceInTemp, sourceOutTemp, loadInTemp, loadOutTemp, compressorSpeed, errorFlags, demandStatus) VALUES  (?,?,?,?,?,?,?,?,?)")
	return db, logStatement, err
}

func init() {
	flag.StringVar(&WebPort, "WebPort", "28080", "Web port")
	flag.StringVar(&webFiles, "webFiles", "/calorek/web", "Path to the WEB files location")
	flag.StringVar(&databaseServer, "sqlServer", "localhost", "MySQL Server")
	flag.StringVar(&databaseName, "database", "heatpump", "Database name")
	flag.StringVar(&databaseLogin, "dbUser", "logger", "Database login user name")
	flag.StringVar(&databasePassword, "dbPassword", "logger", "Database user password")
	flag.StringVar(&databasePort, "dbPort", "3306", "Database port")
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	log.Println("Starting the WEB site.")
	if pdb, logStmt, err := connectToDatabase(); err != nil {
		log.Fatal(err)
	} else {
		logStatement = logStmt
		pDB = pdb
	}
	log.Println("Starting the serial port")
	if s = ConnectSerial(); s == nil {
		log.Println("Failed to strat, trying again...")
		s = RestartSerial()
		if s == nil {
			log.Fatal("Failed to open the serial port on startup.")
		}
	}
	log.Println("Serial port is running")
	go setUpWebSite()
}

func ConnectSerial() serial.Port {
	mode := &serial.Mode{
		BaudRate: 19200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open("/dev/ttyUSB0", mode)
	if err != nil {
		log.Println(err)
		return nil
	}
	if err := port.SetReadTimeout(time.Millisecond * 150); err != nil {
		log.Println(err)
	}
	log.Println("Serial port connected")
	return port
}

func RestartSerial() serial.Port {
	log.Println("Restarting serial port")
	if s != nil {
		log.Println("Closing ttyUSB2")
		if err := s.Close(); err != nil {
			log.Println(err)
		}
		s = nil
	}
	log.Println("Port closed")
	time.Sleep(time.Second * 2)
	log.Println("Running usbreset command")
	// Reset the USB device by using the usbreset command line program
	cmd := exec.Command("usbreset", "0403:6001")
	stdout, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return nil
	} else {
		log.Println(string(stdout))
	}
	log.Println("Reset complete, connecting")
	time.Sleep(time.Second * 2)
	return ConnectSerial()
}

func main() {
	log.Println("Starting main.")
	dataSignal = sync.NewCond(&sync.Mutex{})

	paramBuf := make([]byte, 125)
	idx := 0
	timeout := 0
	for {
		if s == nil {
			s = RestartSerial()
			if s == nil {
				log.Println("Failed to connect")
			}
		}
		if s != nil {
			buf := make([]byte, 1)
			n, err := s.Read(buf)
			timeout++
			if err != nil && err != io.EOF {
				s = RestartSerial()
				timeout = 0
			}
			if n == 0 {
				idx = 0
			} else {
				paramBuf[idx] = buf[0]
				idx++
				timeout = 0
				if idx >= len(paramBuf) {
					Params.setValues(paramBuf)
					dataSignal.Broadcast() // Signal to broadcast values to registered web socket clients
					if pDB == nil {
						if pdb, logStmt, err := connectToDatabase(); err != nil {
							log.Println(err)
						} else {
							pDB = pdb
							logStatement = logStmt
						}
					}
					if logStatement != nil {
						if _, err := logStatement.Exec(Params.DischargePressure, Params.SuctionPressure, Params.SourceInTemp, Params.SourceOutTemp, Params.LoadTempIn, Params.LoadTempOut, Params.CompressorSpeed, Params.getErrorFlags(), Params.getDemandStatus()); err != nil {
							log.Println("Database error - ", err)
							if err := pDB.Close(); err != nil {
								log.Println(err)
							}
							logStatement = nil
							pDB = nil
						}
					}
					idx = 0
				}
			}
			if timeout > 10 {
				log.Println("Timed out...")
				s = RestartSerial()
				timeout = 0
			}
		}
	}
}
