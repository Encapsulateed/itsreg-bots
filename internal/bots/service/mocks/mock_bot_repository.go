package mocks

import (
	"context"
	"sync"

	"github.com/bmstu-itstech/itsreg-bots/internal/bots/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/bots/domain/interfaces"
)

type mockBotRepository struct {
	sync.RWMutex
	m map[string]bots.Bot
}

func NewMockBotRepository() interfaces.BotsRepository {
	return &mockBotRepository{m: make(map[string]bots.Bot)}
}

func (r *mockBotRepository) Save(_ context.Context, bot *bots.Bot) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.m[bot.UUID]; ok {
		return interfaces.BotAlreadyExistsError{UUID: bot.UUID}
	}

	r.m[bot.UUID] = *bot

	return nil
}

func (r *mockBotRepository) Bot(_ context.Context, uuid string) (*bots.Bot, error) {
	r.RLock()
	defer r.RUnlock()

	b, ok := r.m[uuid]
	if !ok {
		return nil, interfaces.BotNotFoundError{UUID: uuid}
	}

	return &b, nil
}

func (r *mockBotRepository) Update(
	ctx context.Context,
	uuid string,
	updateFn func(context.Context, *bots.Bot) error,
) error {
	r.Lock()
	defer r.Unlock()

	b, ok := r.m[uuid]
	if !ok {
		return interfaces.BotNotFoundError{UUID: uuid}
	}

	err := updateFn(ctx, &b)
	if err != nil {
		return err
	}

	r.m[uuid] = b

	return nil
}

func (r *mockBotRepository) Delete(_ context.Context, uuid string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.m[uuid]; !ok {
		return interfaces.BotNotFoundError{UUID: uuid}
	}

	delete(r.m, uuid)

	return nil
}
