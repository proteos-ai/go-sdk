package data

import sdk "go.proteos.ai/sdk"

// Client groups the data services. Construct with New, then access the
// resource services via the public fields:
//
//	d := data.New(c)
//	rec, err := d.Records.Get(ctx, "customer", "r-1")
type Client struct {
	Records *RecordService
	Queries *QueryService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Records: &RecordService{c: c},
		Queries: &QueryService{c: c},
	}
}
