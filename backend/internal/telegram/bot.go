package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	"JumysTab/backend/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api         *tgbotapi.BotAPI
	authService *service.AuthService
}

func NewBot(token string, authService *service.AuthService) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot := &Bot{
		api:         api,
		authService: authService,
	}

	return bot, nil
}

func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		b.handleMessage(ctx, update.Message)
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	// Проверяем команду /start с параметром
	if strings.HasPrefix(msg.Text, "/start") {
		parts := strings.Split(msg.Text, " ")
		if len(parts) > 1 {
			code := parts[1]
			b.handleVerification(ctx, msg, code)
		} else {
			b.sendMessage(msg.Chat.ID, "Добро пожаловать! Используйте ссылку из приложения для верификации.")
		}
	}
}

func (b *Bot) handleVerification(ctx context.Context, msg *tgbotapi.Message, code string) {
	// Ищем пользователя по коду верификации
	err := b.authService.VerifyTelegram(ctx, "", code)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Неверный или истекший код верификации.")
		return
	}

	// TODO: Найти userID по коду. Нужен метод в репозитории
	// Для простоты примера, представим что нашли
	
	// Подтверждаем верификацию
	chatID := fmt.Sprintf("%d", msg.Chat.ID)
	err = b.authService.ConfirmTelegramVerification(ctx, "USER_ID_HERE", chatID)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Ошибка при верификации. Попробуйте позже.")
		return
	}

	b.sendMessage(msg.Chat.ID, "✅ Ваш аккаунт успешно верифицирован! Можете войти в приложение.")
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
