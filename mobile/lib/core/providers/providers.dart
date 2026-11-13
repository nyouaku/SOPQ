// ===== FILE: mobile/lib/core/providers/providers.dart =====
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../network/dio_client.dart';
import '../storage/local_database.dart';
import '../../data/datasources/auth_remote_datasource.dart';
import '../../data/datasources/product_remote_datasource.dart';
import '../../data/datasources/stock_opname_remote_datasource.dart';
import '../../data/datasources/scan_local_datasource.dart';

// Core Providers
final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage();
});

final dioClientProvider = Provider<DioClient>((ref) {
  final secureStorage = ref.watch(secureStorageProvider);
  return DioClient(secureStorage);
});

final localDatabaseProvider = Provider<LocalDatabase>((ref) {
  return LocalDatabase();
});

// Remote DataSource Providers
final authRemoteDataSourceProvider = Provider<AuthRemoteDataSource>((ref) {
  final dioClient = ref.watch(dioClientProvider);
  return AuthRemoteDataSource(dioClient.dio);
});

final productRemoteDataSourceProvider = Provider<ProductRemoteDataSource>((ref) {
  final dioClient = ref.watch(dioClientProvider);
  return ProductRemoteDataSource(dioClient.dio);
});

final stockOpnameRemoteDataSourceProvider = Provider<StockOpnameRemoteDataSource>((ref) {
  final dioClient = ref.watch(dioClientProvider);
  return StockOpnameRemoteDataSource(dioClient.dio);
});

// Local DataSource Providers
final scanLocalDataSourceProvider = Provider<ScanLocalDataSource>((ref) {
  final localDb = ref.watch(localDatabaseProvider);
  return ScanLocalDataSource(localDb);
});
// ===== END FILE =====
