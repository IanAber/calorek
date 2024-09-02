package main

import (
	"encoding/binary"
	"encoding/json"
	"sync"
)

const inputY1 = 0b00000001
const inputY2 = 0b00000010
const inputY3 = 0b00000100
const inputO = 0b00001000
const inputAntiFreeze = 0b00010000
const inputEEVOpen = 0b00100000
const inputWaterValveOpen = 0b01000000
const inputEEVSOpen = 0b10000000

type DemandStatusType struct {
	InputY1        bool
	InputY2        bool
	InputY3        bool
	InputO         bool
	Antifreeze     bool
	EEVOpen        bool
	WaterValveOpen bool
	EEVSOpen       bool
}

// ErrorMem0
const driveCommunicationsErrorFlag = 0b00000001
const evaporatorSuctionTempErrorFlag = 0b00000010
const suctionPressureErrorFlag = 0b00000100
const dischargePressureErrorFlag = 0b00001000
const sourceOutletTempErrorFlag = 0b00010000
const dischargeOverPressureErrorFlag = 0b01000000
const freezeCondition2ErrorFlag = 0b10000000

// ErrorMem1
const crititcalAlarmErrorFlag = 0b00000001
const loadOutletTempErrorFlag = 0b00000010
const sourceInletTempErrorFlag = 0b00000100
const loadInletTempErrorFlag = 0b00001000
const outdoorTempSetErrorFlag = 0b00010000
const freezeCondition1ErrorFlag = 0b00100000
const waterFlowErrorFlag = 0b01000000

// ErrorMem2
const driveInAlarmErrorFlag = 0b00000001

type ErrorFlagsType struct {
	DriveCommunications     bool
	EvaporatorSuctionTemp   bool
	SuctionPressure         bool
	DischargePressure       bool
	DischargeOverPressure   bool
	FreezeCondition2        bool
	CriticalAlarm           bool
	LoadInletTemperature    bool
	LoadOutletTemperature   bool
	SourceInletTemperature  bool
	SourceOutletTemperature bool
	OutdoorTemperature      bool
	FreezeCondition1        bool
	WaterFlow               bool
	DriveAlarm              bool
}

type ParamsType struct {
	DischargePressure            uint16
	SuctionPressure              uint16
	SourceOutTemp                float64
	SuctionEvaporatorTemperature float64
	SourceInTemp                 float64
	LoadTempOut                  float64
	OutdoorTemp                  float64
	EEVRequestedPosition         uint16
	LoadTempIn                   float64
	CompressorSpeed              uint16
	//	EEV2RequestedPosition        uint16
	//	LoadWaterValvePosition       uint8
	Drive1Alarms     uint8
	Drive1AlarmTimer uint16
	//	Drive2Alarms                 uint8
	//	Drive2AlarmTimer             uint8
	//	TimerAuxHeater               uint16
	DemandStatus DemandStatusType
	Errors       ErrorFlagsType
	PIDSetpoint  uint16
	//	Work3                        uint8
	//	DemandG1316                  uint8
	Compressor1Current uint16
	//	Compressor1Power             uint16
	Compressor1Power uint8
	//Y1MinimumOutputTemperature   uint16
	//Y1MaximumOutputTemperature   uint16
	//Y1MinimumWaterTemperature    uint16
	//Y1MaximumWaterTemperature    uint16
	//Y2MinimumOutputTemperature   uint16
	//Y2MaximumOutputTemperature   uint16
	//Y2MinimumWaterTemperature    uint16
	//Y2MaximumWaterTemperature    uint16
	//WaterSetpoint                uint16
	PID1Setpoint       uint16
	CriticalAlarmTimer uint16
	//	LowPressureTimer             uint16
	CompressorVoltage uint8
	CompressorCurrent uint8
	SourcePressure    uint16
	//LoadPressure                 uint16
	//PotSourceHP                  uint16
	//PotSourceLP                  uint16
	//PotLoadHP                    uint16
	//PotLoadLP                    uint16
	CompressorActualSpeed uint16
	//PT100_1_D24_D31              uint8
	//PT100_1_D16_D23              uint8
	//PT100_1_D8_D15               uint8
	//PT100_1_D0_D7                uint8
	//PT100_2_D24_D31              uint8
	//PT100_2_D16_D23              uint8
	//PT100_2_D8_D15               uint8
	//PT100_2_D0_D7                uint8
	//PT100_3_D24_D31              uint8
	//PT100_3_D16_D23              uint8
	//PT100_3_D8_D15               uint8
	//PT100_3_D0_D7                uint8
	//PT100_4_D24_D31              uint8
	//PT100_4_D16_D23              uint8
	//PT100_4_D8_D15               uint8
	//PT100_4_D0_D7                uint8
	Pid1P                        uint16
	SuctionSaturationTemperature float64
	SuperheatTemperature         float64
	mu                           sync.Mutex
}

func (p *ParamsType) setDemandStatus(st uint8) {
	p.DemandStatus.InputY1 = ((st & inputY1) != 0)
	p.DemandStatus.InputY2 = ((st & inputY2) != 0)
	p.DemandStatus.InputY3 = ((st & inputY3) != 0)
	p.DemandStatus.InputO = ((st & inputO) != 0)
	p.DemandStatus.Antifreeze = ((st & inputAntiFreeze) != 0)
	p.DemandStatus.EEVOpen = ((st & inputEEVOpen) != 0)
	p.DemandStatus.WaterValveOpen = ((st & inputWaterValveOpen) != 0)
	p.DemandStatus.EEVSOpen = ((st & inputEEVSOpen) != 0)
}

func (p *ParamsType) getDemandStatus() byte {
	var status byte
	if p.DemandStatus.InputY1 {
		status |= inputY1
	}
	if p.DemandStatus.InputY2 {
		status |= inputY2
	}
	if p.DemandStatus.InputY3 {
		status |= inputY3
	}
	if p.DemandStatus.InputO {
		status |= inputO
	}
	if p.DemandStatus.Antifreeze {
		status |= inputAntiFreeze
	}
	if p.DemandStatus.EEVOpen {
		status |= inputEEVOpen
	}
	if p.DemandStatus.WaterValveOpen {
		status |= inputWaterValveOpen
	}
	if p.DemandStatus.EEVSOpen {
		status |= inputEEVSOpen
	}
	return status
}

func (p *ParamsType) getErrorFlags() uint16 {
	var flags uint16
	if p.Errors.FreezeCondition1 {
		flags |= 1
	}
	if p.Errors.FreezeCondition2 {
		flags |= 2
	}
	if p.Errors.DriveAlarm {
		flags |= 4
	}
	if p.Errors.WaterFlow {
		flags |= 8
	}
	if p.Errors.OutdoorTemperature {
		flags |= 0x10
	}
	if p.Errors.CriticalAlarm {
		flags |= 0x20
	}
	if p.Errors.DriveCommunications {
		flags |= 0x40
	}
	if p.Errors.DischargeOverPressure {
		flags |= 0x80
	}
	if p.Errors.DischargePressure {
		flags |= 0x100
	}
	if p.Errors.SuctionPressure {
		flags |= 0x200
	}
	if p.Errors.EvaporatorSuctionTemp {
		flags |= 0x400
	}
	if p.Errors.LoadInletTemperature {
		flags |= 0x800
	}
	if p.Errors.LoadOutletTemperature {
		flags |= 0x1000
	}
	if p.Errors.SourceInletTemperature {
		flags |= 0x2000
	}
	if p.Errors.SourceOutletTemperature {
		flags |= 0x4000
	}
	return flags
}

func (p *ParamsType) setErrorFlags(error0 byte, error1 byte, error2 byte) {
	p.Errors.DriveCommunications = ((error0 & driveCommunicationsErrorFlag) != 0)
	p.Errors.EvaporatorSuctionTemp = ((error0 & evaporatorSuctionTempErrorFlag) != 0)
	p.Errors.SuctionPressure = ((error0 & suctionPressureErrorFlag) != 0)
	p.Errors.DischargePressure = ((error0 & dischargePressureErrorFlag) != 0)
	p.Errors.SourceOutletTemperature = ((error0 & sourceOutletTempErrorFlag) != 0)
	p.Errors.DischargeOverPressure = ((error0 & dischargeOverPressureErrorFlag) != 0)
	p.Errors.FreezeCondition2 = ((error0 & freezeCondition2ErrorFlag) != 0)
	p.Errors.CriticalAlarm = ((error1 & crititcalAlarmErrorFlag) != 0)
	p.Errors.LoadOutletTemperature = ((error1 & loadOutletTempErrorFlag) != 0)
	p.Errors.SourceInletTemperature = ((error1 & sourceInletTempErrorFlag) != 0)
	p.Errors.LoadInletTemperature = ((error1 & loadInletTempErrorFlag) != 0)
	p.Errors.OutdoorTemperature = ((error1 & outdoorTempSetErrorFlag) != 0)
	p.Errors.FreezeCondition1 = ((error1 & freezeCondition1ErrorFlag) != 0)
	p.Errors.WaterFlow = ((error1 & waterFlowErrorFlag) != 0)
	p.Errors.DriveAlarm = ((error2 & driveInAlarmErrorFlag) != 0)
}

func (p *ParamsType) setValues(paramBuf []byte) {
	Params.DischargePressure = binary.BigEndian.Uint16(paramBuf[0:2])
	Params.SuctionPressure = binary.BigEndian.Uint16(paramBuf[2:4])
	Params.SourceOutTemp = ToTemperature(paramBuf[4:6]) //ToTemperature(paramBuf[4:6]) / 16) - 55
	Params.SuctionEvaporatorTemperature = ToTemperature(paramBuf[6:8])
	Params.SuctionSaturationTemperature = ToTemperature(paramBuf[26:28])
	Params.SuperheatTemperature = Params.SuctionEvaporatorTemperature - Params.SuctionSaturationTemperature
	Params.SourceInTemp = ToTemperature(paramBuf[8:10])
	Params.LoadTempOut = ToTemperature(paramBuf[10:12])
	Params.OutdoorTemp = ToTemperature(paramBuf[12:14])
	Params.EEVRequestedPosition = binary.BigEndian.Uint16(paramBuf[14:16])
	Params.LoadTempIn = ToTemperature(paramBuf[16:18])
	Params.CompressorSpeed = binary.BigEndian.Uint16(paramBuf[18:20])

	Params.Drive1Alarms = paramBuf[35]

	//	Params.EEV2RequestedPosition = binary.BigEndian.Uint16(paramBuf[36:38])
	//	Params.LoadWaterValvePosition = paramBuf[38]
	//	Params.Drive2Alarms = paramBuf[39]
	//	Params.setErrorFlags(paramBuf[43], paramBuf[51], paramBuf[46])
	//	Params.Drive1AlarmTimer = binary.BigEndian.Uint16(paramBuf[44:46])
	Params.Drive1AlarmTimer = uint16(paramBuf[44])
	//	Params.TimerAuxHeater = binary.BigEndian.Uint16(paramBuf[52:54])
	Params.setDemandStatus(paramBuf[54])

	Params.PIDSetpoint = binary.BigEndian.Uint16(paramBuf[87:89])

	//	Params.Drive2AlarmTimer = paramBuf[58]
	//	Params.Work3 = paramBuf[59]
	//	Params.DemandG1316 = paramBuf[60]
	Params.Compressor1Current = binary.BigEndian.Uint16(paramBuf[61:63])
	//	Params.Compressor1Power = binary.BigEndian.Uint16(paramBuf[63:65])
	Params.Compressor1Power = paramBuf[63]

	//Params.Y1MinimumOutputTemperature = binary.BigEndian.Uint16(paramBuf[65:67])
	//Params.Y1MaximumOutputTemperature = binary.BigEndian.Uint16(paramBuf[67:69])
	//Params.Y1MinimumWaterTemperature = binary.BigEndian.Uint16(paramBuf[69:71])
	//Params.Y1MaximumWaterTemperature = binary.BigEndian.Uint16(paramBuf[71:73])
	//Params.Y2MinimumOutputTemperature = binary.BigEndian.Uint16(paramBuf[73:75])
	//Params.Y2MaximumOutputTemperature = binary.BigEndian.Uint16(paramBuf[75:77])
	//Params.Y2MinimumWaterTemperature = binary.BigEndian.Uint16(paramBuf[77:79])
	//Params.Y2MaximumWaterTemperature = binary.BigEndian.Uint16(paramBuf[79:81])
	//	Params.WaterSetpoint = binary.BigEndian.Uint16(paramBuf[81:83])

	Params.PID1Setpoint = binary.BigEndian.Uint16(paramBuf[87:89])

	Params.CriticalAlarmTimer = binary.BigEndian.Uint16(paramBuf[82:84])

	//	Params.LowPressureTimer = binary.BigEndian.Uint16(paramBuf[87:89])

	//	Params.CompressorVoltage = binary.BigEndian.Uint16(paramBuf[89:91])
	Params.CompressorVoltage = paramBuf[62]

	//	Params.CompressorCurrent = binary.BigEndian.Uint16(paramBuf[91:93])
	Params.CompressorCurrent = paramBuf[61]

	//	Params.SourcePressure = binary.BigEndian.Uint16(paramBuf[93:95])
	Params.SourcePressure = binary.BigEndian.Uint16(paramBuf[36:38])

	//	Params.LoadPressure = binary.BigEndian.Uint16(paramBuf[95:97])
	//	Params.PotSourceHP = binary.BigEndian.Uint16(paramBuf[97:99])
	//	Params.PotSourceLP = binary.BigEndian.Uint16(paramBuf[99:101])
	//	Params.PotLoadHP = binary.BigEndian.Uint16(paramBuf[101:103])
	//	Params.PotLoadLP = binary.BigEndian.Uint16(paramBuf[103:105])

	//	Params.CompressorActualSpeed = binary.BigEndian.Uint16(paramBuf[105:107])
	Params.CompressorActualSpeed = binary.BigEndian.Uint16(paramBuf[84:86])

	//Params.PT100_1_D24_D31 = paramBuf[107]
	//Params.PT100_1_D16_D23 = paramBuf[108]
	//Params.PT100_1_D8_D15 = paramBuf[109]
	//Params.PT100_1_D0_D7 = paramBuf[110]
	//Params.PT100_2_D24_D31 = paramBuf[111]
	//Params.PT100_2_D16_D23 = paramBuf[112]
	//Params.PT100_2_D8_D15 = paramBuf[113]
	//Params.PT100_2_D0_D7 = paramBuf[114]
	//Params.PT100_3_D24_D31 = paramBuf[115]
	//Params.PT100_3_D16_D23 = paramBuf[116]
	//Params.PT100_3_D8_D15 = paramBuf[117]
	//Params.PT100_3_D0_D7 = paramBuf[118]
	//Params.PT100_4_D24_D31 = paramBuf[119]
	//Params.PT100_4_D16_D23 = paramBuf[120]
	//Params.PT100_4_D8_D15 = paramBuf[121]
	//Params.PT100_4_D0_D7 = paramBuf[122]

	//	Params.Pid1P = binary.BigEndian.Uint16(paramBuf[123:125])
	Params.Pid1P = binary.BigEndian.Uint16(paramBuf[87:89])

}

func (p *ParamsType) getJSON() ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return json.MarshalIndent(p, "", "    ")
}
