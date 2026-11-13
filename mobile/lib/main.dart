// ===== FILE: mobile/lib/main.dart =====
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/theme/app_theme.dart';
import 'features/auth/screens/login_screen.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const ProviderScope(child: StockOpnameApp()));
}

class StockOpnameApp extends StatelessWidget {
  const StockOpnameApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Stock Opname',
      theme: AppTheme.darkTheme(),
      debugShowCheckedModeBanner: false,
      home: const LoginScreen(),
    );
  }
}
// ===== END FILE =====
