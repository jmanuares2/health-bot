package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	dbpkg "github.com/jmanuares2/health-bot/internal/db"
	"github.com/jmanuares2/health-bot/internal/bot"
	"github.com/jmanuares2/health-bot/internal/groq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Database ---
	pool, err := dbpkg.Connect(ctx)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	// --- Ensure user exists ---
	queries := dbpkg.New(pool)
	userPhone := os.Getenv("USER_PHONE")
	if userPhone == "" {
		userPhone = "self"
	}
	user, err := queries.GetOrCreateUser(ctx, userPhone)
	if err != nil {
		log.Fatalf("get/create user: %v", err)
	}
	log.Printf("Running as user id=%d phone=%s", user.ID, user.Phone)

	// --- Groq parser ---
	parser := groq.NewParser()

	// --- Bot handler ---
	handler := bot.NewHandler(parser, pool, user.ID)

	// --- WhatsApp session store ---
	sessionDir := os.Getenv("WHATSAPP_SESSION_DIR")
	if sessionDir == "" {
		sessionDir = "./session"
	}
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		log.Fatalf("mkdir session: %v", err)
	}

	dbLog := waLog.Stdout("Database", "WARN", true)
	container, err := sqlstore.New("sqlite3", fmt.Sprintf("file:%s/whatsmeow.db?_foreign_keys=on", sessionDir), dbLog)
	if err != nil {
		log.Fatalf("sqlstore: %v", err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		log.Fatalf("get device: %v", err)
	}

	clientLog := waLog.Stdout("Client", "WARN", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// --- Group JID filter ---
	groupJIDStr := os.Getenv("WHATSAPP_GROUP_JID")
	if groupJIDStr == "" {
		log.Fatal("WHATSAPP_GROUP_JID not set")
	}
	groupJID, err := types.ParseJID(groupJIDStr)
	if err != nil {
		log.Fatalf("parse group JID: %v", err)
	}

	// --- Message listener ---
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			// Only handle messages sent by the user themselves
			if !v.Info.IsFromMe {
				return
			}
			// Only in the configured group
			if v.Info.Chat != groupJID {
				return
			}

			text := extractText(v.Message)
			if text == "" {
				return
			}

			log.Printf("Handling message: %q", text)
			reply := handler.Handle(ctx, text)
			if reply == "" {
				return
			}

			_, err := client.SendMessage(ctx, groupJID, &waE2E.Message{
				Conversation: proto.String(reply),
			})
			if err != nil {
				log.Printf("send message error: %v", err)
			}
		}
	})

	// --- Connect or show QR ---
	if client.Store.ID == nil {
		// First time: show QR code
		qrChan, _ := client.GetQRChannel(ctx)
		if err := client.Connect(); err != nil {
			log.Fatalf("connect: %v", err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				log.Printf("QR code: %s", evt.Code)
				// Optionally print QR to terminal using a library
			} else {
				log.Printf("QR event: %s", evt.Event)
			}
		}
	} else {
		if err := client.Connect(); err != nil {
			log.Fatalf("connect: %v", err)
		}
	}

	log.Println("Bot connected. Waiting for messages...")
	<-ctx.Done()
	client.Disconnect()
	log.Println("Bot stopped.")
}

func extractText(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if msg.Conversation != nil {
		return *msg.Conversation
	}
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil {
		return *msg.ExtendedTextMessage.Text
	}
	return ""
}
