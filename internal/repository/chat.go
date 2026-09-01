package repository

import (
	"context"
	"errors"

	"messages/internal/apperr"
	"messages/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

func (r *ChatRepository) CreateChat(ctx context.Context, ownerID int, name string) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	sql := `INSERT INTO chats(name, owner_id)
			VALUES ($1, $2)
			RETURNING id`

	var id int

	err = tx.QueryRow(ctx, sql, name, ownerID).Scan(&id)
	if err != nil {
		return 0, err
	}

	sql = `	INSERT INTO chat_members(chat_id, user_id)
			VALUES ($1, $2) `

	_, err = tx.Exec(ctx, sql, id, ownerID)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *ChatRepository) AddChatMember(ctx context.Context, chatID, userID, requesterID int) error {
	sql := `INSERT INTO chat_members(chat_id, user_id)
			SELECT $1, $2
			WHERE EXISTS (SELECT 1 FROM chats WHERE id = $1 AND owner_id = $3)`

	tag, err := r.pool.Exec(ctx, sql, chatID, userID, requesterID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return apperr.ErrAlreadyMember
			case "23503":
				return apperr.ErrUserNotFound
			}
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotOwner
	}

	return nil
}

func (r *ChatRepository) RemoveChatMember(ctx context.Context, chatID, userID, requesterID int) error {
	sql := `DELETE FROM chat_members
			WHERE chat_id = $1 AND user_id = $2 AND EXISTS (SELECT 1 FROM chats WHERE owner_id = $3 AND id = $1)`

	tag, err := r.pool.Exec(ctx, sql, chatID, userID, requesterID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotOwner
	}
	return nil
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userID int) ([]models.Chat, error) {
	sql := `SELECT c.id, c.name, c.owner_id
			FROM chats c
			JOIN chat_members cm ON cm.chat_id = c.id
			WHERE cm.user_id = $1`

	rows, err := r.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := []models.Chat{}

	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.ID, &chat.Name, &chat.OwnerID); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *ChatRepository) GetChatMessages(ctx context.Context, chatID, requesterID int, f models.MessageFilter) ([]models.Message, error) {
	sql := `SELECT id, chat_id, user_id, text, created_at, edited_at
	FROM messages
	WHERE chat_id = $1
  		AND EXISTS (SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2)
    	AND ($3::int IS NULL OR id < $3)
     	AND ($4::int IS NULL OR user_id = $4)
	ORDER BY id DESC
	LIMIT $5`

	rows, err := r.pool.Query(ctx, sql, chatID, requesterID, f.BeforeID, f.UserID, f.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []models.Message{}

	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.UserID, &message.Text, &message.CreatedAt, &message.EditedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *ChatRepository) GetChatMembers(ctx context.Context, chatID, requesterID int) ([]models.User, error) {
	sql := `SELECT u.id, u.username FROM users u
			JOIN chat_members cm ON u.id = cm.user_id
			WHERE cm.chat_id = $1
			AND EXISTS (SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2 )`

	rows, err := r.pool.Query(ctx, sql, chatID, requesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *ChatRepository) CreateMessage(ctx context.Context, chatID, requesterID int, text string) (models.Message, error) {
	sql := `INSERT INTO messages(chat_id, user_id, text)
			SELECT $1, $2, $3
			WHERE EXISTS (SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2)
			RETURNING id, chat_id, user_id, text, created_at, edited_at`

	var msg models.Message
	err := r.pool.QueryRow(ctx, sql, chatID, requesterID, text).Scan(&msg.ID, &msg.ChatID, &msg.UserID, &msg.Text, &msg.CreatedAt, &msg.EditedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return msg, apperr.ErrNotChatMember
		}
		return msg, err
	}

	return msg, nil
}

func (r *ChatRepository) UpdateMessage(ctx context.Context, newText string, messageID, chatID, requesterID int) (models.Message, error) {
	sql := `UPDATE messages
			SET text = $1, edited_at = now()
			WHERE id = $2 AND chat_id = $3 AND user_id = $4
			RETURNING id, chat_id, user_id, text, created_at, edited_at`

	var msg models.Message
	err := r.pool.QueryRow(ctx, sql, newText, messageID, chatID, requesterID).Scan(&msg.ID, &msg.ChatID, &msg.UserID, &msg.Text, &msg.CreatedAt, &msg.EditedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return msg, apperr.ErrNotChatMember
		}
		return msg, err
	}

	return msg, nil
}

func (r *ChatRepository) DeleteMessage(ctx context.Context, messageID, chatID, requesterID int) error {
	sql := `DELETE FROM messages
		WHERE id = $1 AND chat_id = $2 AND user_id = $3` // TODO: Дописать, что бы владелец чата мог удалять

	tag, err := r.pool.Exec(ctx, sql, messageID, chatID, requesterID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrMessageNotFoundOrForbidden
	}

	return nil
}
