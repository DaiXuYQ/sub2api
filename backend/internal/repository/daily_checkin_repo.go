package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type dailyCheckinRepository struct {
	db *sql.DB
}

func NewDailyCheckinRepository(db *sql.DB, _ *dbent.Client) service.DailyCheckinRepository {
	return &dailyCheckinRepository{db: db}
}

func (r *dailyCheckinRepository) GetByUserAndDate(ctx context.Context, userID int64, date string) (*service.DailyCheckin, error) {
	rows, err := r.queryer(ctx).QueryContext(ctx, `
		SELECT id, user_id, checkin_date::text, reward_amount, redeem_code_id, created_at
		FROM daily_checkins
		WHERE user_id=$1 AND checkin_date=$2`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	item, err := scanDailyCheckinRows(rows)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrDailyCheckinNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r *dailyCheckinRepository) Create(ctx context.Context, item *service.DailyCheckin) error {
	rows, err := r.queryer(ctx).QueryContext(ctx, `
		INSERT INTO daily_checkins (user_id, checkin_date, reward_amount, redeem_code_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, checkin_date::text, reward_amount, redeem_code_id, created_at`,
		item.UserID, item.CheckinDate, item.RewardAmount, nullableInt64(item.RedeemCodeID))
	if err != nil {
		if isUniqueConstraintViolation(err) {
			return service.ErrDailyCheckinAlreadyDone
		}
		return err
	}
	defer rows.Close()
	created, err := scanDailyCheckinRows(rows)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			return service.ErrDailyCheckinAlreadyDone
		}
		return err
	}
	*item = *created
	return nil
}

func (r *dailyCheckinRepository) queryer(ctx context.Context) interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
} {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	return r.db
}

func scanDailyCheckinRows(rows *sql.Rows) (*service.DailyCheckin, error) {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	item, err := scanDailyCheckin(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return item, nil
}

type dailyCheckinScanner interface {
	Scan(dest ...any) error
}

func scanDailyCheckin(row dailyCheckinScanner) (*service.DailyCheckin, error) {
	var item service.DailyCheckin
	var redeemCodeID sql.NullInt64
	var createdAt time.Time
	if err := row.Scan(&item.ID, &item.UserID, &item.CheckinDate, &item.RewardAmount, &redeemCodeID, &createdAt); err != nil {
		return nil, err
	}
	if redeemCodeID.Valid {
		item.RedeemCodeID = &redeemCodeID.Int64
	}
	item.CreatedAt = createdAt
	return &item, nil
}
