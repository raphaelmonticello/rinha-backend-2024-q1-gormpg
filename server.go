package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Cliente struct {
	ID         uint        `gorm:"primaryKey"`
	Limite     int         `json:"limite"`
	Saldo      int         `json:"saldo"`
	Transacoes []Transacao `gorm:"foreignKey:ClienteID"`
}

type Transacao struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	ClienteID uint      `json:"-"`
	Valor     int       `json:"valor"`
	Tipo      string    `json:"tipo"`
	Descricao string    `json:"descricao"`
	CreatedAt time.Time `json:"realizada_em"`
}

var db *gorm.DB
var err error

func initDatabase() {
	// Retrieve database connection details from environment variables
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := "disable"
	timeZone := "America/Sao_Paulo"

	// Construct the DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", host, user, password, dbname, port, sslmode, timeZone)

	// Open the database connection
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get database object")
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(1)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(1)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(0)

	// Migrate the schema
	db.AutoMigrate(&Cliente{}, &Transacao{})

	// Seed the database with initial client data
	seedClientes()
}

func seedClientes() {
	var count int64
	db.Model(&Cliente{}).Count(&count)
	if count == 0 {
		clientes := []Cliente{
			{ID: 1, Limite: 100000, Saldo: 0},
			{ID: 2, Limite: 80000, Saldo: 0},
			{ID: 3, Limite: 1000000, Saldo: 0},
			{ID: 4, Limite: 10000000, Saldo: 0},
			{ID: 5, Limite: 500000, Saldo: 0},
		}

		for _, cliente := range clientes {
			var tempCliente Cliente
			if db.First(&tempCliente, cliente.ID).Error == gorm.ErrRecordNotFound {
				db.Create(&cliente) // Create only if not exists
			}
		}
	}
}

func main() {
	initDatabase()

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
	var transacao Transacao
	clienteID, _ := strconv.Atoi(c.Params("id"))

	// Decode request body into Transacao struct
	if err := c.BodyParser(&transacao); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Check if descricao is empty or longer than 10 characters
	if len(transacao.Descricao) == 0 || len(transacao.Descricao) > 10 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "descricao deve ser uma string de 1 a 10 caracteres.")
	}

	// Start a new DB transaction
	tx := db.Begin()

	// Check for the existence of the client
	var cliente Cliente
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cliente, clienteID)
	if result.Error != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cliente not found"})
	}

	// Apply transaction logic based on type (c: credit, d: debit)
	switch transacao.Tipo {
	case "c":
		cliente.Saldo += transacao.Valor
	case "d":
		if cliente.Saldo-transacao.Valor < -cliente.Limite {
			tx.Rollback()
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "Insufficient funds"})
		}
		cliente.Saldo -= transacao.Valor
	default:
		tx.Rollback()
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "Invalid transaction type"})
	}

	// Save transaction
	transacao.ClienteID = uint(clienteID)
	if err := tx.Create(&transacao).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	// Update client saldo
	if err := tx.Save(&cliente).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update cliente saldo"})
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction commit failed"})
	}

	return c.JSON(fiber.Map{
		"limite": cliente.Limite,
		"saldo":  cliente.Saldo,
	})
}

func extratoHandler(c *fiber.Ctx) error {
	clienteID, _ := strconv.Atoi(c.Params("id"))

	// Check for the existence of the client
	var cliente Cliente
	result := db.First(&cliente, clienteID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cliente not found"})
	}

	// Load the last 10 transactions
	var ultimasTransacoes []Transacao
	db.Where("cliente_id = ?", clienteID).Order("created_at desc").Limit(10).Find(&ultimasTransacoes)

	extrato := fiber.Map{
		"saldo": fiber.Map{
			"total":        cliente.Saldo,
			"data_extrato": time.Now().Format(time.RFC3339),
			"limite":       cliente.Limite,
		},
		"ultimas_transacoes": ultimasTransacoes,
	}

	return c.JSON(extrato)
}
