package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type CotacaoDolar struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type database struct {
	db  *sql.DB
	ctx context.Context
}

var global_db_context database

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/goexpert")
	if err != nil {
		println("Erro ao abrir o servidor mysql")
		panic(err)
	}
	defer db.Close()

	ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelDB()

	global_db_context.db = db
	global_db_context.ctx = ctxDB

	criarTabela := `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INT AUTO_INCREMENT PRIMARY KEY,
		code VARCHAR(10),
		codein VARCHAR(10),
		name VARCHAR(50),
		high VARCHAR(20),
		low VARCHAR(20),
		varBid VARCHAR(20),
		pctChange VARCHAR(20),
		bid VARCHAR(20),
		ask VARCHAR(20),
		timestamp VARCHAR(20),
		create_date VARCHAR(30),
		criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(criarTabela)
	if err != nil {
		println("Erro ao criar tabela:", err)
	}

	http.HandleFunc("/cotacao", CotacaoDolarHandler)
	http.ListenAndServe(":8080", nil)
}

func CotacaoDolarHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	select {
	case <-ctx.Done():
		println("Requisição cancelada")

	default:
		if r.URL.Path != "/cotacao" {
			println("Page not found.")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
		if err != nil {
			println("Bad request.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := io.ReadAll(req.Body)
		if err != nil {
			println("Error to read request response.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		var resultMap map[string]CotacaoDolar
		err = json.Unmarshal(result, &resultMap)
		if err != nil {
			println(err)
		}
		cotacao := resultMap["USDBRL"]

		w.Write([]byte(string(cotacao.Bid)))

		err = InserirCotacao(&global_db_context.ctx, global_db_context.db, &cotacao)
		if err != nil {
			println("Erro ao inserir cotação na tabela do banco de dados, ", err)
			panic(err)
		}
		println("Cotação adicionada no banco de dados")
	}
}

func InserirCotacao(ctx *context.Context, db *sql.DB, cotacao *CotacaoDolar) error {
	stmt, err := db.Prepare(`
		INSERT INTO 
		cotacoes(
			code, 
			codein, 
			name, 
			high, 
			low, 
			varBid, 
			pctChange, 
			bid, 
			ask, 
			timestamp, 
			create_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(cotacao.Code,
		cotacao.Codein,
		cotacao.Name,
		cotacao.High,
		cotacao.Low,
		cotacao.VarBid,
		cotacao.PctChange,
		cotacao.Bid,
		cotacao.Ask,
		cotacao.Timestamp,
		cotacao.CreateDate)
	if err != nil {
		return err
	}
	return nil
}
