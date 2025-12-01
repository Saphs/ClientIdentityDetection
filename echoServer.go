package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type EchoServer struct {
	srvr *http.Server
	tlsAcc *map[net.Conn]*tls.ClientHelloInfo
}

func NewEchoServer() *EchoServer {
	// Server
	handler := http.NewServeMux()
	handler.HandleFunc("/hello", hello)
	handler.HandleFunc("/inspect", inspect)
	handler.HandleFunc("/favicon.ico", favicon)

	var acc = make(map[net.Conn]*tls.ClientHelloInfo)

	// Prepare to get all TLS info via wrapped connection
	tlsConfig := &tls.Config{
		GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			log.Printf("Client ALPN list: %v", chi.SupportedProtos)
			acc[chi.Conn] = chi
			return nil, nil // use default config
		},
	}

	server := &http.Server{
		Addr:      ":8090",
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	return &EchoServer{
		srvr: server,
		tlsAcc: &acc,
	}

}

func (s *EchoServer) start() {
	err := s.srvr.ListenAndServeTLS("server.crt", "server.key")
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func withSelf(s *EchoServer) {}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Server hello =)")
}

const separator = "------------------------------------------------------------------------------------------"
const bigSeparator = "=========================================================================================="

func writeHttpInfo(b *strings.Builder, req *http.Request) {
	b.WriteString("HTTP:\n\n")
	b.WriteString(fmt.Sprintf("%s %s%s %s\n", req.Method, req.Host, req.URL, req.Proto))
	b.WriteString(separator + "\n")
	for key, value := range req.Header {
		var line = fmt.Sprintf("%s: %s", key, strings.Join(value, ";"))
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")
	for key, value := range req.Trailer {
		var line = fmt.Sprintf("%s: %s", key, strings.Join(value, ";"))
		b.WriteString(line + "\n")
	}
	reqBody, _ := io.ReadAll(req.Body)
	b.WriteString(fmt.Sprintf("\nBody: \"%s\" (%d Bytes)\n", reqBody, req.ContentLength))
	b.WriteString(fmt.Sprintf("\nTransferencodings: %v\n", req.TransferEncoding))
	b.WriteString(fmt.Sprintf("Remote: %v\n", req.RemoteAddr))
	b.WriteString(bigSeparator + "\n")
}

func writeTLSInfo(b *strings.Builder, tlsState *tls.ConnectionState) {
	b.WriteString("TLS:\n\n")
	b.WriteString(fmt.Sprintf("%s\n", tls.VersionName(tlsState.Version)))
	b.WriteString(separator + "\n")
	b.WriteString(fmt.Sprintf("HandshakeComplete: %v\n", tlsState.HandshakeComplete))
	b.WriteString(fmt.Sprintf("DidResume: %v\n", tlsState.DidResume))
	b.WriteString(fmt.Sprintf("CipherSuite: %d\n", tlsState.CipherSuite))
	b.WriteString(fmt.Sprintf("NegotiatedProtocol: %s\n", tlsState.NegotiatedProtocol))
	b.WriteString(fmt.Sprintf("ServerName: %s\n", tlsState.ServerName))
	b.WriteString(fmt.Sprintf("PeerCertificates: %v\n", tlsState.PeerCertificates))
	b.WriteString(fmt.Sprintf("VerifiedChains: %v\n", tlsState.VerifiedChains))
	b.WriteString(fmt.Sprintf("SignedCertificateTimestamps: %v\n", tlsState.SignedCertificateTimestamps))
	b.WriteString(fmt.Sprintf("OCSPResponse: %v\n", tlsState.OCSPResponse))
	b.WriteString(fmt.Sprintf("TLSUnique: %v\n", tlsState.TLSUnique))
	b.WriteString("Client Hello:\n")
	b.WriteString(separator + "\n")

	b.WriteString(bigSeparator + "\n")

}

func inspect(w http.ResponseWriter, req *http.Request) {

	var resBody strings.Builder

	writeHttpInfo(&resBody, req)
	resBody.WriteString("\n")
	writeTLSInfo(&resBody, req.TLS)

	// Date, Content-Length, Content-Type (only if Content-Length > 0) are set implicitly by the Write call
	w.WriteHeader(200)
	var i, err = w.Write([]byte(resBody.String()))
	if err != nil {
		fmt.Println("Error writing response:", err)
	} else {
		fmt.Println("Response written: ", i)
	}
}

func favicon(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "resources/favicon/ms-icon-310x310.png")
}
