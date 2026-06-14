import 'package:flutter/material.dart';

import '../models/card_data.dart';
import '../services/api_client.dart';
import '../services/card_codec.dart';
import '../services/nfc_service.dart';

class NfcAdminPage extends StatefulWidget {
  const NfcAdminPage({super.key});

  @override
  State<NfcAdminPage> createState() => _NfcAdminPageState();
}

class _NfcAdminPageState extends State<NfcAdminPage> {
  final NfcService _nfc = NfcService();
  final ApiClient _api = ApiClient();

  final TextEditingController _ownerController =
      TextEditingController(text: 'Иван Иванов');
  final TextEditingController _balanceController =
      TextEditingController(text: '50000');
  final TextEditingController _amountController =
      TextEditingController(text: '1000');
  final TextEditingController _terminalController =
      TextEditingController(text: 'TERM-001');

  String _status =
      'Готово. Сначала положите карту на PN532, затем нажмите кнопку. '
      'Не убирайте карту до окончания операции.';
  String? _lastUid;

  @override
  void dispose() {
    _ownerController.dispose();
    _balanceController.dispose();
    _amountController.dispose();
    _terminalController.dispose();
    _nfc.dispose();
    super.dispose();
  }

  String _formatChip(CardData card) => '''
UID: ${card.uid}

Данные с карты:
Владелец: ${card.ownerName}
Баланс: ${card.balance}
''';

  Future<void> _readCard() async {
    await _runAction(() async {
      try {
        final chip = await _nfc.readCard();
        setState(() {
          _lastUid = chip.uid;
          _ownerController.text = chip.ownerName;
          _balanceController.text = chip.balance.toString();
          _status = '''
Карта считана.

${_formatChip(chip)}''';
        });
      } on CardNotRegisteredException catch (e) {
        setState(() {
          _lastUid = e.uid;
          _status = '''
UID: ${e.uid}

${e.message}
''';
        });
      }
    });
  }

  Future<void> _registerCard() async {
    await _runAction(() async {
      final ownerName = _ownerController.text.trim();
      final balance = int.tryParse(_balanceController.text.trim());

      if (ownerName.isEmpty) {
        throw Exception('Введите имя владельца');
      }
      if (balance == null || balance < 0) {
        throw Exception('Баланс должен быть неотрицательным числом');
      }

      final chip = await _nfc.registerOnCard(
        ownerName: ownerName,
        balance: balance,
      );

      await _api.syncCardToBackend(
        number: chip.uid,
        ownerName: chip.ownerName,
        balance: chip.balance,
      );

      setState(() {
        _lastUid = chip.uid;
        _status = '''
Карта зарегистрирована.

${_formatChip(chip)}''';
      });
    });
  }

  Future<void> _pay() async {
    await _runAction(() async {
      final amount = int.tryParse(_amountController.text.trim());
      if (amount == null || amount <= 0) {
        throw Exception('Сумма списания должна быть положительным числом');
      }

      final terminalSerial = _terminalController.text.trim();
      if (terminalSerial.isEmpty) {
        throw Exception('Введите serial терминала');
      }

      final chip = await _nfc.payOnCard(amount: amount);

      await _api.authorizePayment(
        cardNumber: chip.uid,
        amount: amount,
        terminalSerial: terminalSerial,
      );

      await _api.syncCardToBackend(
        number: chip.uid,
        ownerName: chip.ownerName,
        balance: chip.balance,
      );

      setState(() {
        _lastUid = chip.uid;
        _status = '''
Списание выполнено.

${_formatChip(chip)}''';
      });
    });
  }

  Future<void> _topUp() async {
    await _runAction(() async {
      final amount = int.tryParse(_amountController.text.trim());
      if (amount == null || amount <= 0) {
        throw Exception('Сумма пополнения должна быть положительным числом');
      }

      final terminalSerial = _terminalController.text.trim();
      if (terminalSerial.isEmpty) {
        throw Exception('Введите serial терминала');
      }

      final chip = await _nfc.topUpOnCard(amount: amount);

      await _api.recordTopUpTransaction(
        cardNumber: chip.uid,
        amount: amount,
        terminalSerial: terminalSerial,
      );

      await _api.syncCardToBackend(
        number: chip.uid,
        ownerName: chip.ownerName,
        balance: chip.balance,
      );

      setState(() {
        _lastUid = chip.uid;
        _status = '''
Пополнение выполнено.

${_formatChip(chip)}''';
      });
    });
  }

  Future<void> _runAction(Future<void> Function() action) async {
    setState(() {
      _status = 'Выполняется';
    });

    try {
      await action();
    } catch (e) {
      setState(() {
        _status = 'Ошибка:\n$e';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final uidText =
        _lastUid == null ? 'UID ещё не считан' : 'Последний UID: $_lastUid';

    return Scaffold(
      appBar: AppBar(
        title: const Text('PN532 — данные на карте'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              width: 380,
              child: SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(uidText),
                    const SizedBox(height: 16),
                    TextField(
                      controller: _ownerController,
                      decoration: const InputDecoration(
                        labelText: 'Имя владельца',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _balanceController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Баланс',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    ElevatedButton(
                      onPressed: _registerCard,
                      child: const Text('Зарегистрировать'),
                    ),
                    const Divider(height: 32),
                    TextField(
                      controller: _amountController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Сумма операции',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _terminalController,
                      decoration: const InputDecoration(
                        labelText: 'Serial терминала',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    ElevatedButton(
                      onPressed: _pay,
                      child: const Text('Списать'),
                    ),
                    const SizedBox(height: 12),
                    ElevatedButton(
                      onPressed: _topUp,
                      child: const Text('Пополнить'),
                    ),
                    const Divider(height: 32),
                    OutlinedButton(
                      onPressed: _readCard,
                      child: const Text('Считать карту'),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(width: 24),
            Expanded(
              child: Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.black26),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: SingleChildScrollView(
                  child: SelectableText(
                    _status,
                    style: const TextStyle(fontSize: 16),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
