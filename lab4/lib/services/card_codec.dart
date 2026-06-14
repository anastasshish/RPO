import 'dart:convert';
import 'dart:typed_data';

import '../models/card_data.dart';

class CardCodec {
  static const magic = 'RPO3';
  static const block4 = 4;
  static const block5 = 5;
  static const block6 = 6;
  static const blockSize = 16;
  static const dumpSize1k = 1024;

  static int _offset(int block) => block * blockSize;

  static String uidFromDump(Uint8List dump) {
    if (dump.length < blockSize) {
      throw FormatException('Дамп слишком короткий');
    }

    return dump
        .sublist(0, 4)
        .map((b) => b.toRadixString(16).padLeft(2, '0'))
        .join()
        .toUpperCase();
  }

  static CardData decode(Uint8List dump) {
    if (dump.length < _offset(block6) + blockSize) {
      throw FormatException('Нужен дамп MIFARE 1K ($dumpSize1k байт)');
    }

    final uid = uidFromDump(dump);
    final b4 = dump.sublist(_offset(block4), _offset(block4) + blockSize);
    final b5 = dump.sublist(_offset(block5), _offset(block5) + blockSize);
    final b6 = dump.sublist(_offset(block6), _offset(block6) + blockSize);

    if (!_hasMagic(b4)) {
      throw CardNotRegisteredException(
        uid: uid,
        message:
            'Карта не зарегистрирована (в блоке 4 нет $magic).\n'
            'Нажмите «Записать на карту (регистрация)».',
      );
    }

    final balance =
        ByteData.sublistView(Uint8List.fromList(b4.sublist(4, 12)))
            .getInt64(0, Endian.little);

    final nameBytes = <int>[...b5, ...b6];
    final owner = utf8
        .decode(nameBytes, allowMalformed: true)
        .replaceAll('\x00', '')
        .trim();

    return CardData(uid: uid, ownerName: owner, balance: balance);
  }

  static Uint8List encodeBlocks(String ownerName, int balance) {
    final b4 = Uint8List(blockSize);
    b4.setRange(0, 4, ascii.encode(magic));
    final bal = ByteData(8)..setInt64(0, balance, Endian.little);
    b4.setRange(4, 12, bal.buffer.asUint8List());

    final nameBytes = utf8.encode(ownerName);
    final padded = Uint8List(32);
    final len = nameBytes.length > 32 ? 32 : nameBytes.length;
    padded.setRange(0, len, nameBytes.sublist(0, len));

    final out = Uint8List(blockSize * 3);
    out.setRange(0, blockSize, b4);
    out.setRange(blockSize, blockSize * 2, padded.sublist(0, 16));
    out.setRange(blockSize * 2, blockSize * 3, padded.sublist(16, 32));
    return out;
  }

  static bool _hasMagic(Uint8List block4) {
    final expected = ascii.encode(magic);
    for (var i = 0; i < expected.length; i++) {
      if (block4[i] != expected[i]) return false;
    }
    return true;
  }

  static Uint8List patchDump(Uint8List dump, String ownerName, int balance) {
    final copy = Uint8List.fromList(dump);
    final encoded = encodeBlocks(ownerName, balance);
    copy.setRange(_offset(block4), _offset(block4) + encoded.length, encoded);
    return copy;
  }
}


class CardNotRegisteredException implements Exception {
  CardNotRegisteredException({required this.uid, required this.message});
  final String uid;
  final String message;
  @override
  String toString() => message;
}
