import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;
import 'package:http/io_client.dart';

class ApiClient {
  ApiClient({
    this.baseUrl = 'https://localhost:8888/api/v1',
  }) {
    final httpClient = HttpClient()
      ..badCertificateCallback =
          (X509Certificate cert, String host, int port) {
        return host == 'localhost' || host == '127.0.0.1';
      };

    _client = IOClient(httpClient);
  }

  final String baseUrl;
  late final http.Client _client;

  String? _token;

  Map<String, String> get _headers {
    final headers = <String, String>{
      'Content-Type': 'application/json',
    };

    if (_token != null) {
      headers['Authorization'] = 'Bearer $_token';
    }

    return headers;
  }

  Future<void> loginAsAdmin() async {
    final response = await _client.post(
      Uri.parse('$baseUrl/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'login': 'admin',
        'password': 'admin123',
      }),
    );

    if (response.statusCode != 200) {
      throw Exception(
        'Ошибка авторизации: ${response.statusCode}\n${response.body}',
      );
    }

    final data = jsonDecode(response.body) as Map<String, dynamic>;

    _token = data['token'] as String?;

    if (_token == null || _token!.isEmpty) {
      throw Exception('Backend не вернул JWT-токен');
    }
  }

  Future<void> _ensureLoggedIn() async {
    if (_token == null) {
      await loginAsAdmin();
    }
  }

  Future<List<dynamic>> getCards() async {
    await _ensureLoggedIn();

    final response = await _client.get(
      Uri.parse('$baseUrl/cards'),
      headers: _headers,
    );

    if (response.statusCode != 200) {
      throw Exception(
        'Ошибка получения карт: ${response.statusCode}\n${response.body}',
      );
    }

    final decoded = jsonDecode(response.body);

    if (decoded is List) {
      return decoded;
    }

    if (decoded is Map<String, dynamic> && decoded['cards'] is List) {
      return decoded['cards'] as List<dynamic>;
    }

    throw Exception('Неожиданный формат ответа /cards:\n${response.body}');
  }

  Future<Map<String, dynamic>?> findCardByNumber(String number) async {
    final normalizedNumber = _normalizeCardNumber(number);
    final cards = await getCards();

    for (final item in cards) {
      if (item is! Map<String, dynamic>) {
        continue;
      }

      final cardNumber = _normalizeCardNumber(
        item['number']?.toString() ?? '',
      );

      if (cardNumber == normalizedNumber) {
        return item;
      }
    }

    return null;
  }

  /// Создаёт или обновляет запись в БД по данным с карты.
  Future<Map<String, dynamic>> syncCardToBackend({
    required String number,
    required String ownerName,
    required int balance,
    int userId = 1,
    int keyId = 1,
  }) async {
    await _ensureLoggedIn();

    final normalizedNumber = _normalizeCardNumber(number);
    final existingCard = await findCardByNumber(normalizedNumber);

    if (existingCard != null) {
      return updateCard(
        cardId: _readInt(existingCard['id'], fieldName: 'id'),
        number: normalizedNumber,
        ownerName: ownerName,
        balance: balance,
        blocked: existingCard['blocked'] == true,
        userId: _readOptionalInt(existingCard['user_id']) ?? userId,
        keyId: _readOptionalInt(existingCard['key_id']) ?? keyId,
      );
    }

    return registerCard(
      number: normalizedNumber,
      ownerName: ownerName,
      balance: balance,
      userId: userId,
      keyId: keyId,
    );
  }

  Future<Map<String, dynamic>> registerCard({
    required String number,
    required String ownerName,
    required int balance,
    int userId = 1,
    int keyId = 1,
  }) async {
    await _ensureLoggedIn();

    final normalizedNumber = _normalizeCardNumber(number);

    final existingCard = await findCardByNumber(normalizedNumber);
    if (existingCard != null) {
      return existingCard;
    }

    final response = await _client.post(
      Uri.parse('$baseUrl/cards'),
      headers: _headers,
      body: jsonEncode({
        'number': normalizedNumber,
        'balance': balance,
        'blocked': false,
        'owner_name': ownerName,
        'user_id': userId,
        'key_id': keyId,
      }),
    );

    if (response.statusCode != 200 && response.statusCode != 201) {
      throw Exception(
        'Ошибка регистрации карты: ${response.statusCode}\n${response.body}',
      );
    }

    final decoded = jsonDecode(response.body);

    if (decoded is Map<String, dynamic>) {
      return decoded;
    }

    throw Exception(
      'Неожиданный формат ответа при регистрации карты:\n${response.body}',
    );
  }

  Future<List<dynamic>> getTerminals() async {
    await _ensureLoggedIn();

    final response = await _client.get(
      Uri.parse('$baseUrl/terminals'),
      headers: _headers,
    );

    if (response.statusCode != 200) {
      throw Exception(
        'Ошибка получения терминалов: ${response.statusCode}\n${response.body}',
      );
    }

    final decoded = jsonDecode(response.body);

    if (decoded is List) {
      return decoded;
    }

    throw Exception(
      'Неожиданный формат ответа /terminals:\n${response.body}',
    );
  }

  Future<Map<String, dynamic>?> findTerminalBySerial(String serial) async {
    final normalized = serial.trim();
    final terminals = await getTerminals();

    for (final item in terminals) {
      if (item is! Map<String, dynamic>) {
        continue;
      }

      final terminalSerial = item['serial_number']?.toString().trim() ?? '';

      if (terminalSerial == normalized) {
        return item;
      }
    }

    return null;
  }

  /// Запись пополнения в журнал транзакций (баланс карты в БД не меняет).
  Future<Map<String, dynamic>> recordTopUpTransaction({
    required String cardNumber,
    required int amount,
    required String terminalSerial,
  }) async {
    await _ensureLoggedIn();

    if (amount <= 0) {
      throw Exception('Сумма пополнения должна быть положительной');
    }

    final normalizedNumber = _normalizeCardNumber(cardNumber);
    final card = await findCardByNumber(normalizedNumber);

    if (card == null) {
      throw Exception('Карта $normalizedNumber не найдена в БД');
    }

    final terminal = await findTerminalBySerial(terminalSerial.trim());
    if (terminal == null) {
      throw Exception('Терминал $terminalSerial не найден в БД');
    }

    final response = await _client.post(
      Uri.parse('$baseUrl/transactions'),
      headers: _headers,
      body: jsonEncode({
        'amount': amount,
        'card_id': _readInt(card['id'], fieldName: 'id'),
        'terminal_id': _readInt(terminal['id'], fieldName: 'id'),
        'status': 'approved',
        'message': 'top-up',
      }),
    );

    if (response.statusCode != 200 && response.statusCode != 201) {
      throw Exception(
        'Ошибка записи транзакции пополнения: ${response.statusCode}\n${response.body}',
      );
    }

    final decoded = jsonDecode(response.body);

    if (decoded is Map<String, dynamic>) {
      return decoded;
    }

    throw Exception(
      'Неожиданный формат ответа при создании транзакции:\n${response.body}',
    );
  }

  Future<Map<String, dynamic>> authorizePayment({
    required String cardNumber,
    required int amount,
    required String terminalSerial,
  }) async {
    final normalizedNumber = _normalizeCardNumber(cardNumber);

    final response = await _client.post(
      Uri.parse('$baseUrl/terminal/payments/authorize'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'card_number': normalizedNumber,
        'amount': amount,
        'terminal_serial': terminalSerial,
      }),
    );

    if (response.statusCode != 200) {
      throw Exception(
        'Ошибка транзакции: ${response.statusCode}\n${response.body}',
      );
    }

    final decoded = jsonDecode(response.body);

    if (decoded is Map<String, dynamic>) {
      return decoded;
    }

    throw Exception(
      'Неожиданный формат ответа при транзакции:\n${response.body}',
    );
  }

  Future<Map<String, dynamic>> topUpCard({
    required String cardNumber,
    required int amount,
  }) async {
    await _ensureLoggedIn();

    if (amount <= 0) {
      throw Exception('Сумма пополнения должна быть положительной');
    }

    final normalizedNumber = _normalizeCardNumber(cardNumber);
    final card = await findCardByNumber(normalizedNumber);

    if (card == null) {
      throw Exception('Карта $normalizedNumber не найдена в БД');
    }

    final cardId = _readInt(card['id'], fieldName: 'id');
    final oldBalance = _readInt(card['balance'], fieldName: 'balance');
    final newBalance = oldBalance + amount;

    return updateCard(
      cardId: cardId,
      number: card['number']?.toString() ?? normalizedNumber,
      ownerName: card['owner_name']?.toString() ?? '',
      balance: newBalance,
      blocked: card['blocked'] == true,
      userId: _readOptionalInt(card['user_id']) ?? 1,
      keyId: _readOptionalInt(card['key_id']) ?? 1,
    );
  }

  Future<Map<String, dynamic>> updateCard({
    required int cardId,
    required String number,
    required String ownerName,
    required int balance,
    required bool blocked,
    int userId = 1,
    int keyId = 1,
  }) async {
    await _ensureLoggedIn();

    final response = await _client.put(
      Uri.parse('$baseUrl/cards/$cardId'),
      headers: _headers,
      body: jsonEncode({
        'number': _normalizeCardNumber(number),
        'balance': balance,
        'blocked': blocked,
        'owner_name': ownerName,
        'user_id': userId,
        'key_id': keyId,
      }),
    );

    if (response.statusCode != 200) {
      throw Exception(
        'Ошибка обновления карты: ${response.statusCode}\n${response.body}',
      );
    }

    final decoded = jsonDecode(response.body);

    if (decoded is Map<String, dynamic>) {
      return decoded;
    }

    throw Exception(
      'Неожиданный формат ответа при обновлении карты:\n${response.body}',
    );
  }

  String _normalizeCardNumber(String value) {
    return value.replaceAll(RegExp(r'\s+'), '').toUpperCase();
  }

  int _readInt(dynamic value, {required String fieldName}) {
    final result = _readOptionalInt(value);

    if (result == null) {
      throw Exception('Поле $fieldName не является числом: $value');
    }

    return result;
  }

  int? _readOptionalInt(dynamic value) {
    if (value == null) {
      return null;
    }

    if (value is int) {
      return value;
    }

    if (value is num) {
      return value.toInt();
    }

    return int.tryParse(value.toString());
  }
}
