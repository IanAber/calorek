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
	pDB          *sql.DB
	logStatement *sql.Stmt
	s            serial.Port
	Settings     SettingsType
	Params       ParamsType
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
	var sConnectionString = Settings.DatabaseLogin + ":" + Settings.DatabasePassword + "@tcp(" + Settings.DatabaseServer + ":" + Settings.DatabasePort + ")/" + Settings.DatabaseName + "?loc=Local"

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
	logStatement, err := db.Prepare(`INSERT INTO heatpump.values(dischargePressure,
		suctionPressure, sourceInTemp, sourceOutTemp, loadInTemp, loadOutTemp, 
		compressorSpeed, errorFlags, demandStatus, eev_pos, demand) VALUES  (?,?,?,?,?,?,?,?,?,?,?)`)
	return db, logStatement, err
}

func init() {
	flag.IntVar(&Settings.WebPort, "WebPort", 28080, "Web port")
	flag.IntVar(&Settings.LocalPort, "LocalPort", 8090, "Local port")
	flag.StringVar(&Settings.WebFiles, "webFiles", "/calorek/web", "Path to the WEB files location")
	flag.StringVar(&Settings.DatabaseServer, "sqlServer", "localhost", "MySQL Server")
	flag.StringVar(&Settings.DatabaseName, "database", "heatpump", "Database name")
	flag.StringVar(&Settings.DatabaseLogin, "dbUser", "logger", "Database login user name")
	flag.StringVar(&Settings.DatabasePassword, "dbPassword", "logger", "Database user password")
	flag.StringVar(&Settings.DatabasePort, "dbPort", "3306", "Database port")
	flag.StringVar(&Settings.SSLCertificateFile, "SSLCert", "/certs/fullchain.cer", "Path to the SSL Certificate file")
	flag.StringVar(&Settings.SSLPrivateKeyFile, "SSLPrivateKey", "/certs/elektrik.green.key", "Path to the SSL Private Key file")
	flag.StringVar(&Settings.SerialPort, "SerialPort", "/dev/ttyUSB0", "Port that the heat pump is connected to. You should use a by-path definition.")
	flag.Parse()
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
	port, err := serial.Open(Settings.SerialPort, mode)
	if err != nil {
		log.Println(err)
		return nil
	}
	if err := port.SetReadTimeout(time.Millisecond * 150); err != nil {
		log.Println(err)
	}
	log.Printf("Serial port connected on %s", Settings.SerialPort)
	return port
}

func RestartSerial() serial.Port {
	log.Println("Restarting serial port")
	if s != nil {
		log.Println("Closing ttyUSB")
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

	//	paramBuf := make([]byte, 125)
	paramBuf := make([]byte, 107)
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
				//if idx != 0 {
				//	log.Printf("%d bytes thrown away. Waiting for %d", idx, len(paramBuf))
				//}
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
						if _, err := logStatement.Exec(Params.DischargePressure,
							Params.SuctionPressure,
							Params.SourceInTemp,
							Params.SourceOutTemp,
							Params.LoadTempIn,
							Params.LoadTempOut,
							Params.CompressorSpeed,
							Params.getErrorFlags(),
							Params.getDemandStatus(),
							Params.EEVRequestedPosition,
							0); err != nil {

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
