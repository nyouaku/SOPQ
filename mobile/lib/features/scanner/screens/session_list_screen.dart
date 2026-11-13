// ===== FILE: mobile/lib/features/scanner/screens/session_list_screen.dart =====
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/network/api_response.dart';
import '../../../core/providers/providers.dart';
import '../../../data/models/stock_opname_model.dart';
import '../../auth/providers/auth_provider.dart';
import '../providers/scanner_provider.dart';
import 'scanner_screen.dart';

final _sessionsProvider = FutureProvider.autoDispose<List<StockOpnameSessionModel>>((ref) async {
  final ds = ref.read(stockOpnameRemoteDataSourceProvider);
  final all = await ds.getSessions();
  // Hanya sesi yang masih bisa menerima scan.
  return all.where((s) => s.status == 'in_progress').toList();
});

class SessionListScreen extends ConsumerWidget {
  const SessionListScreen({super.key});

  Future<void> _openSession(
    BuildContext context,
    WidgetRef ref,
    StockOpnameSessionModel session,
  ) async {
    ref.read(scannerProvider.notifier).setSessionId(session.id);
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ScannerScreen(
          sessionId: session.id,
          sessionCode: session.code,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessions = ref.watch(_sessionsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Pilih Sesi Stock Opname'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(_sessionsProvider),
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: 'Logout',
            onPressed: () async {
              await ref.read(authProvider.notifier).logout();
              if (context.mounted) {
                Navigator.of(context).popUntil((route) => route.isFirst);
              }
            },
          ),
        ],
      ),
      body: sessions.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  e is ApiException ? e.message : 'Gagal memuat sesi',
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () => ref.invalidate(_sessionsProvider),
                  child: const Text('Coba Lagi'),
                ),
              ],
            ),
          ),
        ),
        data: (list) {
          if (list.isEmpty) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Text(
                  'Tidak ada sesi aktif.\nMinta supervisor memulai sesi stock opname.',
                  textAlign: TextAlign.center,
                ),
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(_sessionsProvider),
            child: ListView.separated(
              itemCount: list.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, i) {
                final s = list[i];
                return ListTile(
                  leading: const Icon(Icons.inventory_2_outlined),
                  title: Text(s.code),
                  subtitle: Text(
                    '${s.warehouseName ?? s.warehouseId}'
                    '\nMulai: ${s.startedAt.toLocal()}',
                  ),
                  isThreeLine: true,
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => _openSession(context, ref, s),
                );
              },
            ),
          );
        },
      ),
    );
  }
}
