BUFFER = 1024
DECODE = 'utf-8'

class Client:
    def __init__(self, ip, port, sock):
        self._ip = ip
        self._port = port
        self.sock = sock
        self.is_alive = True
        self._recv_buffer = b''

    def close(self):
        self.is_alive = False
        if self.sock is not None:
            self.sock.close()
    
    def send(self, msg):
        if self.sock is None:
            return False
        
        data = msg.encode(DECODE) if isinstance(msg, str) else msg
        total_sent = 0
        while total_sent < len(data):
            sent = self.sock.send(data[total_sent:])
            if sent == 0:
                return False 
            total_sent += sent
        return True
        
    
    def receive_message(self):
        if self.sock is None:
            return None

        while True:
            if b'\n' in self._recv_buffer:
                line, self._recv_buffer = self._recv_buffer.split(b'\n', 1)
                return line.decode(DECODE).rstrip('\r')

            chunk = self.sock.recv(BUFFER)
            if not chunk:
                if self._recv_buffer:
                    line = self._recv_buffer
                    self._recv_buffer = b''
                    return line.decode(DECODE).rstrip('\r')
                return None

            self._recv_buffer += chunk
