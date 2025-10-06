package postgres

import (
	"avito-shop/internal/domain/dto"
	"avito-shop/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewPostgres(ctx context.Context, conn string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, username string, passHash []byte) error {
	const op = "storage.Postgres.SaveUser"

	sql, args, err := squirrel.Insert("users").
		Columns("username", "password").
		Values(username, passHash).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, repository.ErrUserAlreadyExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) LoginUser(ctx context.Context, inputType, input string) (string, []byte, error) {
	const op = "storage.Postgres.LoginUser"

	var id string
	var password []byte

	sql, args, err := squirrel.Select("id", "password").
		From("users").
		Where(squirrel.Eq{inputType: input}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.db.QueryRow(ctx, sql, args...).Scan(&id, &password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}

		return "", nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, password, nil
}

func (s *Storage) CheckUsernameIsAvailable(ctx context.Context, username string) (bool, error) {
	const op = "storage.Postgres.CheckUsernameIsAvailable"

	sql, args, err := squirrel.Select("id").
		From("users").
		Where(squirrel.Eq{"username": username}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	err = s.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return true, nil
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return false, nil
}

func (s *Storage) GetUserById(ctx context.Context, userID uuid.UUID) (dto.UserDTO, error) {
	const op = "storage.Postgres.GetUserById"

	var user dto.UserDTO
	sql, args, err := squirrel.Select("id", "username", "coins").
		From("users").
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}

	err = s.db.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.Username, &user.Coins)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.UserDTO{}, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}
		return dto.UserDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) GetUserPurchases(ctx context.Context, userID uuid.UUID) ([]dto.PurchaseDTO, error) {
	const op = "storage.Postgres.GetUserPurchases"

	sql, args, err := squirrel.Select("m.name AS type", "COUNT(*) AS quantity").
		From("purchases p").
		Join("merch_items m ON p.merch_id = m.id").
		Where(squirrel.Eq{"p.user_id": userID}).
		GroupBy("m.name").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var items []dto.PurchaseDTO
	for rows.Next() {
		var item dto.PurchaseDTO
		if err := rows.Scan(&item.Merch, &item.Amount); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *Storage) GetCoinTransactions(ctx context.Context, userID uuid.UUID) (dto.TransactionDTO, error) {
	const op = "storage.Postgres.GetCoinTransactions"

	var received []dto.CoinTransactionDTO
	var sent []dto.CoinTransactionDTO

	inQuery, inArgs, err := squirrel.Select("from_user_id AS from_user", "SUM(amount) AS amount").
		From("coin_transactions").
		Where(squirrel.Eq{"to_user_id": userID}).
		GroupBy("from_user_id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return dto.TransactionDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	rowsIn, err := s.db.Query(ctx, inQuery, inArgs...)
	if err != nil {
		return dto.TransactionDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rowsIn.Close()

	for rowsIn.Next() {
		var ct dto.CoinTransactionDTO
		if err := rowsIn.Scan(&ct.Username, &ct.TotalAmount); err != nil {
			return dto.TransactionDTO{}, fmt.Errorf("%s: %w", op, err)
		}
		received = append(received, ct)
	}

	outQuery, outArgs, err := squirrel.Select("to_user_id AS to_user", "SUM(amount) AS amount").
		From("coin_transactions").
		Where(squirrel.Eq{"from_user_id": userID}).
		GroupBy("to_user_id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return dto.TransactionDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	rowsOut, err := s.db.Query(ctx, outQuery, outArgs...)
	if err != nil {
		return dto.TransactionDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rowsOut.Close()

	for rowsOut.Next() {
		var ct dto.CoinTransactionDTO
		if err := rowsOut.Scan(&ct.Username, &ct.TotalAmount); err != nil {
			return dto.TransactionDTO{}, fmt.Errorf("%s: %w", op, err)
		}
		sent = append(sent, ct)
	}

	return dto.TransactionDTO{
		Received: received,
		Sent:     sent,
	}, nil
}

func (s *Storage) TransferCoins(ctx context.Context, fromUserID, toUserID uuid.UUID, amount int) error {
	const op = "storage.Postgres.TransferCoins"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	deductQuery, deductArgs, err := squirrel.Update("users").
		Set("coins", squirrel.Expr("coins - ?", amount)).
		Where(squirrel.Eq{"id": fromUserID}).
		Where("coins >= ?", amount).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	cmdTag, err := tx.Exec(ctx, deductQuery, deductArgs...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: insufficient funds", op)
	}

	addQuery, addArgs, err := squirrel.Update("users").
		Set("coins", squirrel.Expr("coins + ?", amount)).
		Where(squirrel.Eq{"id": toUserID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = tx.Exec(ctx, addQuery, addArgs...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	insertQuery, insertArgs, err := squirrel.Insert("coin_transactions").
		Columns("from_user_id", "to_user_id", "amount", "created_at").
		Values(fromUserID, toUserID, amount, time.Now()).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = tx.Exec(ctx, insertQuery, insertArgs...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) BuyItem(ctx context.Context, userID uuid.UUID, item string) error {
	const op = "storage.Postgres.BuyItem"

	var merchID uuid.UUID
	var price int

	sqlSelect, argsSelect, err := squirrel.Select("id", "price").
		From("merch_items").
		Where(squirrel.Eq{"name": item}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.db.QueryRow(ctx, sqlSelect, argsSelect...).Scan(&merchID, &price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: item not found", op)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	deductQuery, deductArgs, err := squirrel.Update("users").
		Set("coins", squirrel.Expr("coins - ?", price)).
		Where(squirrel.Eq{"id": userID}).
		Where("coins >= ?", price).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	cmdTag, err := tx.Exec(ctx, deductQuery, deductArgs...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: insufficient funds", op)
	}

	insertQuery, insertArgs, err := squirrel.Insert("purchases").
		Columns("user_id", "merch_id", "price_at_purchase", "created_at").
		Values(userID, merchID, price, time.Now()).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, insertQuery, insertArgs...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}
