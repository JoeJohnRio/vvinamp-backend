package youtube

import "testing"

func TestPaginateSearchResultsSkipsDuplicates(t *testing.T) {
	results := []SearchResult{
		{ID: "1", Title: "first"},
		{ID: "2", Title: "second"},
		{ID: "2", Title: "second duplicate"},
		{ID: "3", Title: "third"},
		{ID: "4", Title: "fourth"},
	}

	got := paginateSearchResults(results, 0, 2, map[string]struct{}{"2": {}})
	if len(got) != 2 {
		t.Fatalf("expected 2 unique results, got %d", len(got))
	}
	if got[0].ID != "1" || got[1].ID != "3" {
		t.Fatalf("expected page to contain IDs [1 3], got [%s %s]", got[0].ID, got[1].ID)
	}
}
