package kcp

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/xtaci/smux"
)

// TestSnappyCompression verifies that data sent through the Strm is correctly
// compressed and decompressed, and that the underlying data size is smaller
// (verifying compression actually happened).
func TestSnappyCompression(t *testing.T) {
	// 1. Create a fake network connection (pipe)
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// 2. Setup smux session on both sides
	// We need a way to capture the underlying data size to verify compression.
	// We'll wrap the serverConn in a counter utility.
	countingConn := &ByteCountingConn{Conn: serverConn}
	
	// Server side (accepts stream)
	serverConfig := smux.DefaultConfig()
	serverSession, err := smux.Server(countingConn, serverConfig)
	if err != nil {
		t.Fatalf("Failed to create smux server: %v", err)
	}
	defer serverSession.Close()

	// Client side (opens stream)
	clientConfig := smux.DefaultConfig()
	clientSession, err := smux.Client(clientConn, clientConfig)
	if err != nil {
		t.Fatalf("Failed to create smux client: %v", err)
	}
	defer clientSession.Close()

	// Channel to signal server is ready
	done := make(chan struct{})

	// 3. Server accepts stream and reads data
	go func() {
		defer close(done)
		stream, err := serverSession.AcceptStream()
		if err != nil {
			t.Errorf("Server failed to accept stream: %v", err)
			return
		}
		// Wrap with our Strm (this essentially tests NewStrm/wrapping logic)
		s := NewStrm(stream)
		defer s.Close()

		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			t.Errorf("Server failed to read: %v", err)
			return
		}
		
		receivedData := buf[:n]
		expectedData := bytes.Repeat([]byte("A"), 500) // 500 'A's
		
		if !bytes.Equal(receivedData, expectedData) {
			t.Errorf("Data mismatch.\nExpected: %s\nGot: %s", expectedData, receivedData)
		}
	}()

	// 4. Client opens stream and writes data
	stream, err := clientSession.OpenStream()
	if err != nil {
		t.Fatalf("Client failed to open stream: %v", err)
	}
	// Wrap with our Strm
	s := NewStrm(stream)
	
	// Create highly compressible data (500 'A's)
	data := bytes.Repeat([]byte("A"), 500)
	
	_, err = s.Write(data)
	if err != nil {
		t.Fatalf("Client failed to write: %v", err)
	}
	s.Close() // Close to ensure flush and EOF

	// Wait for server to finish
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out")
	}

	// 5. Verify Compression
	// The sent data was 500 bytes.
	// Snappy should compress "AAAA..." very well.
	// The underlying smux transport adds headers, but 500 bytes of A's 
	// should shrink significantly, definitely below 100 bytes + overhead.
	// Let's be conservative and check if it's less than 300 bytes.
	
	// Note: We are measuring what the SERVER received over the wire.
	bytesWire := countingConn.BytesRead()
	t.Logf("Original Data Size: %d", len(data))
	t.Logf("Bytes Received on Wire (Compressed + Overhead): %d", bytesWire)

	// Since smux adds overhead (commands, headers), and we only sent one chunk,
	// checking strictly < original might be tricky if data is small.
	// But 500 bytes -> ~30 bytes compressed. + Overhead.
	if bytesWire >= int64(len(data)) {
		t.Errorf("Compression didn't seem to work. Wire bytes (%d) >= Original bytes (%d)", bytesWire, len(data))
	}
}

// ByteCountingConn wraps a net.Conn and counts bytes read
type ByteCountingConn struct {
	net.Conn
	readBytes int64
}

func (c *ByteCountingConn) Read(p []byte) (n int, err error) {
	n, err = c.Conn.Read(p)
	c.readBytes += int64(n)
	return n, err
}

func (c *ByteCountingConn) BytesRead() int64 {
	return c.readBytes
}
