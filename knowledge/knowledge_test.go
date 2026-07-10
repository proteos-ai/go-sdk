package knowledge_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/knowledge"
	knowledgeapi "go.proteos.ai/model/knowledge/api"
)

func newClient(t *testing.T, handler http.HandlerFunc) *knowledge.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("t"))
	require.NoError(t, err)
	return knowledge.New(c)
}

func TestNodeService_Get_Path(t *testing.T) {
	k := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/knowledge/v1/nodes/n1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "n1", "title": "Alpha"})
	})
	got, err := k.Nodes.Get(context.Background(), "n1")
	require.NoError(t, err)
	require.Equal(t, "Alpha", got.Title)
}

func TestNodeService_GetContent_Path(t *testing.T) {
	k := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/knowledge/v1/nodes/n1/content", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "n1", "content": "body", "lines_total": 1})
	})
	got, err := k.Nodes.GetContent(context.Background(), "n1")
	require.NoError(t, err)
	require.Equal(t, "body", got.Content)
	require.Equal(t, 1, got.LinesTotal)
}

func TestNodeService_EditContent_Path(t *testing.T) {
	k := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/knowledge/v1/nodes/n1/content/actions/edit", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "n1", "replacements": 1, "lines_total": 2})
	})
	got, err := k.Nodes.EditContent(context.Background(), "n1", knowledgeapi.EditContentRequest{OldString: "a", NewString: "b"})
	require.NoError(t, err)
	require.Equal(t, 1, got.Replacements)
}

func TestNodeService_Search_Path(t *testing.T) {
	k := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/knowledge/v1/nodes/actions/search", r.URL.Path)
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":1,"pages_total":1},"data":[{"id":"n1","title":"Alpha","score":0.5}]}`))
	})
	got, err := k.Nodes.Search(context.Background(), knowledgeapi.SearchNodesRequest{Query: "alpha"})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	require.Equal(t, 0.5, got.Data[0].Score)
}

func TestNodeService_Neighbors_Path(t *testing.T) {
	var seenURL string
	k := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		_, _ = w.Write([]byte(`{"nodes":[{"id":"n2","title":"B","type":"markdown","status":"draft","distance":1}],"edges":[]}`))
	})
	got, err := k.Nodes.Neighbors(context.Background(), "n1", &knowledge.NeighborsOptions{Direction: "out", Depth: 2})
	require.NoError(t, err)
	require.Len(t, got.Nodes, 1)
	require.Equal(t, 1, got.Nodes[0].Distance)
	require.Contains(t, seenURL, "/knowledge/v1/nodes/n1/neighbors?")
	require.Contains(t, seenURL, "direction=out")
	require.Contains(t, seenURL, "depth=2")
}

func TestLinkService_Create_Path(t *testing.T) {
	k := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/knowledge/v1/links", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "l1", "from_id": "n1", "to_id": "n2", "type": "references"})
	})
	got, err := k.Links.Create(context.Background(), knowledgeapi.CreateKnowledgeLinkRequest{FromId: "n1", ToId: "n2", Type: "references"})
	require.NoError(t, err)
	require.Equal(t, "l1", got.Id)
}

func TestLabelService_ListPage_Path(t *testing.T) {
	var seenURL string
	k := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":100,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := k.Labels.ListPage(context.Background(), &knowledge.ListLabelsOptions{NameContains: "bill"})
	require.NoError(t, err)
	require.Contains(t, seenURL, "/knowledge/v1/labels?")
	require.Contains(t, seenURL, "name%5Bcontains%5D=bill")
}
