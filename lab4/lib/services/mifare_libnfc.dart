import 'dart:io';
import 'dart:typed_data';

import '../models/card_data.dart';
import 'card_codec.dart';


class MifareLibnfc {
  MifareLibnfc({
    this.nativeDir = r'C:\src\libnfc_flutter_test\native\windows',
  });

  final String nativeDir;

  String get _mfclassic => '$nativeDir${Platform.pathSeparator}nfc-mfclassic.exe';
  String get _mfsector1 => '$nativeDir${Platform.pathSeparator}nfc-mfsector1.exe';
  String get _nfcList => '$nativeDir${Platform.pathSeparator}nfc-list.exe';

  static const _sector1DataOffset = CardCodec.block4 * CardCodec.blockSize;
  static const _sector1DataLength = CardCodec.blockSize * 3;

  Future<CardData> readCard() async {
    final dump = await _readDump();
    return CardCodec.decode(dump);
  }

  Future<CardData> registerCard({
    required String ownerName,
    required int balance,
  }) async {
    final dump = await _readDump();
    final patched = CardCodec.patchDump(dump, ownerName, balance);
    await _writeDump(patched);
    return _readBackAfterWrite(expectedBalance: balance);
  }

  Future<CardData> pay({required int amount}) async {
    if (amount <= 0) {
      throw Exception('Сумма должна быть больше 0');
    }
    final dump = await _readDump();
    final current = CardCodec.decode(dump);
    if (current.balance < amount) {
      throw Exception(
        'Недостаточно средств на карте: ${current.balance}, нужно $amount',
      );
    }
    final newBalance = current.balance - amount;
    final patched = CardCodec.patchDump(
      dump,
      current.ownerName,
      newBalance,
    );
    await _writeDump(patched);
    return _readBackAfterWrite(expectedBalance: newBalance);
  }

  Future<CardData> topUp({required int amount}) async {
    if (amount <= 0) {
      throw Exception('Сумма должна быть больше 0');
    }
    final dump = await _readDump();
    final current = CardCodec.decode(dump);
    final newBalance = current.balance + amount;
    final patched = CardCodec.patchDump(
      dump,
      current.ownerName,
      newBalance,
    );
    await _writeDump(patched);
    return _readBackAfterWrite(expectedBalance: newBalance);
  }


  Future<CardData> _readBackAfterWrite({required int expectedBalance}) async {
    await Future<void>.delayed(const Duration(milliseconds: 350));
    final dump = await _readDump();
    final card = CardCodec.decode(dump);
    if (card.balance != expectedBalance) {
      throw Exception(
        'Запись на карту не подтверждена.\n'
        'Ожидалось: $expectedBalance, на карте сейчас: ${card.balance}.\n'
        'Держите карту на антенне и повторите.',
      );
    }
    return card;
  }

  Future<String> readUid() async {
    final result = await Process.run(
      _nfcList,
      [],
      workingDirectory: nativeDir,
    );
    final stdoutText = result.stdout.toString();
    if (result.exitCode != 0) {
      throw Exception(
        'nfc-list.exe: код ${result.exitCode}\n$stdoutText\n${result.stderr}',
      );
    }
    final match =
        RegExp(r'UID \(NFCID1\):\s*([0-9a-fA-F ]+)').firstMatch(stdoutText);
    if (match == null) {
      throw Exception('UID не найден:\n$stdoutText');
    }
    return match.group(1)!.replaceAll(RegExp(r'\s+'), '').toUpperCase();
  }

 
  Future<String> _waitForTag({
    Duration timeout = const Duration(seconds: 12),
  }) async {
    _ensureToolsExist();
    final deadline = DateTime.now().add(timeout);
    Object? lastError;

    while (DateTime.now().isBefore(deadline)) {
      try {
        return await readUid();
      } catch (e) {
        lastError = e;
        await Future<void>.delayed(const Duration(milliseconds: 350));
      }
    }

    throw Exception(
      'Карта не обнаружена за ${timeout.inSeconds} с.\n'
      'Держите MIFARE Classic 1K на антенне 1–2 см, не двигайте до конца операции.\n'
      '$lastError',
    );
  }

  Future<Uint8List> _readDump() async {
    await _waitForTag();
    await Future<void>.delayed(const Duration(milliseconds: 250));

    Object? lastError;
    for (var attempt = 1; attempt <= 6; attempt++) {
      try {
        return await _readDumpOnce();
      } catch (e) {
        lastError = e;
        if (!_isNoTagError(e.toString()) || attempt >= 6) {
          rethrow;
        }
        await Future<void>.delayed(const Duration(milliseconds: 400));
      }
    }
    throw lastError ?? Exception('Не удалось прочитать карту');
  }

  bool get _useSector1Tool => File(_mfsector1).existsSync();

  Future<Uint8List> _readDumpOnce() async {
    _ensureToolsExist();
    if (_useSector1Tool) {
      return _readDumpViaSector1Tool();
    }
    return _readDumpViaMfclassic();
  }

  Future<Uint8List> _readDumpViaSector1Tool() async {
    final uid = await readUid();
    final blocks = await _readSector1Data();
    final dump = Uint8List(CardCodec.dumpSize1k);
    _patchUidInDump(dump, uid);
    dump.setRange(_sector1DataOffset, _sector1DataOffset + _sector1DataLength, blocks);
    return dump;
  }

  Map<String, String> get _sector1Env => {
        ...Platform.environment,
        'NFC_MFCLASSIC_MAX_BLOCK': '7',
      };

  Future<Uint8List> _readDumpViaMfclassic() async {
    final path = await _tempDumpPath();
    try {
      final result = await Process.run(
        _mfclassic,
        ['r', 'a', 'u', path],
        workingDirectory: nativeDir,
        environment: _sector1Env,
      );
      if (result.exitCode != 0) {
        throw Exception(
          'nfc-mfclassic read не удался (код ${result.exitCode}).\n'
          'Ключ A = FFFFFFFFFFFF, карта MIFARE Classic 1K.\n'
          'Если ошибка на блоке 0x0b: закройте приложение, выполните\n'
          '  cmake --build c:\\src\\libnfc-build --target nfc-mfclassic nfc-mfsector1\n'
          'и скопируйте exe в native\\windows, либо build-mfsector1.ps1\n'
          '${_toolOutput(result)}',
        );
      }
      final file = File(path);
      if (!file.existsSync()) {
        throw Exception('Дамп не создан: $path');
      }
      final bytes = await file.readAsBytes();
      if (bytes.length < CardCodec.dumpSize1k) {
        throw Exception(
          'Неверный размер дампа: ${bytes.length} байт (ожидалось ${CardCodec.dumpSize1k})',
        );
      }
      final full = Uint8List(CardCodec.dumpSize1k);
      full.setRange(0, bytes.length.clamp(0, CardCodec.dumpSize1k), bytes);
      return full;
    } finally {
      final f = File(path);
      if (f.existsSync()) {
        try {
          await f.delete();
        } catch (_) {}
      }
    }
  }

  Future<Uint8List> _readSector1Data() async {
    final path = await _tempDumpPath(suffix: '.s1');
    try {
      final result = await Process.run(
        _mfsector1,
        ['r', path],
        workingDirectory: nativeDir,
      );
      if (result.exitCode != 0) {
        throw Exception(
          'nfc-mfsector1 read не удался (код ${result.exitCode}).\n'
          '${_toolOutput(result)}',
        );
      }
      final bytes = await File(path).readAsBytes();
      if (bytes.length != _sector1DataLength) {
        throw Exception(
          'Неверный размер данных сектора 1: ${bytes.length} байт',
        );
      }
      return Uint8List.fromList(bytes);
    } finally {
      final f = File(path);
      if (f.existsSync()) {
        try {
          await f.delete();
        } catch (_) {}
      }
    }
  }

  void _patchUidInDump(Uint8List dump, String uidHex) {
    final normalized = uidHex.replaceAll(RegExp(r'\s+'), '');
    if (normalized.length < 8 || normalized.length.isOdd) {
      return;
    }
    final len = normalized.length > 8 ? 8 : normalized.length;
    for (var i = 0; i < len ~/ 2; i++) {
      dump[i] = int.parse(normalized.substring(i * 2, i * 2 + 2), radix: 16);
    }
  }

  bool _isNoTagError(String message) {
    final lower = message.toLowerCase();
    return lower.contains('no tag') ||
        lower.contains('не обнаружена') ||
        lower.contains('tag was found');
  }

  Future<void> _writeDump(Uint8List dump) async {
    await _waitForTag(timeout: const Duration(seconds: 8));
    await Future<void>.delayed(const Duration(milliseconds: 250));

    Object? lastError;
    for (var attempt = 1; attempt <= 6; attempt++) {
      try {
        await _writeDumpOnce(dump);
        return;
      } catch (e) {
        lastError = e;
        if (!_isNoTagError(e.toString()) || attempt >= 6) {
          rethrow;
        }
        await Future<void>.delayed(const Duration(milliseconds: 400));
      }
    }
    throw lastError ?? Exception('Не удалось записать на карту');
  }

  Future<void> _writeDumpOnce(Uint8List dump) async {
    _ensureToolsExist();
    if (_useSector1Tool) {
      await _writeSector1Data(
        dump.sublist(_sector1DataOffset, _sector1DataOffset + _sector1DataLength),
      );
      return;
    }
    final dataPath = await _tempDumpPath();
    try {
      await File(dataPath).writeAsBytes(dump, flush: true);
      final result = await Process.run(
        _mfclassic,
        ['w', 'A', 'u', dataPath],
        workingDirectory: nativeDir,
        environment: _sector1Env,
      );
      if (result.exitCode != 0) {
        throw Exception(
          'nfc-mfclassic write не удался (код ${result.exitCode}).\n'
          'Соберите nfc-mfsector1.exe для записи только блоков 4–6.\n'
          '${_toolOutput(result)}',
        );
      }
    } finally {
      final f = File(dataPath);
      if (f.existsSync()) {
        try {
          await f.delete();
        } catch (_) {}
      }
    }
  }

  Future<void> _writeSector1Data(Uint8List blocks48) async {
    if (blocks48.length != _sector1DataLength) {
      throw Exception('Нужно 48 байт для блоков 4–6');
    }
    final path = await _tempDumpPath(suffix: '.s1');
    try {
      await File(path).writeAsBytes(blocks48, flush: true);
      final result = await Process.run(
        _mfsector1,
        ['w', path],
        workingDirectory: nativeDir,
      );
      if (result.exitCode != 0) {
        throw Exception(
          'nfc-mfsector1 write не удался (код ${result.exitCode}).\n'
          '${_toolOutput(result)}',
        );
      }
    } finally {
      final f = File(path);
      if (f.existsSync()) {
        try {
          await f.delete();
        } catch (_) {}
      }
    }
  }

  String _toolOutput(ProcessResult result) {
    final out = '${result.stdout}${result.stderr}'.trim();
    if (out.contains('Usage:')) {
      return out.split('Usage:').first.trim();
    }
    return out;
  }

  void _ensureToolsExist() {
    if (!File(_mfclassic).existsSync()) {
      throw Exception(
        'Не найден nfc-mfclassic.exe в $nativeDir\n'
        'Скопируйте из сборки libnfc, например:\n'
        'copy c:\\src\\libnfc-build\\utils\\nfc-mfclassic.exe $nativeDir',
      );
    }
    if (!File(_nfcList).existsSync()) {
      throw Exception('Не найден nfc-list.exe в $nativeDir');
    }
  }

  Future<String> _tempDumpPath({String suffix = '.mfd'}) async {
    final dir = Directory.systemTemp;
    return '${dir.path}${Platform.pathSeparator}nfc_${DateTime.now().microsecondsSinceEpoch}$suffix';
  }
}
