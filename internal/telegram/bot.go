package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// OnActivate is called when user sends /start <token> to the bot
type OnActivate func(ctx context.Context, token string, chatID int64) error

// OnVerify is called when we need to verify OTP for an existing user
// (user sends code directly to the bot)
type OnOTPRequest func(ctx context.Context, chatID int64) (string, error)

type Bot struct {
	api        *tgbotapi.BotAPI
	onActivate OnActivate
}

func NewBot(token string, onActivate OnActivate) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}
	log.Printf("[TG] authorized as @%s", api.Self.UserName)
	return &Bot{api: api, onActivate: onActivate}, nil
}

// BotUsername returns the bot's username (used to build deep links)
func (b *Bot) BotUsername() string {
	return b.api.Self.UserName
}

// SendOTP sends a one-time code to the user's Telegram
func (b *Bot) SendOTP(chatID int64, code string) error {
	text := fmt.Sprintf(
		"🔐 Ваш код для входа в JumysTap:\n\n`%s`\n\nКод действителен 5 минут.",
		code,
	)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("send otp: %w", err)
	}
	return nil
}

// SendWelcome sends a welcome message after successful activation
func (b *Bot) SendWelcome(chatID int64, name string) error {
	text := fmt.Sprintf(
		"✅ Аккаунт активирован!\n\nДобро пожаловать, *%s*!\nТеперь вы можете войти на сайт JumysTap.",
		name,
	)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := b.api.Send(msg)
	return err
}

// StartPolling begins long-polling for updates (blocking — run in goroutine)
func (b *Bot) StartPolling(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Message == nil {
				continue
			}
			go b.handleMessage(ctx, update.Message)
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)
	chatID := msg.Chat.ID

	// /start <token>  — registration activation
	if strings.HasPrefix(text, "/start ") {
		token := strings.TrimPrefix(text, "/start ")
		token = strings.TrimSpace(token)
		if token == "" {
			b.reply(chatID, "❌ Токен не найден. Пожалуйста, используйте ссылку с сайта.")
			return
		}
		if err := b.onActivate(ctx, token, chatID); err != nil {
			log.Printf("[TG] activate error token=%s: %v", token, err)
			b.reply(chatID, "❌ Ссылка недействительна или уже была использована.")
			return
		}
		return
	}

	// /start without token — generic greeting
	if text == "/start" {
		b.reply(chatID, "👋 Привет! Я бот JumysTap.\n\nЧтобы активировать аккаунт, перейдите по ссылке с сайта регистрации.")
		return
	}

	b.reply(chatID, "ℹ️ Используйте ссылку с сайта для активации аккаунта.")
}

func (b *Bot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("[TG] reply error: %v", err)
	}
}
