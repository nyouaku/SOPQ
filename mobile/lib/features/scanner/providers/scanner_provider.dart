// ===== FILE: mobile/lib/features/scanner/providers/scanner_provider.dart =====
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:uuid/uuid.dart';
import '../../../core/providers/providers.dart';
import '../../../core/network/api_response.dart';
import '../../../data/models/product_model.dart';
import '../../../data/models/stock_opname_model.dart';
import '../../../data/datasources/scan_local_datasource.dart';

class ScannerState {
  final ProductModel? scannedProduct;
  final bool isScanning;
  final bool isLoading;
  final String? errorMessage;
  final String? successMessage;
  final bool isOnline;
  final int pendingScansCount;

  ScannerState({
    this.scannedProduct,
    this.isScanning = false,
    this.isLoading = false,
    this.errorMessage,
    this.successMessage,
    this.isOnline = true,
    this.pendingScansCount = 0,
  });

  ScannerState copyWith({
    ProductModel? scannedProduct,
    bool? isScanning,
    bool? isLoading,
    String? errorMessage,
    String? successMessage,
    bool? isOnline,
    int? pendingScansCount,
  }) {
    return ScannerState(
      scannedProduct: scannedProduct ?? this.scannedProduct,
      isScanning: isScanning ?? this.isScanning,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      successMessage: successMessage,
      isOnline: isOnline ?? this.isOnline,
      pendingScansCount: pendingScansCount ?? this.pendingScansCount,
    );
  }
}

class ScannerNotifier extends StateNotifier<ScannerState> {
  final Ref ref;
  String? _currentSessionId;

  ScannerNotifier(this.ref) : super(ScannerState()) {
    _initConnectivityListener();
    _updatePendingCount();
  }

  void setSessionId(String sessionId) {
    _currentSessionId = sessionId;
  }

  void _initConnectivityListener() {
    Connectivity().onConnectivityChanged.listen((result) {
      final isOnline = result != ConnectivityResult.none;
      state = state.copyWith(isOnline: isOnline);

      if (isOnline && state.pendingScansCount > 0) {
        // Trigger sync when coming back online
        ref.read(syncServiceProvider).syncPendingScans();
      }
    });
  }

  Future<void> _updatePendingCount() async {
    final scanLocalDataSource = ref.read(scanLocalDataSourceProvider);
    final count = await scanLocalDataSource.getPendingCount();
    state = state.copyWith(pendingScansCount: count);
  }

  Future<void> scanCode(String code) async {
    if (_currentSessionId == null) {
      state = state.copyWith(
        errorMessage: 'Tidak ada sesi aktif. Pilih sesi terlebih dahulu.',
      );
      return;
    }

    state = state.copyWith(isLoading: true, errorMessage: null, successMessage: null);

    try {
      final productDataSource = ref.read(productRemoteDataSourceProvider);
      final product = await productDataSource.getProductByCode(code);

      state = state.copyWith(
        isLoading: false,
        scannedProduct: product,
      );
    } on ApiException catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.message,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: 'Gagal memindai kode produk. Silakan coba lagi.',
      );
    }
  }

  Future<void> submitQuantity(int quantity, String inputMethod) async {
    if (_currentSessionId == null || state.scannedProduct == null) {
      return;
    }

    state = state.copyWith(isLoading: true, errorMessage: null, successMessage: null);

    try {
      final product = state.scannedProduct!;

      if (state.isOnline) {
        final stockOpnameDataSource = ref.read(stockOpnameRemoteDataSourceProvider);
        await stockOpnameDataSource.submitScan(
          _currentSessionId!,
          SubmitScanRequest(
            productCode: product.productCode,
            stockPhysical: quantity,
            inputMethod: inputMethod,
            // Kunci sekali pakai: satu perintah user, satu baris di server.
            idempotencyKey: const Uuid().v4(),
          ),
        );

        state = state.copyWith(
          isLoading: false,
          successMessage: 'Data berhasil disimpan',
          scannedProduct: null,
        );
      } else {
        // Simpan ke queue offline. `id` baris antrean menjadi idempotency key,
        // jadi retry pengiriman tidak menambah scan baru.
        final scanLocalDataSource = ref.read(scanLocalDataSourceProvider);
        await scanLocalDataSource.addToQueue(
          sessionId: _currentSessionId!,
          productId: product.id,
          productCode: product.code,
          productName: product.name,
          quantityPhysical: quantity,
          inputMethod: inputMethod,
        );

        await _updatePendingCount();

        state = state.copyWith(
          isLoading: false,
          successMessage: 'Disimpan offline (${state.pendingScansCount} pending)',
          scannedProduct: null,
        );
      }
    } on ApiException catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.message,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: 'Gagal menyimpan data. Silakan coba lagi.',
      );
    }
  }

  void clearProduct() {
    state = state.copyWith(scannedProduct: null, errorMessage: null, successMessage: null);
  }

  void clearMessages() {
    state = state.copyWith(errorMessage: null, successMessage: null);
  }
}

final scannerProvider = StateNotifierProvider<ScannerNotifier, ScannerState>((ref) {
  return ScannerNotifier(ref);
});

// Sync Service Provider
final syncServiceProvider = Provider<SyncService>((ref) {
  final service = SyncService(ref);
  // Buang dari antrean yang benar-benar basi, bukan setiap kali gagal.
  ref.onDispose(service.dispose);
  return service;
});

class SyncService {
  final Ref ref;
  bool _isSyncing = false;
  bool _disposed = false;

  /// Lebih dari ini dan baris dianggap dead-letter: berhenti dicoba otomatis
  /// supaya antrean tidak berputar selamanya pada data yang memang ditolak.
  static const int maxAttempts = 5;

  SyncService(this.ref);

  void dispose() {
    _disposed = true;
  }

  Future<void> syncPendingScans() async {
    // Guard re-entrancy: listener connectivity bisa memicu ini berkali-kali.
    if (_isSyncing || _disposed) return;
    _isSyncing = true;

    try {
      final scanLocalDataSource = ref.read(scanLocalDataSourceProvider);
      final stockOpnameDataSource = ref.read(stockOpnameRemoteDataSourceProvider);

      // Baris yang tertinggal di `syncing` (app ditutup di tengah kirim).
      await scanLocalDataSource.resetStuckSyncing();

      // Hanya `pending` dan `failed` di bawah batas attempt.
      final pendingScans = await scanLocalDataSource.getPendingScans();
      final failedScans = await scanLocalDataSource.getFailedScans();

      for (final scan in [...pendingScans, ...failedScans]) {
        if (_disposed) return;

        // Klaim baris dulu: satu baris hanya diproses satu putaran.
        final claimed = await scanLocalDataSource.claimForSync(scan.id);
        if (!claimed) continue;

        try {
          final request = SubmitScanRequest(
            productCode: scan.productCode,
            stockPhysical: scan.quantityPhysical,
            inputMethod: scan.inputMethod,
            idempotencyKey: scan.idempotencyKey,
          );

          await stockOpnameDataSource.submitScan(scan.sessionId, request);
          await scanLocalDataSource.updateSyncStatus(scan.id, SyncStatus.synced);
        } catch (e) {
          final attempts = scan.retryCount + 1;
          await scanLocalDataSource.updateSyncStatus(
            scan.id,
            attempts >= maxAttempts ? SyncStatus.failed : SyncStatus.pending,
            errorMessage: e.toString(),
          );
          await scanLocalDataSource.incrementRetryCount(scan.id);
        }
      }

      await scanLocalDataSource.deleteSynced();
    } finally {
      _isSyncing = false;
    }
  }
}
// ===== END FILE =====
