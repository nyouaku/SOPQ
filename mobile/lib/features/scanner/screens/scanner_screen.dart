// ===== FILE: mobile/lib/features/scanner/screens/scanner_screen.dart =====
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:flutter_tts/flutter_tts.dart';
import '../providers/scanner_provider.dart';
import '../widgets/save_notification.dart';
import '../widgets/session_completion_dialog.dart';

class ScannerScreen extends ConsumerStatefulWidget {
  final String sessionId;
  final String? sessionCode;

  const ScannerScreen({super.key, required this.sessionId, this.sessionCode});

  @override
  ConsumerState<ScannerScreen> createState() => _ScannerScreenState();
}

class _ScannerScreenState extends ConsumerState<ScannerScreen> {
  final MobileScannerController _scannerController = MobileScannerController();
  final FlutterTts _tts = FlutterTts();
  final TextEditingController _quantityController = TextEditingController();
  bool _isProcessing = false;
  bool _isCameraActive = false;

  @override
  void initState() {
    super.initState();
    _initTts();
    ref.read(scannerProvider.notifier).setSessionId(widget.sessionId);

    // Auto-open camera jika ada session
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (widget.sessionId.isNotEmpty) {
        _openCamera();
      }
    });
  }

  Future<void> _initTts() async {
    await _tts.setLanguage('id-ID');
    await _tts.setSpeechRate(0.5);
    await _tts.setVolume(1.0);
    await _tts.setPitch(1.0);
  }

  Future<void> _speak(String text) async {
    await _tts.speak(text);
  }

  Future<void> _openCamera() async {
    if (_isCameraActive) return;
    setState(() => _isCameraActive = true);
    await _scannerController.start();
  }

  Future<void> _closeCamera() async {
    if (!_isCameraActive) return;
    await _scannerController.stop();
    setState(() => _isCameraActive = false);
  }

  void _onCodeDetect(BarcodeCapture capture) async {
    if (_isProcessing) return;

    final code = capture.barcodes.firstOrNull?.rawValue;
    if (code == null || code.isEmpty) return;

    setState(() => _isProcessing = true);
    await _scannerController.stop();

    // Scan kode produk via provider
    await ref.read(scannerProvider.notifier).scanCode(code);

    setState(() => _isProcessing = false);
    _quantityController.clear();
  }

  Future<void> _submitQuantity() async {
    final quantity = int.tryParse(_quantityController.text);
    if (quantity == null || quantity < 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Masukkan jumlah yang valid'), backgroundColor: Colors.red),
      );
      return;
    }

    await ref.read(scannerProvider.notifier).submitQuantity(quantity, 'manual');

    // Clear dan auto-start next scan (tidak perlu dialog)
    _quantityController.clear();
    ref.read(scannerProvider.notifier).clearMessages();
    await _scannerController.start();
  }

  @override
  void dispose() {
    _scannerController.dispose();
    _quantityController.dispose();
    _tts.stop();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final scannerState = ref.watch(scannerProvider);
    final theme = Theme.of(context);

    ref.listen<ScannerState>(scannerProvider, (previous, next) async {
      // Speak product info when scanned
      if (next.scannedProduct != null && previous?.scannedProduct != next.scannedProduct) {
        final product = next.scannedProduct!;
        final announcement = '${product.name}. Stok sistem ${product.stockSystem ?? 0}';
        await _speak(announcement);

        // Close camera untuk fokus input
        await _closeCamera();
      }
    });

    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Scanner'),
            if (widget.sessionCode != null)
              Text(
                widget.sessionCode!,
                style: const TextStyle(fontSize: 12, fontWeight: FontWeight.normal),
              ),
          ],
        ),
        actions: [
          // Connection status badge
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: scannerState.isOnline ? Colors.green : Colors.grey,
                borderRadius: BorderRadius.circular(16),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    scannerState.isOnline ? Icons.cloud_done : Icons.cloud_off,
                    size: 16,
                    color: Colors.white,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    scannerState.isOnline ? 'Online' : 'Offline',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ),
          // Pending scans count
          if (scannerState.pendingScansCount > 0)
            Padding(
              padding: const EdgeInsets.only(right: 16),
              child: Chip(
                avatar: const Icon(Icons.sync, size: 16),
                label: Text('${scannerState.pendingScansCount} pending'),
                backgroundColor: Colors.orange,
                labelStyle: const TextStyle(
                  color: Colors.white,
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
        ],
      ),
      body: Column(
        children: [
          // ===== TOP: Product Info Card + Quantity Input =====
          if (scannerState.scannedProduct != null)
            Container(
              color: Colors.white,
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Product info
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.green.withOpacity(0.1),
                      border: Border.all(color: Colors.green, width: 2),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '✓ Produk Ditemukan',
                          style: const TextStyle(
                            color: Colors.green,
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          scannerState.scannedProduct!.name,
                          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'Stok Sistem: ${scannerState.scannedProduct!.stockSystem ?? 0}',
                          style: TextStyle(fontSize: 14, color: Colors.grey[700]),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Quantity input
                  Text(
                    'Jumlah Fisik',
                    style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _quantityController,
                    keyboardType: TextInputType.number,
                    decoration: InputDecoration(
                      hintText: 'Masukkan jumlah',
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                    ),
                    style: const TextStyle(fontSize: 16),
                  ),
                  const SizedBox(height: 12),

                  // Save button
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: _isProcessing ? null : _submitQuantity,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.blue,
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        disabledBackgroundColor: Colors.grey,
                      ),
                      child: Text(
                        _isProcessing ? 'Menyimpan...' : 'Simpan Data',
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                          color: Colors.white,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            )
          else
            Container(
              color: Colors.blue.withOpacity(0.1),
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Siap Scan',
                          style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Colors.blue),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          _isCameraActive ? 'Arahkan kamera ke QR Code' : 'Aktifkan kamera untuk mulai',
                          style: TextStyle(fontSize: 12, color: Colors.grey[700]),
                        ),
                      ],
                    ),
                  ),
                  if (!_isCameraActive)
                    ElevatedButton.icon(
                      onPressed: _openCamera,
                      icon: const Icon(Icons.videocam),
                      label: const Text('Aktifkan Kamera'),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.blue,
                      ),
                    )
                  else
                    const SizedBox.shrink(),
                ],
              ),
            ),

          // ===== BOTTOM: Camera Preview =====
          if (_isCameraActive)
            Expanded(
              child: Stack(
                children: [
                  MobileScanner(
                    controller: _scannerController,
                    onDetect: _onCodeDetect,
                  ),

                  // Overlay dengan scan area
                  CustomPaint(
                    painter: ScannerOverlay(),
                    child: Container(),
                  ),

                  // Loading indicator
                  if (scannerState.isLoading || _isProcessing)
                    Container(
                      color: Colors.black54,
                      child: const Center(
                        child: CircularProgressIndicator(
                          valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                        ),
                      ),
                    ),
                ],
              ),
            )
          else
            Expanded(
              child: Container(
                color: Colors.grey[200],
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.videocam_off, size: 64, color: Colors.grey[400]),
                      const SizedBox(height: 16),
                      Text(
                        'Kamera tidak aktif',
                        style: TextStyle(fontSize: 16, color: Colors.grey[600]),
                      ),
                    ],
                  ),
                ),
              ),
            ),
        ],
      ),
      floatingActionButton: _isCameraActive
          ? Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                FloatingActionButton(
                  heroTag: 'flash',
                  onPressed: () => _scannerController.toggleTorch(),
                  child: const Icon(Icons.flash_on),
                ),
                const SizedBox(height: 12),
                FloatingActionButton(
                  heroTag: 'camera',
                  onPressed: () => _scannerController.switchCamera(),
                  child: const Icon(Icons.flip_camera_ios),
                ),
              ],
            )
          : null,
    );
  }
}

class ScannerOverlay extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.black54
      ..style = PaintingStyle.fill;

    final scanArea = Rect.fromCenter(
      center: Offset(size.width / 2, size.height / 2),
      width: 250,
      height: 250,
    );

    canvas.drawPath(
      Path()
        ..addRect(Rect.fromLTWH(0, 0, size.width, size.height))
        ..addRRect(RRect.fromRectAndRadius(scanArea, const Radius.circular(12)))
        ..fillType = PathFillType.evenOdd,
      paint,
    );

    final borderPaint = Paint()
      ..color = Colors.green
      ..style = PaintingStyle.stroke
      ..strokeWidth = 3;

    canvas.drawRRect(
      RRect.fromRectAndRadius(scanArea, const Radius.circular(12)),
      borderPaint,
    );
  }

  @override
  bool shouldRepaint(CustomPainter oldDelegate) => false;
}
// ===== END FILE =====
