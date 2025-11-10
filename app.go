package main

import (
	"context"
	"embed"
	"fmt"
	"log"

	"github.com/wangle201210/wachat/backend"
	"github.com/wangle201210/wachat/backend/config"
	"github.com/wangle201210/wachat/backend/model"
	"github.com/wangle201210/wachat/backend/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed bin/*
var binaries embed.FS

// App struct
type App struct {
	ctx           context.Context
	chatAPI       *backend.API
	binaryManager *service.BinaryManager
}

// NewApp creates new App
func NewApp(cfg *config.Config) *App {
	api, err := backend.NewAPI()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize API: %v", err))
	}

	// Create binary manager from config
	binaryManager, err := service.NewBinaryManagerFromConfig(cfg.Binaries, binaries)
	if err != nil {
		log.Printf("Binary manager: %v", err)
	}

	return &App{
		chatAPI:       api,
		binaryManager: binaryManager,
	}
}

// startup is called when app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.chatAPI.SetContext(ctx)

	// Start all embedded binaries
	if a.binaryManager != nil {
		if err := a.binaryManager.StartAll(ctx); err != nil {
			log.Printf("Warning: Failed to start binaries: %v", err)
		}
	}
}

// shutdown is called when app stops
func (a *App) shutdown(ctx context.Context) {
	// Cleanup managed binaries
	if a.binaryManager != nil {
		a.binaryManager.Cleanup()
	}
}

// CreateConversation creates new conversation
func (a *App) CreateConversation(title string) (*model.Conversation, error) {
	return a.chatAPI.CreateConversation(title)
}

// ListConversations returns all conversations
func (a *App) ListConversations() ([]*model.Conversation, error) {
	return a.chatAPI.ListConversations()
}

// GetConversation returns conversation with messages
func (a *App) GetConversation(id string) (*model.Conversation, error) {
	return a.chatAPI.GetConversation(id)
}

// DeleteConversation deletes conversation
func (a *App) DeleteConversation(id string) error {
	return a.chatAPI.DeleteConversation(id)
}

// SendMessageStream streams AI response using eino
func (a *App) SendMessageStream(conversationID, content string) error {
	// Create event callback that emits Wails runtime events
	eventCallback := func(eventName string, data interface{}) {
		runtime.EventsEmit(a.ctx, eventName, data)
	}

	// Delegate to service layer
	return a.chatAPI.SendMessageStream(conversationID, content, eventCallback)
}
