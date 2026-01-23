package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode/utf16"
)

// Message represents a single parsed message
type Message struct {
	PeerID uint64 `json:"peer_id"`
	MsgID  int64  `json:"msg_id"`
	Date   int32  `json:"date"`
	FromID uint64 `json:"from_id"`
	Text   string `json:"text"`
}

// Indexer holds the loaded messages
type Indexer struct {
	messages []Message
	mutex    sync.RWMutex
	key      []byte
}

func NewIndexer() *Indexer {
	return &Indexer{
		messages: make([]Message, 0),
	}
}

// readQString reads a Qt string (length + utf16be bytes)
func readQString(r io.Reader) (string, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}

	if length == 0xFFFFFFFF {
		return "", nil
	}
	if length == 0 {
		return "", nil
	}

	// Length is in bytes
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}

	// Convert UTF-16BE bytes to uint16 slice
	u16s := make([]uint16, length/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = binary.BigEndian.Uint16(data[i*2 : i*2+2])
	}

	return string(utf16.Decode(u16s)), nil
}

func (idx *Indexer) LoadKey(dir string) error {
	keyPath := filepath.Join(dir, "key")
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	if len(key) != 32 {
		return fmt.Errorf("invalid key length: %d", len(key))
	}
	idx.key = key
	return nil
}

// ParseDatFile parses a single .dat file
func (idx *Indexer) ParseDatFile(path string) error {
	// Extract PeerID from filename
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	peerIDStr := filename[0 : len(filename)-len(ext)]
	var peerID uint64
	fmt.Sscanf(peerIDStr, "%d", &peerID)

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
    
    if idx.key != nil {
        if len(data) < 12 {
             if len(data) == 0 {
                 return nil
             }
             return fmt.Errorf("file too short for encryption")
        }
        
        block, err := aes.NewCipher(idx.key)
        if err != nil {
            return err
        }
        
        aesgcm, err := cipher.NewGCM(block)
        if err != nil {
            return err
        }
        
        nonce := data[:12]
        ciphertext := data[12:]
        
        plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
        if err != nil {
            return fmt.Errorf("decryption failed: %v", err)
        }
        data = plaintext
    }

	r := bytes.NewReader(data)
	var newMessages []Message

	for {
		var msgID int64
		if err := binary.Read(r, binary.BigEndian, &msgID); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		var date int32
		if err := binary.Read(r, binary.BigEndian, &date); err != nil {
			return err
		}

		var fromID uint64
		if err := binary.Read(r, binary.BigEndian, &fromID); err != nil {
			return err
		}

		text, err := readQString(r)
		if err != nil {
			return err
		}

		newMessages = append(newMessages, Message{
			PeerID: peerID,
			MsgID:  msgID,
			Date:   date,
			FromID: fromID,
			Text:   text,
		})
	}

	idx.mutex.Lock()
	idx.messages = append(idx.messages, newMessages...)
	idx.mutex.Unlock()

	return nil
}

// LoadAll loads all .dat files from directory
func (idx *Indexer) LoadAll(dir string) error {
    // Try to load key first
    if err := idx.LoadKey(dir); err != nil {
        // fmt.Printf("Warning: Could not load key from %s/key: %v. Assuming unencrypted.\n", dir, err)
    }

	idx.mutex.Lock()
	idx.messages = make([]Message, 0)
	idx.mutex.Unlock()

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	count := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".dat" {
			if err := idx.ParseDatFile(filepath.Join(dir, file.Name())); err == nil {
				count++
			}
		}
	}
	fmt.Printf("Loaded %d files.\n", count)
	return nil
}

// Search performs a simple text search
func (idx *Indexer) Search(query string) map[uint64][]Message {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	results := make(map[uint64][]Message)
	query = strings.ToLower(query)
	
	// Tokenize query
	tokens := strings.Fields(query)
	if len(tokens) == 0 {
		return results
	}

	for _, msg := range idx.messages {
		msgTextLower := strings.ToLower(msg.Text)
		match := true
		for _, token := range tokens {
			if !strings.Contains(msgTextLower, token) {
				match = false
				break
			}
		}
		if match {
			results[msg.PeerID] = append(results[msg.PeerID], msg)
		}
	}

	// Sort results within each peer by date desc
	for peerID := range results {
		msgs := results[peerID]
		sort.Slice(msgs, func(i, j int) bool {
			return msgs[i].Date > msgs[j].Date
		})
		results[peerID] = msgs
	}

	return results
}

// GetContext returns context messages around a specific message
// Returns messages before and after the target message within the same peer
func (idx *Indexer) GetContext(peerID uint64, msgID int64, beforeCount, afterCount int) []Message {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	// Find all messages for this peer
	var peerMessages []Message
	for _, msg := range idx.messages {
		if msg.PeerID == peerID {
			peerMessages = append(peerMessages, msg)
		}
	}

	// Sort by date ascending (oldest first)
	sort.Slice(peerMessages, func(i, j int) bool {
		return peerMessages[i].Date < peerMessages[j].Date
	})

	// Find the target message index
	targetIdx := -1
	for i, msg := range peerMessages {
		if msg.MsgID == msgID {
			targetIdx = i
			break
		}
	}

	if targetIdx == -1 {
		return []Message{}
	}

	// Calculate start and end indices
	start := targetIdx - beforeCount
	if start < 0 {
		start = 0
	}
	end := targetIdx + afterCount + 1
	if end > len(peerMessages) {
		end = len(peerMessages)
	}

	return peerMessages[start:end]
}
