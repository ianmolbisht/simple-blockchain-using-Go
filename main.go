package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Block struct {
	Pos       int
	Data      BookCheckout
	Timestamp string
	Hash      string
	Prevhash  string
}

type Book struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	PublishDate string `json:"publish_date"`
	ISBN        string `json:"isbn"`
}

type BookCheckout struct {
	BookId       string `json:"bookid"`
	User         string `json:"user"`
	CheckoutDate string `json:"checkout_date"`
	IsGenesis    bool   `json:"is_genesis"`
}

type Blockchain struct {
	Blocks []*Block `json:"blocks"`
}

var BlockChain *Blockchain
const chainFile = "blockchain.json"

func (b *Block) generateHash() {
	bytes, _ := json.Marshal(b.Data)
	data := fmt.Sprintf("%d%s%s%s", b.Pos, b.Timestamp, string(bytes), b.Prevhash)
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func CreateBlock(prevBlock *Block, checkoutitem BookCheckout) *Block {
	block := &Block{}
	block.Pos = prevBlock.Pos + 1
	block.Timestamp = time.Now().Format(time.RFC3339)
	block.Prevhash = prevBlock.Hash
	block.Data = checkoutitem
	block.generateHash()
	return block
}

func (bc *Blockchain) AddBlock(data BookCheckout) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	block := CreateBlock(prevBlock, data)
	if validBlock(block, prevBlock) {
		bc.Blocks = append(bc.Blocks, block)
		saveBlockchain(bc)
	}
}

func validBlock(block, prevBlock *Block) bool {
	if prevBlock.Hash != block.Prevhash {
		return false
	}
	if !block.ValidateHash(block.Hash) {
		return false
	}
	if prevBlock.Pos+1 != block.Pos {
		return false
	}
	return true
}

func (b *Block) ValidateHash(hash string) bool {
	b.generateHash()
	return b.Hash == hash
}

func GenesisBlock() *Block {
	genesis := &Block{
		Pos:       0,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      BookCheckout{IsGenesis: true},
		Prevhash:  "",
	}
	genesis.generateHash()
	return genesis
}

func NewBlockChain() *Blockchain {
	bc := &Blockchain{}
	if fileExists(chainFile) {
		loaded := loadBlockchain()
		if loaded != nil && len(loaded.Blocks) > 0 {
			return loaded
		}
	}
	bc.Blocks = []*Block{GenesisBlock()}
	saveBlockchain(bc)
	return bc
}

func saveBlockchain(bc *Blockchain) {
	tmp := chainFile + ".tmp"

	file, err := os.Create(tmp)
	if err != nil {
		log.Printf("Error creating temp blockchain file: %v", err)
		return
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(bc); err != nil {
		log.Printf("Error encoding blockchain: %v", err)
		file.Close()
		return
	}
	file.Close()

	if _, err := os.Stat(chainFile); err == nil {
		os.Remove(chainFile)
	}

	if err := os.Rename(tmp, chainFile); err != nil {
		log.Printf("Error renaming blockchain file: %v", err)
	}
}

func loadBlockchain() *Blockchain {
	data, err := os.ReadFile(chainFile)
	if err != nil {
		log.Printf("Error reading chain file: %v", err)
		return nil
	}
	var bc Blockchain
	if err := json.Unmarshal(data, &bc); err != nil {
		log.Printf("Error unmarshalling chain: %v", err)
		return nil
	}
	return &bc
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func getBlockChain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.Blocks, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jbytes)
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkoutitem BookCheckout
	if err := json.NewDecoder(r.Body).Decode(&checkoutitem); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Could not decode block: %v", err)
		w.Write([]byte(`{"error":"invalid payload"}`))
		return
	}

	BlockChain.AddBlock(checkoutitem)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "block added",
	})
}

func newBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid book data"})
		return
	}
	h := md5.New()
	io.WriteString(h, book.ISBN+book.PublishDate)
	book.Id = fmt.Sprintf("%x", h.Sum(nil))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	BlockChain = NewBlockChain()
	r := mux.NewRouter()
	r.Use(middlewareCORS)

	r.HandleFunc("/", getBlockChain).Methods("GET", "OPTIONS")
	r.HandleFunc("/", writeBlock).Methods("POST", "OPTIONS")
	r.HandleFunc("/new", newBook).Methods("POST", "OPTIONS")

	log.Println("Listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
