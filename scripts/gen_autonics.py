"""Generate 10 produk Autonics sesuai template import produk backend.

Kolom (sama seperti Schneider): product_code, name, brand, unit,
minimum_stock, status, warehouse_code, stock_system.
"""
import openpyxl
from openpyxl.styles import Font, PatternFill

OUT = "autonics-10-produk.xlsx"

cols = [
    "product_code", "name", "brand", "unit", "minimum_stock", "status",
    "warehouse_code", "stock_system",
]

STOCK_AWAL = 10  # stok sistem awal per produk

# 10 produk Autonics realistis (kode SKU-style + nama lengkap).
products = [
    ("AUT-PR18-5DP", "Proximity Sensor PR18-5DP Autonics"),
    ("AUT-PR12-4DP", "Proximity Sensor PR12-4DP Autonics"),
    ("AUT-BEN5M-MDT", "Photoelectric Sensor BEN5M-MDT Autonics"),
    ("AUT-CT6S-2P4", "Digital Counter CT6S-2P4 Autonics"),
    ("AUT-TIME-FS4E", "Digital Timer FS4E Autonics"),
    ("AUT-TC4S-24R", "Temperature Controller TC4S-24R Autonics"),
    ("AUT-E3Z-R61", "Photoelectric Sensor E3Z-R61 Autonics"),
    ("AUT-SR-2P", "Solid State Relay SR-2P Autonics"),
    ("AUT-M4Y-AC", "Multi Panel Meter M4Y Autonics"),
    ("AUT-SW7M-PF", "SMPS Power Supply SW7M-PF Autonics"),
]
assert len(products) == 10, f"jumlah produk {len(products)}"

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
    ws.cell(row=r, column=1, value=code)              # product_code
    ws.cell(row=r, column=2, value=name)              # name
    ws.cell(row=r, column=3, value="Autonics")        # brand
    ws.cell(row=r, column=4, value="PCS")             # unit
    ws.cell(row=r, column=5, value=5)                 # minimum_stock
    ws.cell(row=r, column=6, value="active")          # status
    ws.cell(row=r, column=7, value="WH01")            # warehouse_code
    ws.cell(row=r, column=8, value=STOCK_AWAL)        # stock_system

wb.save(OUT)
print(f"OK: {OUT} — {len(products)} baris data")
