import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:stock_opname_mobile/core/theme/app_theme.dart';
import 'package:stock_opname_mobile/features/auth/screens/login_screen.dart';

Widget _wrap(Widget child) => ProviderScope(
      child: MaterialApp(theme: AppTheme.darkTheme(), home: child),
    );

void main() {
  testWidgets('login screen renders form', (tester) async {
    await tester.pumpWidget(_wrap(const LoginScreen()));

    expect(find.text('Stock Opname'), findsOneWidget);
    expect(find.byType(TextFormField), findsNWidgets(2));
    expect(find.text('LOGIN'), findsOneWidget);
  });

  testWidgets('login rejects empty fields', (tester) async {
    await tester.pumpWidget(_wrap(const LoginScreen()));

    await tester.tap(find.text('LOGIN'));
    await tester.pump();

    expect(find.text('Username tidak boleh kosong'), findsOneWidget);
    expect(find.text('Password tidak boleh kosong'), findsOneWidget);
  });
}
