import 'package:flutter/material.dart';

import 'pages/nfc_admin_page.dart';

void main() {
  runApp(const NfcAdminApp());
}

class NfcAdminApp extends StatelessWidget {
  const NfcAdminApp({super.key});

  @override
  Widget build(BuildContext context) {
    return const MaterialApp(
      debugShowCheckedModeBanner: false,
      home: NfcAdminPage(),
    );
  }
}
