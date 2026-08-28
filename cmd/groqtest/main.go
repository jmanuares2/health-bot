package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmanuares2/health-bot/internal/groq"
)

func main() {
	if os.Getenv("GROQ_API_KEY") == "" {
		log.Fatal("GROQ_API_KEY no está seteado")
	}

	messages := []string{
		"almorcé milanesa con puré y coca",
		"hice press de banca 4x8 a 80kg y sentadilla 3x10 a 100kg",
		"corrí 5km en 28 minutos",
		"peso 78.5kg",
		"tomé 1.5 litros de agua",
	}

	parser := groq.NewParser()
	ctx := context.Background()

	for i, msg := range messages {
		if i > 0 {
			fmt.Println("(esperando 30s por rate limit...)")
			time.Sleep(30 * time.Second)
		}
		fmt.Printf("\n--- Mensaje: %q ---\n", msg)
		result, err := parser.Parse(ctx, msg)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
	}
}
