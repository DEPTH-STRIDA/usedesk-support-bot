package cache

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// User модель пользователя в БД
type User struct {
	TelegramID    int64 `gorm:"primaryKey"`
	CurrentTicket int64
	ActivTickets  []ActiveTicket `gorm:"foreignKey:TelegramID"`
}

// ActiveTicket модель для хранения активных тикетов
type ActiveTicket struct {
	TicketID   int64     `gorm:"primaryKey"`
	TelegramID int64     `gorm:"index"`
	ExpiretAt  time.Time `gorm:"index"`
}

type TicketCache struct {
	db              *gorm.DB
	ticketTTL       time.Duration
	cleanupInterval time.Duration
}

func NewTicketCache(db *gorm.DB) *TicketCache {
	// Автомиграция
	err := db.AutoMigrate(&User{}, &ActiveTicket{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	c := &TicketCache{
		db:              db,
		ticketTTL:       30 * 24 * time.Hour, // 30 дней
		cleanupInterval: 6 * time.Hour,       // 6 часов
	}

	// Выполняем начальную очистку при запуске
	c.cleanExpiredTickets()

	return c
}

// StartCleaning запускает периодическую очистку устаревших тикетов
func (c *TicketCache) StartCleaning() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanExpiredTickets()
		}
	}
}

// cleanExpiredTickets очищает просроченные тикеты
func (c *TicketCache) cleanExpiredTickets() {
	now := time.Now()

	// Находим всех пользователей с просроченными текущими тикетами
	var usersWithExpiredTickets []User
	c.db.Joins("JOIN active_tickets ON active_tickets.ticket_id = users.current_ticket").
		Where("active_tickets.expiret_at < ?", now).
		Find(&usersWithExpiredTickets)

	// Обнуляем current_ticket у этих пользователей
	for _, user := range usersWithExpiredTickets {
		c.db.Model(&User{}).
			Where("telegram_id = ?", user.TelegramID).
			Update("current_ticket", 0)
	}

	// Удаляем все просроченные тикеты
	c.db.Where("expiret_at < ?", now).Delete(&ActiveTicket{})
}

// GetCurrentTicketIDByTgId получить тикет ID текущего диалога по тг id
func (c *TicketCache) GetCurrentTicketIDByTgId(telegramID int64) (int64, bool) {
	var user User
	result := c.db.First(&user, telegramID)
	if result.Error != nil {
		log.Printf("Пользователь с telegramID %d не найден", telegramID)
		return 0, false
	}

	if user.CurrentTicket == 0 {
		return 0, false
	}

	// Проверяем, не истек ли текущий тикет
	var activeTicket ActiveTicket
	result = c.db.Where("ticket_id = ?", user.CurrentTicket).First(&activeTicket)
	if result.Error != nil || time.Now().After(activeTicket.ExpiretAt) {
		// Если тикет истек или не найден, обнуляем current_ticket
		c.db.Model(&user).Update("current_ticket", 0)
		return 0, false
	}

	log.Printf("Получен текущий тикет %d для пользователя %d", user.CurrentTicket, telegramID)
	return user.CurrentTicket, true
}

// GetTelegramByAnyTicket возвращает телеграмм id пользователя по любому тикету
func (c *TicketCache) GetTelegramByAnyTicket(ticketID int64) (int64, bool) {
	now := time.Now()
	var user User

	// Проверяем current_ticket
	result := c.db.Where("current_ticket = ?", ticketID).First(&user)
	if result.Error == nil {
		// Проверяем, не истек ли тикет
		var activeTicket ActiveTicket
		result = c.db.Where("ticket_id = ?", ticketID).First(&activeTicket)
		if result.Error == nil && now.Before(activeTicket.ExpiretAt) {
			log.Printf("Найден текущий тикет %d для пользователя %d", ticketID, user.TelegramID)
			return user.TelegramID, true
		}
	}

	// Проверяем active_tickets
	var activeTicket ActiveTicket
	result = c.db.Where("ticket_id = ? AND expiret_at > ?", ticketID, now).First(&activeTicket)
	if result.Error == nil {
		log.Printf("Найден активный тикет %d для пользователя %d", ticketID, activeTicket.TelegramID)
		return activeTicket.TelegramID, true
	}

	return 0, false
}

// SaveTicket сохраняет тикет для пользователя и устанавливает его текущим
func (c *TicketCache) SaveTicket(telegramID, ticketID int64) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		var user User
		result := tx.First(&user, telegramID)

		if result.Error == gorm.ErrRecordNotFound {
			// Создаем нового пользователя
			user = User{
				TelegramID:    telegramID,
				CurrentTicket: ticketID,
			}
			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("ошибка при создании пользователя: %v", err)
			}
		} else if result.Error != nil {
			return fmt.Errorf("ошибка при поиске пользователя: %v", result.Error)
		} else {
			// Если есть текущий тикет и он не совпадает с новым
			if user.CurrentTicket != 0 && user.CurrentTicket != ticketID {
				// Проверяем, существует ли уже такой тикет в active_tickets
				var existingTicket ActiveTicket
				checkResult := tx.Where("ticket_id = ?", user.CurrentTicket).First(&existingTicket)

				// Если тикет не существует в active_tickets, добавляем его
				if checkResult.Error == gorm.ErrRecordNotFound {
					oldActiveTicket := ActiveTicket{
						TicketID:   user.CurrentTicket,
						TelegramID: telegramID,
						ExpiretAt:  time.Now().Add(c.ticketTTL),
					}
					if err := tx.Create(&oldActiveTicket).Error; err != nil {
						return fmt.Errorf("ошибка при сохранении старого тикета: %v", err)
					}
				}
			}

			// Обновляем текущий тикет
			if err := tx.Model(&user).Update("current_ticket", ticketID).Error; err != nil {
				return fmt.Errorf("ошибка при обновлении текущего тикета: %v", err)
			}
		}

		// Проверяем существование нового тикета в active_tickets
		var existingNewTicket ActiveTicket
		checkResult := tx.Where("ticket_id = ?", ticketID).First(&existingNewTicket)

		// Создаем запись в active_tickets только если тикет не существует
		if checkResult.Error == gorm.ErrRecordNotFound {
			newActiveTicket := ActiveTicket{
				TicketID:   ticketID,
				TelegramID: telegramID,
				ExpiretAt:  time.Now().Add(c.ticketTTL),
			}
			if err := tx.Create(&newActiveTicket).Error; err != nil {
				return fmt.Errorf("ошибка при создании нового тикета: %v", err)
			}
		}

		log.Printf("Тикет %d сохранен для пользователя %d", ticketID, telegramID)
		return nil
	})
}

// DeleteTicket удаляет тикет из текущего или активных
func (c *TicketCache) DeleteTicket(ticketID int64) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		// Проверяем current_ticket
		var user User
		result := tx.Where("current_ticket = ?", ticketID).First(&user)
		if result.Error == nil {
			log.Printf("Удаление текущего тикета %d для пользователя %d", ticketID, user.TelegramID)
			user.CurrentTicket = 0
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
		}

		// Удаляем из active_tickets
		result = tx.Where("ticket_id = ?", ticketID).Delete(&ActiveTicket{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 && user.CurrentTicket != ticketID {
			return fmt.Errorf("тикет %d не найден", ticketID)
		}

		return nil
	})
}
