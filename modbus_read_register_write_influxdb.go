package main

import (
	"fmt"
	"time"

	tm "github.com/buger/goterm"
	"github.com/goburrow/modbus"
	"github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
)

const (
	tsdb       = "powermeterdb"
	tsuser     = "poweruser"
	tspassword = "P@ssw0rd"
)

var normal chan bool

type EnergyMeter struct {
	MODEL string  `json:"model,omitempty"`
	SN    string  `json:"sn,omitempty"`
	Freq  float64 `json:"freq,omitempty"`
	Uavg  float64 `json:"uavg,omitempty"`
	Iavg  float64 `json:"iavg,omitempty"`
	Psum  int32   `json:"psum,omitempty"`
	Qsum  int32   `json:"qsum,omitempty"`
	Ssum  int32   `json:"ssum,omitempty"`
	PFavg float64 `json:"pfavg,omitempty"`
	EA    float64 `json:"ea,omitempty"`
	ER    float64 `json:"er,omitempty"`
	CO2   float64 `json:"co2,omitempty"`
}

func main() {
	BKK, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Error(err)
		return
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: tsuser,
		Password: tspassword,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	normal = make(chan bool, 1)
	SlaveId := byte(1)
	addr := uint16(0x59)
	// handler := modbus.NewTCPClientHandler("127.0.0.1:502")
	handler := modbus.NewRTUClientHandler("/dev/ttyUSB0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 2
	handler.SlaveId = SlaveId
	handler.Timeout = 1 * time.Second

	err = handler.Connect()
	if err != nil {
		fmt.Println(err)
	}
	defer handler.Close()

	client := modbus.NewClient(handler)
	for {
		var output string
		results, err := client.ReadHoldingRegisters(addr, 6)
		if err != nil {
			fmt.Println(err)
			break
		}
		year := (int32(results[0])<<8 | int32(results[1]))
		month := (int32(results[2])<<8 | int32(results[3]))
		date := (int32(results[4])<<8 | int32(results[5]))
		hour := (int32(results[6])<<8 | int32(results[7]))
		minute := (int32(results[8])<<8 | int32(results[9]))
		second := (int32(results[10])<<8 | int32(results[11]))
		output = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", year, month, date, hour, minute, second)
		dtString := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", year, month, date, hour, minute, second)

		results, err = client.ReadHoldingRegisters(0x01, 1)
		if err != nil {
			log.Println(err)
			break
		}
		freq := (int32(results[0])<<8 | int32(results[1]))
		output += fmt.Sprintf("\n Freq : %0.2f Hz", float64(freq)/100)

		results, err = client.ReadHoldingRegisters(0x02, 2)
		if err != nil {
			log.Println(err)
			break
		}
		avgU := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n U avg : %0.1f V", float64(avgU)/10)

		results, err = client.ReadHoldingRegisters(0x04, 2)
		if err != nil {
			log.Println(err)
			break
		}
		avgLU := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n UL avg : %0.1f V", float64(avgLU)/10)

		results, err = client.ReadHoldingRegisters(0x06, 2)
		if err != nil {
			log.Println(err)
			break
		}
		avgI := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n I avg : %0.3f A", float64(avgI)/1000)

		results, err = client.ReadHoldingRegisters(0x08, 2)
		if err != nil {
			log.Println(err)
			break
		}
		In := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n In : %0.3f A", float64(In)/1000)

		results, err = client.ReadHoldingRegisters(0x0A, 2)
		if err != nil {
			log.Println(err)
			break
		}
		Psum := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n Psum : %d W", Psum)

		results, err = client.ReadHoldingRegisters(0x0C, 2)
		if err != nil {
			log.Println(err)
			break
		}
		Qsum := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n Qsum : %d VAR", Qsum)

		results, err = client.ReadHoldingRegisters(0x0C, 2)
		if err != nil {
			log.Println(err)
			break
		}
		Ssum := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n Ssum : %d VA", Ssum)

		results, err = client.ReadHoldingRegisters(0x10, 2)
		if err != nil {
			log.Println(err)
			break
		}
		PFavg := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n PF avg : %.3f VA", float64(PFavg)/1000)

		results, err = client.ReadHoldingRegisters(0x12, 2)
		if err != nil {
			log.Println(err)
			break
		}
		ea := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n Ea : %0.1f kWh", float64(ea)/10)

		results, err = client.ReadHoldingRegisters(0x14, 2)
		if err != nil {
			log.Println(err)
			break
		}
		er := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n Er : %0.1f kVARh", float64(er)/10)

		// results, err = client.ReadHoldingRegisters(0x16, 2)
		// if err != nil {
		// 	log.Println(err)
		// 	break
		// }
		// cost := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		// log.Printf("Cost : %0.1f ฿", float64(cost)/10)

		results, err = client.ReadHoldingRegisters(0x18, 2)
		if err != nil {
			log.Println(err)
			break
		}
		co2 := (int32(results[0])<<24 | int32(results[1])<<16 | int32(results[2])<<8 | int32(results[3]))
		output += fmt.Sprintf("\n CO2 : %0.1f kg", float64(co2)/10)

		datetime, err := time.ParseInLocation(dtString, "2006-01-02 15:04:05", BKK)
		if err != nil {
			datetime = time.Now()
		}
		data := EnergyMeter{
			MODEL: "CPM-20-A5V6-ADH",
			SN:    "1410170013-3500",
			Freq:  float64(freq) / 100,
			Uavg:  float64(avgU) / 10,
			Iavg:  float64(avgI) / 1000,
			Psum:  Psum,
			Qsum:  Qsum,
			Ssum:  Ssum,
			PFavg: float64(PFavg) / 1000,
			EA:    float64(ea) / 10,
			ER:    float64(er) / 10,
			CO2:   float64(co2) / 10,
		}
		writePoints(c, datetime, data)

		tm.Clear() // Clear current screen
		tm.MoveCursor(1, 1)
		tm.Print(output)
		tm.Flush() // Call it every time at the end of rendering

		// time.Sleep(time.Second)
	}
}

func writePoints(c client.Client, dt time.Time, data EnergyMeter) {

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  tsdb,
		Precision: "s",
	})
	if err != nil {
		log.Error(err)
	}

	fields := map[string]interface{}{
		"Freq":  data.Freq,
		"Uavg":  data.Uavg,
		"Iavg":  data.Iavg,
		"Psum":  data.Psum,
		"Qsum":  data.Qsum,
		"PFavg": data.PFavg,
		"EA":    data.EA,
		"ER":    data.ER,
		"CO2":   data.CO2,
	}
	tags := map[string]string{
		"Model": data.MODEL,
		"SN":    data.SN,
	}
	pt, err := client.NewPoint(
		"energylog",
		tags,
		fields,
		dt,
	)
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}
}
