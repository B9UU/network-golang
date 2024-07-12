package tftp

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"
)

type Server struct {
	Payload []byte        // the payload served for all read requests
	Retries uint8         // number of retires on failed requests
	Timeout time.Duration // waiting duration of an achnowledgment
}

// ListenAndServe takes an addr as argument.
// Start listening to the address in "udp" network
// then calls Serve() method to handle the serve part
func (s Server) ListenAndServe(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	log.Printf("Listening on %s ...\n", conn.LocalAddr())
	return s.Serve(conn)
}

// Serve takes net.PacketConn as argument
// checks if conn, payload are valid
// then checks and sets the retries and timeout to appropriate values.
// Start listening to the network for connections,
// reads from a conn into a buffer with Datagram size.
// Checks the payload and passes the rrq and addr to handle method
func (s *Server) Serve(conn net.PacketConn) error {
	if conn == nil {
		return errors.New("nil conn")
	}
	if s.Payload == nil {
		return errors.New("payload is required")
	}
	if s.Retries == 0 {
		s.Retries = 10
	}
	if s.Timeout == 0 {
		s.Timeout = time.Second * 6
	}
	var rrq ReadReq

	for {
		buf := make([]byte, DatagramSize)
		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}
		err = rrq.UnmarshalBinary(buf)
		if err != nil {
			log.Printf("[%s] bad request: %v", addr, err)
			continue
		}
		go s.handle(addr.String(), rrq)
	}
}

func (s *Server) handle(clientAddr string, rrq ReadReq) {
	log.Printf("[%s] requested file %s", clientAddr, rrq.Filename)
	// connect to the address in udp network
	conn, err := net.Dial("udp", clientAddr)
	if err != nil {
		log.Printf("[%s] dial %v", clientAddr, err)
		return
	}
	// defer connection closing
	defer func() { _ = conn.Close() }()

	// initiating objects we will need
	var (
		ackPkt  Ack
		errPkt  ErrReq
		dataPkt = Data{Payload: bytes.NewReader(s.Payload)}
		buf     = make([]byte, DatagramSize)
	)
	// a label for continue to label since we are doing nested loops
NEXTPACKET:
	// loop until n != datagramsize
	// we keep sending bytes from the file/data until we send less than 516
	// then this will be false
	for n := DatagramSize; n == DatagramSize; {
		// preparing the packet before sending it
		data, err := dataPkt.MarshalBinary()
		if err != nil {
			log.Printf("[%s] preparing --data packet: %v", clientAddr, err)
			return
		}
		// a label for continue to label since we are doing nested loops
	RETRY:
		for i := s.Retries; i > 0; i-- {
			// writing the data from buffer to the connection
			n, err = conn.Write(data) // send the data packet
			if err != nil {
				log.Printf("[%s] write: %v", clientAddr, err)
				return
			}
			// block until we receive a message
			_ = conn.SetReadDeadline(time.Now().Add(s.Timeout))
			// read the message from the connection to the buffer
			_, err = conn.Read(buf)
			if err != nil {
				// checking if the err is a timeout error and retry else we log and return
				if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
					continue RETRY // goto label
				}
				log.Printf("[%s] waiting for ACK: %v", clientAddr, err)
				return
			}
			switch {
			// checking if the message we received is Acknowledgment message
			case ackPkt.UnmarshalBinary(buf) == nil:
				// Check if the block number == the ackpkt
				if uint16(ackPkt) == dataPkt.Block {
					continue NEXTPACKET // goto label
				}
			// checking if the ack is an error and return
			case errPkt.UnmarshalBinary(buf) == nil:
				log.Printf("[%s] received error: %v", clientAddr, errPkt.Message)
				return
			default:
				log.Printf("[%s] bad packet", clientAddr)
			}
		}
		log.Printf("[%s] exhausted retries", clientAddr)
		return
	}
	log.Printf("[%s] sent %d blocks", clientAddr, dataPkt.Block)
}
