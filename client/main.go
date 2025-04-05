package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		println("Requisição cancelada")

	case <-time.After(300 * time.Millisecond):
		println("Tempo esgotado para fazer a requisição")

	default:
		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
		if err != nil {
			println("Erro ao criar requisição")
			panic(err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			println("Erro ao executar requisição", err)
			panic(err)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Erro ao ler o corpo da resposta:", err)
			return
		}
		defer res.Body.Close()
		println("Cotação: ")
		//io.Copy(os.Stdout, res.Body)

		file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Erro ao criar o arquivo:", err)
			return
		}
		defer file.Close()

		// Escreve no arquivo
		_, err = file.WriteString(fmt.Sprintf("Dólar:{%s}\n", string(body)))
		if err != nil {
			fmt.Println("Erro ao escrever no arquivo:", err)
			return
		}

		fmt.Println("Arquivo criado e conteúdo salvo com sucesso!")
	}
}
