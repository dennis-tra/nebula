package monitor

import (
	"context"
	"fmt"

	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

// insertRawVisit builds up a raw_visit database entry.
func (s *Scheduler) insertRawVisit(ctx context.Context, cr Result) error {
	rv := &models.RawVisit{
		VisitStartedAt: cr.DialStartTime,
		VisitEndedAt:   cr.DialEndTime,
		DialDuration:   null.StringFrom(fmt.Sprintf("%f seconds", cr.DialDuration().Seconds())),
		Type:           models.VisitTypeDial,
		PeerMultiHash:  cr.Peer.ID.Pretty(),
		MultiAddresses: maddrsToAddrs(cr.Peer.Addrs),
	}
	if cr.Error != nil {
		rv.Error = null.StringFrom(cr.DialError)
		if len(cr.Error.Error()) > 255 {
			rv.ErrorMessage = null.StringFrom(cr.Error.Error()[:255])
		} else {
			rv.ErrorMessage = null.StringFrom(cr.Error.Error())
		}
	}

	return s.dbc.InsertRawVisit(ctx, rv)
}
