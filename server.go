package main

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Transaction struct {
	Valor       int       `json:"valor"`
	Tipo        string    `json:"tipo"`
	Descricao   string    `json:"descricao"`
	RealizadaEm time.Time `json:"realizada_em"`
}

type Cliente struct {
	ID         int
	Limite     int `json:"limite"`
	Saldo      int `json:"saldo"`
	Transacoes []Transaction
}

var clientes = map[int]*Cliente{
	1: {ID: 1, Limite: 100000, Saldo: 0, Transacoes: []Transaction{}},
	2: {ID: 2, Limite: 80000, Saldo: 0, Transacoes: []Transaction{}},
	3: {ID: 3, Limite: 1000000, Saldo: 0, Transacoes: []Transaction{}},
	4: {ID: 4, Limite: 10000000, Saldo: 0, Transacoes: []Transaction{}},
	5: {ID: 5, Limite: 500000, Saldo: 0, Transacoes: []Transaction{}},
}

func main() {
	// Get the port number from the environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default port if not specified
	}
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Server-ID", os.Getenv("SERVER_ID"))
		return c.Next()
	})

	app.Post("/clientes/:id/transacoes", transacaoHandler)
	app.Get("/clientes/:id/extrato", extratoHandler)

	// Use the port variable in the Listen method
	err := app.Listen(":" + port)
	if err != nil {
		panic(err)
	}
}

func transacaoHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id == 6 {
		return c.SendStatus(fiber.StatusNotFound)
	}

	cliente, exists := clientes[id]
	if !exists {
		return c.SendStatus(fiber.StatusNotFound)
	}

	var transacao Transaction
	if err := c.BodyParser(&transacao); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse transaction"})
	}

	// Validate the transaction
	if transacao.Tipo != "c" && transacao.Tipo != "d" ||
		transacao.Descricao == "" || len(transacao.Descricao) > 10 ||
		transacao.Valor <= 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Apply transaction
	novoSaldo := cliente.Saldo
	if transacao.Tipo == "c" {
		novoSaldo += transacao.Valor
	} else {
		novoSaldo -= transacao.Valor
		if novoSaldo < -cliente.Limite {
			return c.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}

	cliente.Saldo = novoSaldo
	transacao.RealizadaEm = time.Now()
	cliente.Transacoes = append(cliente.Transacoes, transacao)

	return c.JSON(fiber.Map{"limite": cliente.Limite, "saldo": cliente.Saldo})
}

func extratoHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id == 6 {
		return c.SendStatus(fiber.StatusNotFound)
	}

	cliente, exists := clientes[id]
	if !exists {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.JSON(fiber.Map{
		"saldo": fiber.Map{
			"total":        cliente.Saldo,
			"data_extrato": time.Now().Format(time.RFC3339),
			"limite":       cliente.Limite,
		},
		"ultimas_transacoes": cliente.Transacoes,
	})
}
