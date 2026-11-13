// ===== FILE: mobile/lib/data/datasources/scan_local_datasource.dart =====
import 'package:sqflite/sqflite.dart';
import 'package:uuid/uuid.dart';
import '../../core/storage/local_database.dart';

enum SyncStatus { pending, syncing, synced, failed }

class ScanQueueItem {
  final String id;
  final String sessionId;
  final String productId;
  final String productCode;
  final String productName;
  final int quantityPhysical;
  final String inputMethod;
  final DateTime createdAt;
  final SyncStatus syncStatus;
  final int retryCount;
  final String? errorMessage;

  ScanQueueItem({
    required this.id,
    required this.sessionId,
    required this.productId,
    required this.productCode,
    required this.productName,
    required this.quantityPhysical,
    required this.inputMethod,
    required this.createdAt,
    required this.syncStatus,
    this.retryCount = 0,
    this.errorMessage,
  });

  /// [id] adalah UUID yang dibuat sekali saat baris masuk antrean. Nilai yang
  /// sama dikirim sebagai `idempotency_key`, jadi retry tidak menambah scan baru.
  String get idempotencyKey => id;

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'session_id': sessionId,
      'product_id': productId,
      'product_code': productCode,
      'product_name': productName,
      'quantity_physical': quantityPhysical,
      'input_method': inputMethod,
      'created_at': createdAt.toIso8601String(),
      'sync_status': syncStatus.name,
      'retry_count': retryCount,
      'error_message': errorMessage,
    };
  }

  factory ScanQueueItem.fromMap(Map<String, dynamic> map) {
    return ScanQueueItem(
      id: map['id'] as String,
      sessionId: map['session_id'] as String,
      productId: map['product_id'] as String,
      productCode: map['product_code'] as String,
      productName: map['product_name'] as String,
      quantityPhysical: map['quantity_physical'] as int,
      inputMethod: map['input_method'] as String,
      createdAt: DateTime.parse(map['created_at'] as String),
      syncStatus: SyncStatus.values.firstWhere(
        (e) => e.name == map['sync_status'],
        orElse: () => SyncStatus.pending,
      ),
      retryCount: map['retry_count'] as int? ?? 0,
      errorMessage: map['error_message'] as String?,
    );
  }
}

class ScanLocalDataSource {
  final LocalDatabase _localDatabase;
  final _uuid = const Uuid();

  ScanLocalDataSource(this._localDatabase);

  Future<String> addToQueue({
    required String sessionId,
    required String productId,
    required String productCode,
    required String productName,
    required int quantityPhysical,
    required String inputMethod,
  }) async {
    final db = await _localDatabase.database;
    final id = _uuid.v4();

    final item = ScanQueueItem(
      id: id,
      sessionId: sessionId,
      productId: productId,
      productCode: productCode,
      productName: productName,
      quantityPhysical: quantityPhysical,
      inputMethod: inputMethod,
      createdAt: DateTime.now(),
      syncStatus: SyncStatus.pending,
    );

    await db.insert('scan_queue', item.toMap());
    return id;
  }

  Future<List<ScanQueueItem>> getPendingScans() async {
    final db = await _localDatabase.database;
    final result = await db.query(
      'scan_queue',
      where: 'sync_status = ? AND retry_count < ?',
      whereArgs: [SyncStatus.pending.name, 5],
      orderBy: 'created_at ASC',
    );
    return result.map((map) => ScanQueueItem.fromMap(map)).toList();
  }

  /// Baris yang tertinggal di `syncing` karena aplikasi ditutup di tengah
  /// pengiriman. Tanpa ini baris itu tidak pernah diklaim lagi.
  Future<void> resetStuckSyncing() async {
    final db = await _localDatabase.database;
    await db.update(
      'scan_queue',
      {'sync_status': SyncStatus.pending.name},
      where: 'sync_status = ?',
      whereArgs: [SyncStatus.syncing.name],
    );
  }

  /// Baris `failed` yang masih di bawah batas percobaan. Yang melewatinya
  /// tetap tersimpan sebagai dead-letter dan tidak dikirim otomatis lagi.
  Future<List<ScanQueueItem>> getFailedScans() async {
    final db = await _localDatabase.database;
    final result = await db.query(
      'scan_queue',
      where: 'sync_status = ? AND retry_count < ?',
      whereArgs: [SyncStatus.failed.name, 5],
      orderBy: 'created_at ASC',
    );
    return result.map((map) => ScanQueueItem.fromMap(map)).toList();
  }

  Future<int> getPendingCount() async {
    final db = await _localDatabase.database;
    final result = await db.rawQuery(
      'SELECT COUNT(*) as count FROM scan_queue WHERE sync_status IN (?, ?)',
      [SyncStatus.pending.name, SyncStatus.failed.name],
    );
    return Sqflite.firstIntValue(result) ?? 0;
  }

  /// Klaim satu baris untuk satu putaran sinkronisasi. UPDATE bersyarat
  /// membuat dua putaran yang berjalan bersamaan tidak mengirim scan yang sama.
  Future<bool> claimForSync(String id) async {
    final db = await _localDatabase.database;
    final affected = await db.update(
      'scan_queue',
      {'sync_status': SyncStatus.syncing.name},
      where: 'id = ? AND sync_status IN (?, ?)',
      whereArgs: [id, SyncStatus.pending.name, SyncStatus.failed.name],
    );
    return affected > 0;
  }

  Future<void> updateSyncStatus(
    String id,
    SyncStatus status, {
    String? errorMessage,
  }) async {
    final db = await _localDatabase.database;
    await db.update(
      'scan_queue',
      {
        'sync_status': status.name,
        'error_message': errorMessage,
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  Future<void> incrementRetryCount(String id) async {
    final db = await _localDatabase.database;
    await db.rawUpdate(
      'UPDATE scan_queue SET retry_count = retry_count + 1 WHERE id = ?',
      [id],
    );
  }

  Future<void> deleteSynced() async {
    final db = await _localDatabase.database;
    await db.delete(
      'scan_queue',
      where: 'sync_status = ?',
      whereArgs: [SyncStatus.synced.name],
    );
  }

  Future<List<ScanQueueItem>> getSessionScans(String sessionId) async {
    final db = await _localDatabase.database;
    final result = await db.query(
      'scan_queue',
      where: 'session_id = ?',
      whereArgs: [sessionId],
      orderBy: 'created_at DESC',
    );
    return result.map((map) => ScanQueueItem.fromMap(map)).toList();
  }
}
// ===== END FILE =====
