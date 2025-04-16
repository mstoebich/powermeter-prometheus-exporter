# powermeter-prometheus-exporter

This exporter exposes metrics from from an Orno OR-WE-514 (tested) and OR-WE-515 (untested) using a modbus-tcp gateway.


## registers

| Wert                   | Register (Hex) | Größe  | Skalierung | Einheit |
|------------------------|----------------|--------|------------|---------|
| Frequenz               | 0x0130         | 1 WORD | 0.01       | Hz      |
| Spannung L1            | 0x0131         | 1 WORD | 0.01       | V       |
| Strom L1               | 0x0139         | 2 WORD | 0.001      | A       |
| Wirkleistung L1        | 0x0140         | 2 WORD | 0.001      | kW      |
| Blindleistung L1       | 0x0148         | 2 WORD | 0.001      | kvar    |
| Scheinleistung L1      | 0x0150         | 2 WORD | 0.001      | kVA     |
| Leistungsfaktor L1     | 0x0158         | 1 WORD | 0.001      | –       |
| Gesamtenergieverbrauch | 0xA000         | 2 WORD | 0.001      | kWh     |

# modbus-rtu connection

Baudrate: 9600
Databits: 8
Stopbits: 1
Parity: Even
Flow control: None