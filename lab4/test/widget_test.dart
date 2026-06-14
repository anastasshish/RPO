import 'dart:ui';

import 'package:flutter_test/flutter_test.dart';

import 'package:libnfc_flutter_test/main.dart';

void main() {
  testWidgets('App shows main actions', (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1200, 900);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(const NfcAdminApp());

    expect(find.text('Зарегистрировать'), findsOneWidget);
    expect(find.text('Списать'), findsOneWidget);
    expect(find.text('Считать карту'), findsOneWidget);
  });
}
