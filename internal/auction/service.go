package auction

import (
	"context"
	"log/slog"
	"time"

	"yabbi_test/internal/dsp"
	"yabbi_test/internal/partner"
)

type Service struct {
	store   partner.Store
	client  dsp.Client
	logger  *slog.Logger
	timeout time.Duration
}

func NewService(store partner.Store, client dsp.Client, logger *slog.Logger, timeout time.Duration) *Service {
	return &Service{
		store:   store,
		client:  client,
		logger:  logger,
		timeout: timeout,
	}
}

func (s *Service) RunAuction(ctx context.Context, req Request) Response {
	start := time.Now()

	filterResults := Filter(req, s.store.All())
	matched := Matched(filterResults)

	for _, r := range filterResults {
		if !r.Passed {
			s.logger.Debug("partner filtered out",
				"request_id", req.RequestID,
				"partner_uid", r.Partner.UID,
				"reason", r.Reason,
			)
		}
	}

	outcomes := Dispatch(ctx, s.client, matched, req, s.timeout)

	matchedUIDs := make([]string, 0, len(matched))
	for _, p := range matched {
		matchedUIDs = append(matchedUIDs, p.UID)
	}

	succeeded := 0
	for _, o := range outcomes {
		if o.Success {
			succeeded++
			continue
		}
		s.logger.Warn("bid failed",
			"request_id", req.RequestID,
			"partner_uid", o.PartnerUID,
			"error", o.Err,
		)
	}

	duration := time.Since(start)

	s.logger.Info("auction completed",
		"request_id", req.RequestID,
		"matched", len(matched),
		"sent", len(outcomes),
		"succeeded", succeeded,
		"duration_ms", duration.Milliseconds(),
	)

	return Response{
		RequestID:  req.RequestID,
		MatchedDSP: matchedUIDs,
		Sent:       len(outcomes),
		Succeeded:  succeeded,
		DurationMs: duration.Milliseconds(),
	}
}
