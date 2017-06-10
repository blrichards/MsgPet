import 'dart:core';
import 'dart:async';
import 'dart:io';

main() async {
  var server = await ServerSocket.bind("0.0.0.0", 8080);
  server.listen((Socket sock) {
    sock.listen((List<int> data) {
      String message = new String.fromCharCodes(data);
      sock.write(message);
    });
  });
}
