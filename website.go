package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"time"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		// status -1 doesn't overwrite existing status code
		ctx.String(-1, ctx.Errors.String())
	}
}

func setUpWebSite() {
	errs := make(chan error)
	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Println(err)
	}

	//	router := mux.NewRouter().StrictSlash(true)
	// Register with the WebSocket which will then push a JSON payload with data to keep the displayed data up to date. No polling necessary.
	router.LoadHTMLGlob(Settings.WebFiles + "/templates/*")
	router.Use(ErrorHandler())

	router.Static("/css", Settings.WebFiles+"/css/")
	router.Static("/images", Settings.WebFiles+"/images/")
	router.Static("/scripts", Settings.WebFiles+"/scripts/")
	router.StaticFile("/favicon.ico", Settings.WebFiles+"/images/favicon.ico")
	router.GET("/ws", startDataWebSocket)
	router.GET("/status", getStatus)
	router.GET("/", defaultPage)
	router.GET("/historyData", getHistoryData)
	router.GET("/chart.html", getChart)
	router.PATCH("/toggleCoil", toggleCoil)

	log.Printf("Starting secure site on port %d", Settings.WebPort)
	go startSecureSite(router, errs)
	log.Printf("Starting http site on port %d", Settings.LocalPort)
	go startInsecureSite(router, errs)

	select {
	case err := <-errs:
		log.Println("Web service failed to sstart - %v", err)
	}
}

func getChart(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "chart.html", nil)
}
func startInsecureSite(router *gin.Engine, errs chan error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", GetLocalIP(), Settings.LocalPort))
	if err != nil {
		log.Fatal(err)
	}
	errs <- http.Serve(l, router)
}

func startSecureSite(router *gin.Engine, errs chan error) {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", Settings.WebPort))
	if err != nil {
		log.Fatal(err)
	}
	errs <- http.ServeTLS(l, router, Settings.SSLCertificateFile, Settings.SSLPrivateKeyFile)
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func defaultPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "default.html", nil)
}

func getStatus(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &Params)
}

func getHistoryData(ctx *gin.Context) {
	const DeviceString = "getHistoryData"

	if pDB == nil {
		ReturnJSONErrorString(ctx, DeviceString, "No Database", http.StatusInternalServerError, true)
		return
	}

	if start, end, err := GetTimeRange(ctx); err != nil {
		ReturnJSONError(ctx, DeviceString, err, http.StatusBadRequest, false)
	} else {
		if end.Sub(start) > time.Hour {
			SendDataAsJSON(ctx, DeviceString, DataByMinute, start, end)
		} else {
			SendDataAsJSON(ctx, DeviceString, DataBySecond, start, end)
		}
	}
}

func toggleCoil(ctx *gin.Context) {
	client := &http.Client{}
	address := ctx.PostForm("coil")
	url := fmt.Sprintf("https://firefly.home:20080/toggleCoil?coil=" + address)

	req, err := http.NewRequest(http.MethodPatch, url, nil) //bytes.NewBuffer(payload))
	if err != nil {
		log.Println(err)
	}
	if _, err := client.Do(req); err != nil {
		log.Println(err)
	}
}
