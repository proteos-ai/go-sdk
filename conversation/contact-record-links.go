package conversation

import (
	"context"
	"net/http"

	conversationapi "go.proteos.ai/model/conversation/api"
	sdk "go.proteos.ai/sdk"
)

const contactRecordLinksBasePath = "/conversations/v1/contact-record-links"

// ContactRecordLinkService drives the contact ↔ record binding surface of
// conversation-service. Today it exposes the one call data-service makes
// inline on a duplicate_policy=reject write: Resolve.
type ContactRecordLinkService struct{ c *sdk.Client }

// Resolve binds each record of the batch to exactly ONE contact by its
// canonical contact-address values: mints or adopts the contact, maintains
// the record's source=record link, and reports per-record outcomes.
//
// A ONE-record batch is transparent: a duplicate_policy=reject collision is
// surfaced as *sdk.Error{HTTPStatus: 409, Code: "record_duplicate"} whose
// Details carry {entity_slug, contact_id, record_ids, address_keys}; a batch
// of many always answers 200 with per-record outcomes (rejected rows carry
// the same facts in their `duplicate` field).
func (s *ContactRecordLinkService) Resolve(ctx context.Context, req conversationapi.ResolveContactRecordLinksRequest) (conversationapi.ResolveContactRecordLinksResponse, error) {
	var out conversationapi.ResolveContactRecordLinksResponse
	err := s.c.Do(ctx, http.MethodPost, contactRecordLinksBasePath+"/resolve", req, &out)
	return out, err
}
