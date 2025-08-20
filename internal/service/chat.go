package service

import (
	"errors"

	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
	"github.com/RX90/Chat/internal/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type chatService struct {
	repo repo.ChatRepo
}

func newChatService(r repo.ChatRepo) ChatService {
	return &chatService{repo: r}
}

func (s *chatService) CreateMessage(msg *entities.Message) (dto.OutgoingMessage, error) {
	return s.repo.CreateMessage(msg)
}

func (s *chatService) GetMessages() ([]dto.OutgoingMessage, error) {
	return s.repo.GetMessages()
}

func (s *chatService) UpdateMessage(msgID int, userID uuid.UUID, content string) (dto.OutgoingMessage, error) {
	msg, err := s.repo.FindMessageByID(msgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.OutgoingMessage{}, errors.New("message not found")
		}
		return dto.OutgoingMessage{}, err
	}

	if msg.UserID != userID {
		return dto.OutgoingMessage{}, errors.New("cannot update another user's message")
	}

	msg.Content = content

	updatedMsg, err := s.repo.UpdateMessageContent(msg)
	if err != nil {
		return dto.OutgoingMessage{}, err
	}
	return updatedMsg, nil
}

func (s *chatService) DeleteMessage(msgID int, userID uuid.UUID) error {
	msg, err := s.repo.FindMessageByID(msgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("message not found")
		}
		return err
	}

	if msg.UserID != userID {
		return errors.New("cannot delete another user's message")
	}
	return s.repo.DeleteMessageByID(msgID)
}
