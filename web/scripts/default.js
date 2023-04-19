function loadStatus() {
    fetch("/getStatus")
        .then( function(response) {
            if (response.status === 200) {
                response.json()
                    .then(function (data) {
                        updateFields(data);
                    });
            }
        })
}

function setOnOff(field, value) {
    if (value) {
        field.attr('src',"/images/on.png");
    } else {
        field.attr('src',"/images/off.png");
    }
}

function setHotCold(field, value) {
    if (value) {
        field.attr('src',"/images/cooling.png");
    } else {
        field.attr('src',"/images/heating.png");
    }
}

function showTimeoutMessage() {
    $("#connection").show();
}

var wstimeout;

function receiveStatus() {
    setupGauges();
    let url = "ws://" + window.location.host + "/ws";
    let conn = new WebSocket(url);
    wsTimeout = 0;

    // let  Data = document.getElementById("Data");

    conn.onclose = function () {
        $("#connection").show();
    }

    conn.onmessage = function (evt) {
        if (wstimeout !== 0) {
            clearTimeout(wstimeout);
            $("#connection").hide();
        }
        wstimeout = setTimeout(showTimeoutMessage, 15000)
        try {
            jsonData = JSON.parse(evt.data);
            updateFields(jsonData);
        } catch (e) {
            alert(e);
        }
    }
}

function setAlarm(status, control) {
    if (status) {
        control.show();
        $("#alarmHeader").show();
    } else {
        control.hide();
    }
}

function updateFields(data) {
    $("#DischargePressure").text(data.DischargePressure);
    $("#SuctionPressure").text(data.SuctionPressure);
    $("#SourceInTemp").text(data.SourceInTemp.toFixed(2));
    $("#SourceOutTemp").text(data.SourceOutTemp.toFixed(2));
    $("#LoadTempIn").text(data.LoadTempIn.toFixed(2));
    $("#LoadTempOut").text(data.LoadTempOut.toFixed(2));
    $("#SuctionTemp").text(data.SuctionEvaporatorTemperature.toFixed(2));
    $("#SaturationTemp").text(data.SuctionSaturationTemperature.toFixed(2));
    $("#SuperHeatTemp").text(data.SuperheatTemperature.toFixed(2));
    $("#CompressorSpeed").text(data.CompressorSpeed);
    $("#EEVPosition").text(data.EEVRequestedPosition);
    setOnOff($("#InputY1"), data.DemandStatus.InputY1);
    setOnOff($("#InputY2"), data.DemandStatus.InputY2);
    setOnOff($("#InputY3"), data.DemandStatus.InputY3);
    setHotCold($("#InputO"), data.DemandStatus.InputO);
    $("#alarmHeader").hide();
    setAlarm(data.Errors.DriveCommunications, $("#DriveCommunicationsAlarm"));
    setAlarm(data.Errors.EvaporatorSuctionTemp, $("#EvaporatorSuctionTempAlarm"));
    setAlarm(data.Errors.SuctionPressure, $("#SuctionPressureAlarm"));
    setAlarm(data.Errors.DischargePressure, $("#DischargePressureAlarm"));
    setAlarm(data.Errors.DischargeOverPressure, $("#DischargeOverPressureAlarm"));
    setAlarm(data.Errors.FreezeCondition2, $("#FreezeCondition2Alarm"));
    setAlarm(data.Errors.CriticalAlarm, $("#CriticalAlarm"));
    setAlarm(data.Errors.LoadInletTemperature, $("#LoadInletTemperatureAlarm"));
    setAlarm(data.Errors.LoadOutletTemperature, $("#LoadOutletTemperatureAlarm"));
    setAlarm(data.Errors.SourceInletTemperature, $("#SourceInletTemperatureAlarm"));
    setAlarm(data.Errors.SourceOutletTemperature, $("#SourceOutletTemperatureAlarm"));
    setAlarm(data.Errors.FreezeCondition1, $("#FreezeCondition1Alarm"));
    setAlarm(data.Errors.WaterFlow, $("#WaterFlowAlarm"));
    setAlarm(data.Errors.DriveAlarm, $("#DriveAlarm"));
    $("#Pressures").val([data.DischargePressure, data.SuctionPressure]);
    $("#SourceTemp").val([data.SourceInTemp, data.SourceOutTemp]);
    $("#LoadTemp").val([data.LoadTempIn, data.LoadTempOut]);
    $("#SuctionTemperatures").val([data.SuctionEvaporatorTemperature, data.SuctionSaturationTemperature, data.SuperheatTemperature]);
    $("#CompressorSpeedDial").val(data.CompressorSpeed);
    $("#EEVPositionDial").val(data.EEVRequestedPosition);
}

function toggleCoil(id) {
    var xhr = new XMLHttpRequest();
    xhr.open('PATCH','toggleCoil?coil=' + id);
    xhr.send();
}

function setupGauges() {
    let controlsHeight = window.innerHeight / 2;
    let controlsWidth = window.innerWidth  / 4;
    let gaugeRadius = controlsWidth / 2;
    if (controlsWidth > controlsHeight) {
        gaugeRadius = controlsHeight / 2;
    }
    let gaugeDiameter = gaugeRadius * 2;
    $("#Pressures").jqxBarGauge({
        width: controlsWidth,
        height: controlsHeight,
        values: [0.0, 0.0],
        min: 50,
        max: 300,
        animationDuration: 500,
        startAngle: 265,
        endAngle: 275,
        title:	{
            text: 'Pressures',
            font: { size: 12, color: 'black', weight: 'bold', family:"Segoi-UI"},
            margin: { top: 0, bottom: 2, left: 0, right: 0},
            verticalAlignment: 'bottom'
        },
        labels: {
            font: {size: 8,},
            precision: 1,
        },
        colorScheme: 'customColors',
        customColorScheme: { name: 'customColors', colors: ['#000066', '#006600'] },
        tooltip: {visible: true,
            precision: 1,
            formatFunction: function (value, index){
                switch(index)
                {
                    case 0 : return("Discharge = " + value);
                    default : return ("Suction = " + value);
                }
            }}
    });
    $("#SourceTemp").jqxBarGauge({
        width: controlsWidth,
        height: controlsHeight,
        values: [0.0, 0.0],
        min: 5,
        max: 40,
        animationDuration: 500,
        startAngle: 265,
        endAngle: 275,
        title:	{
            text: 'Source Temperatures',
            font: { size: 12, color: 'black', weight: 'bold', family:"Segoi-UI"},
            margin: { top: 0, bottom: 2, left: 0, right: 0},
            verticalAlignment: 'bottom'
        },
        labels: {
            font: {size: 8,},
            precision: 1,
        },
        colorScheme: 'customColors',
        customColorScheme: { name: 'customColors', colors: ['#00A600', '#008233'] },
        tooltip: {visible: true,
            precision: 1,
            formatFunction: function (value, index){
                switch(index)
                {
                    case 0 : return("Inlet = " + value);
                    default : return ("Outlet = " + value);
                }
            }}
    });
    $("#LoadTemp").jqxBarGauge({
        width: controlsWidth,
        height: controlsHeight,
        values: [0.0, 0.0],
        min: 0,
        max: 60,
        animationDuration: 500,
        startAngle: 265,
        endAngle: 275,
        title:	{
            text: 'Load Temperatures',
            font: { size: 12, color: 'black', weight: 'bold', family:"Segoi-UI"},
            margin: { top: 0, bottom: 2, left: 0, right: 0},
            verticalAlignment: 'bottom'
        },
        labels: {
            font: {size: 8,},
            precision: 1,
        },
        colorScheme: 'customColors',
        customColorScheme: { name: 'customColors', colors: ['#A60000', '#820033'] },
        tooltip: {visible: true,
            precision: 1,
            formatFunction: function (value, index){
                switch(index)
                {
                    case 0 : return("Inlet = " + value);
                    default : return ("Outlet = " + value);
                }
            }}
    });
    $("#SuctionTemperatures").jqxBarGauge({
        width: controlsWidth,
        height: controlsHeight,
        values: [0.0, 0.0, 0.0],
        min: -5,
        max: 30,
        animationDuration: 500,
        startAngle: 265,
        endAngle: 275,
        title:	{
            text: 'Suction/Superheat Temperatures',
            font: { size: 12, color: 'black', weight: 'bold', family:"Segoi-UI"},
            margin: { top: 0, bottom: 2, left: 0, right: 0},
            verticalAlignment: 'bottom'
        },
        labels: {
            font: {size: 8,},
            precision: 1,
        },
        colorScheme: 'customColors',
        customColorScheme: { name: 'customColors', colors: ['#A60000', '#820033', '#330082'] },
        tooltip: {visible: true,
            precision: 1,
            formatFunction: function (value, index){
                switch(index)
                {
                    case 0 : return("Evaporator = " + value);
                    case 1 : return("Saturation = " + value);
                    default : return ("Superheat = " + value);
                }
            }}
    });
    $('#CompressorSpeedDial').jqxGauge({
        height: gaugeDiameter - 50,
        width: gaugeDiameter - 50,
        radius: gaugeRadius - 25,
        ticksMinor: {interval: 50, size: '5%'},
        ticksMajor: {interval: 200,size: '9%'},
        labels: {interval:400},
        min: 0,
        max: 1400,
        value: 0,
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Compressor RPM', position: 'bottom', offset: [0, 10], visible: true},
    });
    $('#EEVPositionDial').jqxGauge({
        height: gaugeDiameter - 50,
        width: gaugeDiameter - 50,
        radius: gaugeRadius - 25,
        ticksMinor: {interval: 50, size: '5%'},
        ticksMajor: {interval: 200,size: '9%'},
        labels: {interval:400},
        min: 0,
        max: 1500,
        value: 0,
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'EEV Position', position: 'bottom', offset: [0, 10], visible: true},
    });
}