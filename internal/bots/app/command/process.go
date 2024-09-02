package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/bots/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/bots/domain/interfaces"
	"github.com/bmstu-itstech/itsreg-bots/internal/common/decorator"
)

type Process struct {
	BotUUID string
	UserID  int64
	Text    string
}

type ProcessHandler decorator.CommandHandler[Process]

type processHandler struct {
	bots         interfaces.BotsRepository
	participants interfaces.ParticipantRepository
	sender       interfaces.SenderService
}

func NewProcessHandler(
	bots interfaces.BotsRepository,
	participants interfaces.ParticipantRepository,
	sender interfaces.SenderService,

	logger *slog.Logger,
	metricsClient decorator.MetricsClient,
) ProcessHandler {
	if bots == nil {
		panic("bots repository is nil")
	}

	if participants == nil {
		panic("participants repository is nil")
	}

	if sender == nil {
		panic("sender service is nil")
	}

	return decorator.ApplyCommandDecorators[Process](
		processHandler{bots: bots, participants: participants, sender: sender},
		logger,
		metricsClient,
	)
}

func (h processHandler) Handle(ctx context.Context, cmd Process) error {
	bot, err := h.bots.Bot(ctx, cmd.BotUUID)
	if err != nil {
		return err
	}

	return h.participants.UpdateOrCreate(ctx, cmd.BotUUID, cmd.UserID, func(
		innerCtx context.Context, prt *bots.Participant,
	) error {
		messages, err := bot.Process(prt, cmd.Text)
		if err != nil {
			return err
		}

		for _, message := range messages {
			err = h.sender.Send(innerCtx, cmd.BotUUID, cmd.UserID, message)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
