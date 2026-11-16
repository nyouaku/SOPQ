"""Generate 50 produk Schneider sesuai template import produk backend.

Kolom wajib (dibaca parser backend import_usecase.go):
  0 product_code, 1 name, 2 brand, 3 unit, 4 minimum_stock, 5 status
"""
import openpyxl
from openpyxl.styles import Font, PatternFill

OUT = "schneider-50-produk.xlsx"

cols = ["product_code", "name", "brand", "unit", "minimum_stock", "status"]

# 50 produk Schneider realistis (kode SKU-style + nama).
products = [
    ("SCH-MCB-1P-6A", "MCB 1P 6A iC60N Schneider"),
    ("SCH-MCB-1P-10A", "MCB 1P 10A iC60N Schneider"),
    ("SCH-MCB-1P-16A", "MCB 1P 16A iC60N Schneider"),
    ("SCH-MCB-1P-20A", "MCB 1P 20A iC60N Schneider"),
    ("SCH-MCB-1P-25A", "MCB 1P 25A iC60N Schneider"),
    ("SCH-MCB-1P-32A", "MCB 1P 32A iC60N Schneider"),
    ("SCH-MCB-2P-16A", "MCB 2P 16A iC60N Schneider"),
    ("SCH-MCB-2P-25A", "MCB 2P 25A iC60N Schneider"),
    ("SCH-MCB-2P-32A", "MCB 2P 32A iC60N Schneider"),
    ("SCH-MCB-3P-16A", "MCB 3P 16A iC60N Schneider"),
    ("SCH-MCB-3P-25A", "MCB 3P 25A iC60N Schneider"),
    ("SCH-MCB-3P-40A", "MCB 3P 40A iC60N Schneider"),
    ("SCH-MCCB-3P-100A", "MCCB 3P 100A NSX Schneider"),
    ("SCH-MCCB-3P-160A", "MCCB 3P 160A NSX Schneider"),
    ("SCH-MCCB-3P-250A", "MCCB 3P 250A NSX Schneider"),
    ("SCH-RCD-2P-30MA", "RCD ELCB 2P 30mA Schneider"),
    ("SCH-RCD-4P-30MA", "RCD ELCB 4P 30mA Schneider"),
    ("SCH-SPD-T2-40KA", "Surge Protector T2 40kA Schneider"),
    ("SCH-CONT-9A", "Kontaktor LC1D09 9A Schneider"),
    ("SCH-CONT-12A", "Kontaktor LC1D12 12A Schneider"),
    ("SCH-CONT-18A", "Kontaktor LC1D18 18A Schneider"),
    ("SCH-CONT-25A", "Kontaktor LC1D25 25A Schneider"),
    ("SCH-CONT-40A", "Kontaktor LC1D40 40A Schneider"),
    ("SCH-CONT-50A", "Kontaktor LC1D50 50A Schneider"),
    ("SCH-CONT-65A", "Kontaktor LC1D65 65A Schneider"),
    ("SCH-CONT-80A", "Kontaktor LC1D80 80A Schneider"),
    ("SCH-CONT-95A", "Kontaktor LC1D95 95A Schneider"),
    ("SCH-OLR-1-1.6A", "Overload Relay LRD14 1-1.6A Schneider"),
    ("SCH-OLR-2.5-4A", "Overload Relay LRD16 2.5-4A Schneider"),
    ("SCH-OLR-4-6A", "Overload Relay LRD18 4-6A Schneider"),
    ("SCH-OLR-9-13A", "Overload Relay LRD21 9-13A Schneider"),
    ("SCH-PB-1NO", "Push Button XB4 1NO Schneider"),
    ("SCH-PB-1NC", "Push Button XB4 1NC Schneider"),
    ("SCH-PB-GRN", "Push Button XB4 Green Schneider"),
    ("SCH-PB-RED", "Push Button XB4 Red Schneider"),
    ("SCH-PB-EMG", "Emergency Stop XB4 Schneider"),
    ("SCH-SEL-SW-2P", "Selector Switch XB4 2P Schneider"),
    ("SCH-IND-LAMP-G", "Pilot Lamp XB4 Green Schneider"),
    ("SCH-IND-LAMP-R", "Pilot Lamp XB4 Red Schneider"),
    ("SCH-LIM-SW-XC", "Limit Switch XC Schneider"),
    ("SCH-PROX-PNP", "Proximity Sensor XS PNP Schneider"),
    ("SCH-SWITCH-1G", "Saklar Tunggal 1G Schneider"),
    ("SCH-SWITCH-2G", "Saklar Ganda 2G Schneider"),
    ("SCH-SOCKET-2P-E", "Stopkontak 2P+E Schneider"),
    ("SCH-DISTRIB-12W", "Distribution Box 12 Way Schneider"),
    ("SCH-METER-KWH", "KWh Meter PM210 Schneider"),
    ("SCH-SENSOR-TMP", "Temperature Sensor Schneider"),
    ("SCH-RELAY-24VDC", "Relay RXM 24VDC Schneider"),
    ("SCH-RELAY-230VAC", "Relay RXM 230VAC Schneider"),
    ("SCH-TIMER-24V", "Timer Relay RE17 24V Schneider"),
]
assert len(products) == 50, f"jumlah produk {len(products)}"

wb = openpyxl.Workbook()
ws = wb.active
ws.title = "Products"

header_font = Font(bold=True, color="FFFFFF")
header_fill = PatternFill(fill_type="solid", start_color="2E7D32", end_color="2E7D32")

for c, name in enumerate(cols, start=1):
    cell = ws.cell(row=1, column=c, value=name)
    cell.font = header_font
    cell.fill = header_fill
    ws.column_dimensions[ws.cell(row=1, column=c).column_letter].width = 22

for r, (code, name) in enumerate(products, start=2):
    ws.cell(row=r, column=1, value=code)   # product_code
    ws.cell(row=r, column=2, value=name)   # name
    ws.cell(row=r, column=3, value="Schneider Electric")  # brand
    ws.cell(row=r, column=4, value="PCS")  # unit
    ws.cell(row=r, column=5, value=5)      # minimum_stock
    ws.cell(row=r, column=6, value="active")  # status

wb.save(OUT)
print(f"OK: {OUT} — {len(products)} baris data")
