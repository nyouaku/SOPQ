// ===== FILE: mobile/lib/data/models/stock_opname_model.dart =====
class StockOpnameSessionModel {
  final String id;
  final String code;
  final String warehouseId;
  final String? warehouseName;
  final String? locationId;
  final String status;
  final String startedBy;
  final String? startedByName;
  final DateTime startedAt;
  final DateTime? completedAt;

  StockOpnameSessionModel({
    required this.id,
    required this.code,
    required this.warehouseId,
    this.warehouseName,
    this.locationId,
    required this.status,
    required this.startedBy,
    this.startedByName,
    required this.startedAt,
    this.completedAt,
  });

  factory StockOpnameSessionModel.fromJson(Map<String, dynamic> json) {
    final warehouse = json['warehouse'] as Map<String, dynamic>?;
    final startedUser = json['started_user'] as Map<String, dynamic>?;
    return StockOpnameSessionModel(
      id: json['id'] as String,
      code: json['code'] as String,
      warehouseId: json['warehouse_id'] as String,
      warehouseName: warehouse?['name'] as String?,
      locationId: json['location_id'] as String?,
      status: json['status'] as String,
      startedBy: json['started_by'] as String,
      startedByName: startedUser?['full_name'] as String?,
      startedAt: DateTime.parse(json['started_at'] as String),
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
    );
  }
}

class StockOpnameDetailModel {
  final String id;
  final String stockOpnameId;
  final String productId;
  final String productCode;
  final String productName;
  final int stockSystem;
  final int stockPhysical;
  final int difference;
  final String inputMethod;
  final DateTime scannedAt;

  StockOpnameDetailModel({
    required this.id,
    required this.stockOpnameId,
    required this.productId,
    required this.productCode,
    required this.productName,
    required this.stockSystem,
    required this.stockPhysical,
    required this.difference,
    required this.inputMethod,
    required this.scannedAt,
  });

  factory StockOpnameDetailModel.fromJson(Map<String, dynamic> json) {
    final product = json['product'] as Map<String, dynamic>?;
    return StockOpnameDetailModel(
      id: json['id'] as String,
      stockOpnameId: json['stock_opname_id'] as String,
      productId: json['product_id'] as String,
      productCode: product?['product_code'] as String? ?? '',
      productName: product?['name'] as String? ?? '',
      stockSystem: json['stock_system'] as int,
      stockPhysical: json['stock_physical'] as int,
      difference: json['difference'] as int,
      inputMethod: json['input_method'] as String,
      scannedAt: DateTime.parse(json['scanned_at'] as String),
    );
  }
}

class SubmitScanRequest {
  final String productCode;
  final int stockPhysical;
  final String inputMethod;
  final String? note;
  final String? idempotencyKey;

  SubmitScanRequest({
    required this.productCode,
    required this.stockPhysical,
    required this.inputMethod,
    this.note,
    this.idempotencyKey,
  });

  Map<String, dynamic> toJson() {
    return {
      'product_code': productCode,
      'stock_physical': stockPhysical,
      'input_method': inputMethod,
      if (note != null) 'note': note,
      if (idempotencyKey != null) 'idempotency_key': idempotencyKey,
    };
  }
}
// ===== END FILE =====
