import '../models/card_data.dart';
import 'mifare_libnfc.dart';


  NfcService() : _libnfc = MifareLibnfc();

  final MifareLibnfc _libnfc;

  Future<CardData> readCard() => _libnfc.readCard();

  Future<CardData> registerOnCard({
    required String ownerName,
    required int balance,
  }) =>
      _libnfc.registerCard(ownerName: ownerName, balance: balance);

  Future<CardData> payOnCard({required int amount}) => _libnfc.pay(amount: amount);

  Future<CardData> topUpOnCard({required int amount}) =>
      _libnfc.topUp(amount: amount);

  Future<String> readUid() => _libnfc.readUid();

  void dispose() {}
}
