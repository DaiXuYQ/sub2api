package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	DailyCheckinRewardMin = 0.5
	DailyCheckinRewardMax = 5.0
)

var (
	ErrDailyCheckinNotFound    = infraerrors.NotFound("DAILY_CHECKIN_NOT_FOUND", "daily check-in record not found")
	ErrDailyCheckinAlreadyDone = infraerrors.Conflict("DAILY_CHECKIN_ALREADY_DONE", "daily check-in already completed")
)

type DailyCheckin struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	CheckinDate  string    `json:"checkin_date"`
	RewardAmount float64   `json:"reward_amount"`
	RedeemCodeID *int64    `json:"redeem_code_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type DailyCheckinStatus struct {
	CheckedInToday bool          `json:"checked_in_today"`
	Today          string        `json:"today"`
	RewardMin      float64       `json:"reward_min"`
	RewardMax      float64       `json:"reward_max"`
	Checkin        *DailyCheckin `json:"checkin,omitempty"`
}

type DailyCheckinResult struct {
	CheckedInToday bool          `json:"checked_in_today"`
	RewardAmount   float64       `json:"reward_amount"`
	Balance        float64       `json:"balance"`
	Checkin        *DailyCheckin `json:"checkin"`
}

type DailyCheckinRepository interface {
	GetByUserAndDate(ctx context.Context, userID int64, date string) (*DailyCheckin, error)
	Create(ctx context.Context, item *DailyCheckin) error
}

type DailyCheckinService struct {
	repo                 DailyCheckinRepository
	userRepo             UserRepository
	redeemRepo           RedeemCodeRepository
	entClient            *dbent.Client
	authCacheInvalidator TokenCacheInvalidator
	billingCacheService  *BillingCacheService
}

func NewDailyCheckinService(repo DailyCheckinRepository, userRepo UserRepository, redeemRepo RedeemCodeRepository, entClient *dbent.Client, authCacheInvalidator TokenCacheInvalidator, billingCacheService *BillingCacheService) *DailyCheckinService {
	return &DailyCheckinService{repo: repo, userRepo: userRepo, redeemRepo: redeemRepo, entClient: entClient, authCacheInvalidator: authCacheInvalidator, billingCacheService: billingCacheService}
}

func (s *DailyCheckinService) Status(ctx context.Context, userID int64) (*DailyCheckinStatus, error) {
	if userID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	today := checkinDate(time.Now())
	item, err := s.repo.GetByUserAndDate(ctx, userID, today)
	if err != nil && err != ErrDailyCheckinNotFound {
		return nil, fmt.Errorf("get daily checkin: %w", err)
	}
	return &DailyCheckinStatus{
		CheckedInToday: item != nil,
		Today:          today,
		RewardMin:      DailyCheckinRewardMin,
		RewardMax:      DailyCheckinRewardMax,
		Checkin:        item,
	}, nil
}

func (s *DailyCheckinService) Checkin(ctx context.Context, userID int64) (*DailyCheckinResult, error) {
	if userID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}

	today := checkinDate(time.Now())
	if existing, err := s.repo.GetByUserAndDate(ctx, userID, today); err == nil && existing != nil {
		return nil, ErrDailyCheckinAlreadyDone
	} else if err != nil && err != ErrDailyCheckinNotFound {
		return nil, fmt.Errorf("get daily checkin: %w", err)
	}

	reward, err := randomCheckinReward()
	if err != nil {
		return nil, fmt.Errorf("generate checkin reward: %w", err)
	}
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	code, err := GenerateRedeemCode()
	if err != nil {
		return nil, fmt.Errorf("generate reward code: %w", err)
	}
	redeemRecord := &RedeemCode{
		Code:   fmt.Sprintf("CHECKIN-%s-%d-%s", todayCompact(today), userID, code[:8]),
		Type:   RedeemTypeBalance,
		Value:  reward,
		Status: StatusUsed,
		UsedBy: &userID,
		UsedAt: &now,
		Notes:  "每日签到奖励",
	}
	if err := s.redeemRepo.Create(txCtx, redeemRecord); err != nil {
		return nil, fmt.Errorf("create checkin reward record: %w", err)
	}

	item := &DailyCheckin{UserID: userID, CheckinDate: today, RewardAmount: reward, RedeemCodeID: &redeemRecord.ID, CreatedAt: now}
	if err := s.repo.Create(txCtx, item); err != nil {
		if err == ErrDailyCheckinAlreadyDone {
			return nil, ErrDailyCheckinAlreadyDone
		}
		return nil, fmt.Errorf("create daily checkin: %w", err)
	}

	if err := s.userRepo.UpdateBalance(txCtx, userID, reward); err != nil {
		return nil, fmt.Errorf("update user balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit daily checkin: %w", err)
	}

	s.invalidateBalanceCaches(ctx, userID)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get updated user: %w", err)
	}

	return &DailyCheckinResult{CheckedInToday: true, RewardAmount: reward, Balance: user.Balance, Checkin: item}, nil
}

func (s *DailyCheckinService) invalidateBalanceCaches(ctx context.Context, userID int64) {
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	if s.billingCacheService == nil {
		return
	}
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, userID)
	}()
}

func checkinDate(t time.Time) string {
	return t.Local().Format("2006-01-02")
}

func todayCompact(date string) string {
	if len(date) == len("2006-01-02") {
		return date[:4] + date[5:7] + date[8:10]
	}
	return date
}

func randomCheckinReward() (float64, error) {
	// 0.5 到 5.0，按分精度随机，避免浮点小数长尾。
	minCents := int64(50)
	maxCents := int64(500)
	n, err := rand.Int(rand.Reader, big.NewInt(maxCents-minCents+1))
	if err != nil {
		return 0, err
	}
	return float64(minCents+n.Int64()) / 100, nil
}
