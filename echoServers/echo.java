import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.ServerSocket;
import java.net.Socket;

public class TestServer {
    public static void main(String[] args) throws IOException {
        new TestServer().start(new ServerSocket(8080));
    }

    @SuppressWarnings("InfiniteLoopStatement")
    private void start(ServerSocket sock) throws IOException {
        while (true) {
            Socket conn = sock.accept();
            new Thread(() -> {
                try {
                    String message = new BufferedReader(new InputStreamReader(conn.getInputStream())).readLine();
                    new PrintWriter(conn.getOutputStream(), true).println(message);
		    conn.close();
                } catch (IOException e) {
                    System.err.println(e.toString());
                }
            }).start();
        }
    }
}
