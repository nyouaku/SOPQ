// ===== FILE: mobile/lib/core/storage/local_database.dart =====
import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';
import '../constants/app_constants.dart';

class LocalDatabase {
  static Database? _database;

  Future<Database> get database async {
    if (_database != null) return _database!;
    _database = await _initDatabase();
    return _database!;
  }

  Future<Database> _initDatabase() async {
    final dbPath = await getDatabasesPath();
    final path = join(dbPath, AppConstants.dbName);

    return await openDatabase(
      path,
      version: AppConstants.dbVersion,
      onCreate: _onCreate,
      onUpgrade: _onUpgrade,
    );
  }

  Future<void> _onCreate(Database db, int version) async {
    // Tabel untuk queue scan offline
    await db.execute('''
      CREATE TABLE scan_queue (
        id TEXT PRIMARY KEY,
        session_id TEXT NOT NULL,
        product_id TEXT NOT NULL,
        product_code TEXT NOT NULL,
        product_name TEXT NOT NULL,
        quantity_physical INTEGER NOT NULL,
        input_method TEXT NOT NULL,
        created_at TEXT NOT NULL,
        sync_status TEXT NOT NULL,
        retry_count INTEGER DEFAULT 0,
        error_message TEXT
      )
    ''');

    // Tabel untuk cache session
    await db.execute('''
      CREATE TABLE cached_sessions (
        id TEXT PRIMARY KEY,
        session_code TEXT NOT NULL,
        warehouse_id TEXT NOT NULL,
        warehouse_name TEXT,
        status TEXT NOT NULL,
        created_by TEXT NOT NULL,
        created_at TEXT NOT NULL,
        synced_at TEXT NOT NULL
      )
    ''');

    // Tabel untuk cache product (untuk search offline)
    await db.execute('''
      CREATE TABLE cached_products (
        id TEXT PRIMARY KEY,
        code TEXT NOT NULL,
        name TEXT NOT NULL,
        brand TEXT,
        stock_system INTEGER NOT NULL,
        supplier_name TEXT,
        synced_at TEXT NOT NULL
      )
    ''');

    // Index untuk performa
    await db.execute('CREATE INDEX idx_scan_queue_session ON scan_queue(session_id)');
    await db.execute('CREATE INDEX idx_scan_queue_status ON scan_queue(sync_status)');
    await db.execute('CREATE INDEX idx_cached_products_code ON cached_products(code)');
  }

  /// v1 -> v2: kolom `barcode` dihapus (identitas produk = product_code).
  /// scan_queue DIPERTAHANKAN (bisa berisi scan pending yang belum tersinkron):
  /// tabel dibangun ulang tanpa kolom barcode dan isinya disalin.
  /// cached_products dibangun ulang (cache, aman dibuat ulang dari server).
  Future<void> _onUpgrade(Database db, int oldVersion, int newVersion) async {
    if (oldVersion < 2) {
      await db.transaction((txn) async {
        await txn.execute('DROP INDEX IF EXISTS idx_scan_queue_session');
        await txn.execute('DROP INDEX IF EXISTS idx_scan_queue_status');
        await txn.execute('''
          CREATE TABLE scan_queue_v2 (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL,
            product_id TEXT NOT NULL,
            product_code TEXT NOT NULL,
            product_name TEXT NOT NULL,
            quantity_physical INTEGER NOT NULL,
            input_method TEXT NOT NULL,
            created_at TEXT NOT NULL,
            sync_status TEXT NOT NULL,
            retry_count INTEGER DEFAULT 0,
            error_message TEXT
          )
        ''');
        await txn.execute('''
          INSERT INTO scan_queue_v2
            (id, session_id, product_id, product_code, product_name,
             quantity_physical, input_method, created_at, sync_status, retry_count, error_message)
          SELECT id, session_id, product_id, product_code, product_name,
                 quantity_physical, input_method, created_at, sync_status, retry_count, error_message
          FROM scan_queue
        ''');
        await txn.execute('DROP TABLE scan_queue');
        await txn.execute('ALTER TABLE scan_queue_v2 RENAME TO scan_queue');
        await txn.execute('CREATE INDEX idx_scan_queue_session ON scan_queue(session_id)');
        await txn.execute('CREATE INDEX idx_scan_queue_status ON scan_queue(sync_status)');

        await txn.execute('DROP TABLE IF EXISTS cached_products');
        await txn.execute('''
          CREATE TABLE cached_products (
            id TEXT PRIMARY KEY,
            code TEXT NOT NULL,
            name TEXT NOT NULL,
            brand TEXT,
            stock_system INTEGER NOT NULL,
            supplier_name TEXT,
            synced_at TEXT NOT NULL
          )
        ''');
        await txn.execute('CREATE INDEX idx_cached_products_code ON cached_products(code)');
      });
    }
  }

  Future<void> close() async {
    final db = await database;
    await db.close();
    _database = null;
  }

  Future<void> clearAll() async {
    final db = await database;
    await db.delete('scan_queue');
    await db.delete('cached_sessions');
    await db.delete('cached_products');
  }
}
// ===== END FILE =====
