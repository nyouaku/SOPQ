// ===== FILE: mobile/lib/data/models/product_model.dart =====
/// Bentuk mengikuti `domain.Product` di backend.
/// Identitas produk = `product_code` (juga isi QR). Tidak ada barcode/kategori.
/// Catatan: `stock_system` TIDAK dikirim oleh endpoint produk — nilai stok sistem
/// dihitung di backend saat scan, jadi di sini null dan hanya untuk tampilan.
class ProductModel {
  final String id;
  final String productCode;
  final String name;
  final String? brand;
  final String unit;
  final String? supplierName;
  final String? photoUrl;
  final int? stockSystem;

  ProductModel({
    required this.id,
    required this.productCode,
    required this.name,
    this.brand,
    this.unit = 'PCS',
    this.supplierName,
    this.photoUrl,
    this.stockSystem,
  });

  /// Alias lama: layar memakai `product.code` sebagai identifier tampilan.
  String get code => productCode;

  factory ProductModel.fromJson(Map<String, dynamic> json) {
    final supplier = json['supplier'] as Map<String, dynamic>?;
    return ProductModel(
      id: json['id'] as String,
      productCode: json['product_code'] as String? ?? '',
      name: json['name'] as String? ?? '',
      brand: json['brand'] as String?,
      unit: json['unit'] as String? ?? 'PCS',
      supplierName: supplier?['name'] as String?,
      photoUrl: json['photo_url'] as String?,
      stockSystem: json['stock_system'] as int?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'product_code': productCode,
      'name': name,
      'brand': brand,
      'unit': unit,
      'photo_url': photoUrl,
      'stock_system': stockSystem,
    };
  }
}
// ===== END FILE =====
