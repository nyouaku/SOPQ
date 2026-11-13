// ===== FILE: mobile/lib/features/scanner/screens/scanner_screen.dart =====
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:flutter_tts/flutter_tts.dart';
import '../providers/scanner_provider.dart';
import '../widgets/quantity_input_dialog.dart';

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
  bool _isProcessing = false;

  @override
  void initState() {
    super.initState();
    _initTts();
    ref.read(scannerProvider.notifier).setSessionId(widget.sessionId);
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

  void _onCodeDetect(BarcodeCapture capture) async {
    if (_isProcessing) return;

    final code = capture.barcodes.firstOrNull?.rawValue;
    if (code == null || code.isEmpty) return;

    setState(() => _isProcessing = true);

    // Pause scanner while processing
    await _scannerController.stop();

    // Scan kode produk (QR) via provider
    await ref.read(scannerProvider.notifier).scanCode(code);

    setState(() => _isProcessing = false);
  }

  void _showQuantityInput(String productName, int stockSystem) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => QuantityInputDialog(
        productName: productName,
        stockSystem: stockSystem,
        onSubmit: (quantity, inputMethod) async {
          await ref.read(scannerProvider.notifier).submitQuantity(quantity, inputMethod);
        },
      ),
    );
  }

  @override
  void dispose() {
    _scannerController.dispose();
    _tts.stop();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final scannerState = ref.watch(scannerProvider);
    final theme = Theme.of(context);

    ref.listen<ScannerState>(scannerProvider, (previous, next) async {
      // Handle scanned product - speak info and show input dialog
      if (next.scannedProduct != null && previous?.scannedProduct != next.scannedProduct) {
        final product = next.scannedProduct!;

        // Speak product info
        final announcement =
            '${product.name}. Stok sistem ${product.stockSystem ?? 0}';
        await _speak(announcement);

        // Show quantity input dialog
        _showQuantityInput(product.name, product.stockSystem ?? 0);
      }

      // Handle success - restart scanner
      if (next.successMessage != null) {
        if (!context.mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(next.successMessage!),
            backgroundColor: Colors.green,
            duration: const Duration(seconds: 2),
          ),
        );

        // Clear and restart scanner
        ref.read(scannerProvider.notifier).clearMessages();
        await _scannerController.start();
      }

      // Handle error
      if (next.errorMessage != null) {
        if (!context.mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(next.errorMessage!),
            backgroundColor: theme.colorScheme.error,
          ),
        );
        ref.read(scannerProvider.notifier).clearMessages();
        await _scannerController.start();
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
      body: Stack(
        children: [
          // Scanner view
          MobileScanner(
            controller: _scannerController,
            onDetect: _onCodeDetect,
          ),

          // Scanner overlay
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

          // Instructions
          Positioned(
            bottom: 80,
            left: 0,
            right: 0,
            child: Container(
              margin: const EdgeInsets.symmetric(horizontal: 24),
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.black54,
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Text(
                'Arahkan kamera ke QR Code produk',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 16,
                  fontWeight: FontWeight.w500,
                ),
                textAlign: TextAlign.center,
              ),
            ),
          ),
        ],
      ),
      floatingActionButton: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Toggle flash
          FloatingActionButton(
            heroTag: 'flash',
            onPressed: () => _scannerController.toggleTorch(),
            child: const Icon(Icons.flash_on),
          ),
          const SizedBox(height: 12),
          // Switch camera
          FloatingActionButton(
            heroTag: 'camera',
            onPressed: () => _scannerController.switchCamera(),
            child: const Icon(Icons.flip_camera_ios),
          ),
        ],
      ),
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

    // Draw overlay with transparent scan area
    canvas.drawPath(
      Path()
        ..addRect(Rect.fromLTWH(0, 0, size.width, size.height))
        ..addRRect(RRect.fromRectAndRadius(scanArea, const Radius.circular(12)))
        ..fillType = PathFillType.evenOdd,
      paint,
    );

    // Draw scan area border
    final borderPaint = Paint()
      ..color = Colors.green
      ..style = PaintingStyle.stroke
      ..strokeWidth = 3;

    canvas.drawRRect(
      RRect.fromRectAndRadius(scanArea, const Radius.circular(12)),
      borderPaint,
    );

    // Draw corner indicators
    final cornerPaint = Paint()
      ..color = Colors.green
      ..style = PaintingStyle.stroke
      ..strokeWidth = 6;

    const cornerLength = 30.0;

    // Top-left corner
    canvas.drawLine(
      Offset(scanArea.left, scanArea.top),
      Offset(scanArea.left + cornerLength, scanArea.top),
      cornerPaint,
    );
    canvas.drawLine(
      Offset(scanArea.left, scanArea.top),
      Offset(scanArea.left, scanArea.top + cornerLength),
      cornerPaint,
    );

    // Top-right corner
    canvas.drawLine(
      Offset(scanArea.right, scanArea.top),
      Offset(scanArea.right - cornerLength, scanArea.top),
      cornerPaint,
    );
    canvas.drawLine(
      Offset(scanArea.right, scanArea.top),
      Offset(scanArea.right, scanArea.top + cornerLength),
      cornerPaint,
    );

    // Bottom-left corner
    canvas.drawLine(
      Offset(scanArea.left, scanArea.bottom),
      Offset(scanArea.left + cornerLength, scanArea.bottom),
      cornerPaint,
    );
    canvas.drawLine(
      Offset(scanArea.left, scanArea.bottom),
      Offset(scanArea.left, scanArea.bottom - cornerLength),
      cornerPaint,
    );

    // Bottom-right corner
    canvas.drawLine(
      Offset(scanArea.right, scanArea.bottom),
      Offset(scanArea.right - cornerLength, scanArea.bottom),
      cornerPaint,
    );
    canvas.drawLine(
      Offset(scanArea.right, scanArea.bottom),
      Offset(scanArea.right, scanArea.bottom - cornerLength),
      cornerPaint,
    );
  }

  @override
  bool shouldRepaint(CustomPainter oldDelegate) => false;
}
// ===== END FILE =====
