package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"backend/internal/db"
)

func mustLocID(t *testing.T) interface{} {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test")
	}
	return locs[0]["id"]
}

func createEntry(t *testing.T, name string, locID interface{}) (interface{}, string) {
	payload := map[string]interface{}{
		"name":        name,
		"location_id": locID,
		"type":        "entry",
		"category":    "Egyéb",
	}
	rr := doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	slug, _ := created["slug"].(string)
	if slug == "" {
		t.Fatalf("POST entries missing slug, got: %v", created)
	}
	return created["id"], slug
}

func TestUnclaimedPublicEntryHidesExtras(t *testing.T) {
	payload := map[string]interface{}{
		"name":        "UnclaimedExtras A",
		"location_id": mustLocID(t),
		"type":        "entry",
		"category":    "Egyéb",
		"hours":       map[string]any{"mon": map[string]any{"open": "09:00", "close": "17:00", "closed": false}},
	}
	rr := doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	slug, _ := created["slug"].(string)
	if slug == "" {
		t.Fatal("POST entries missing slug")
	}
	id := created["id"]
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s (slug=%s, id=%v)", rr.Code, rr.Body.String(), slug, id)
	}
	var got map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["claimed"] == true {
		t.Fatal("expected unclaimed")
	}
	if _, ok := got["rating"]; ok {
		t.Fatal("unclaimed must omit rating")
	}
	if _, ok := got["reviews"]; ok {
		t.Fatal("unclaimed must omit reviews")
	}
	if got["ratings_enabled"] != false {
		t.Fatalf("ratings_enabled: %v", got["ratings_enabled"])
	}
	photos, _ := got["photos"].([]interface{})
	if len(photos) != 0 {
		t.Fatalf("unclaimed photos must be empty, got %v", got["photos"])
	}
}

func TestClaimedRatingsEnabledEmptyReviews(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ClaimedRatingsEmpty A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	ownerCookie := mustLogin("owner@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	entryIDInt := int(entryID.(float64))
	if _, err := db.DB.Exec(`UPDATE entries SET ratings_enabled = true WHERE id = $1`, entryIDInt); err != nil {
		t.Fatalf("enable ratings: %v", err)
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"reviews":[]`) {
		t.Fatalf("claimed with ratings on must include reviews:[], body: %s", body)
	}
	var got map[string]interface{}
	json.Unmarshal([]byte(body), &got)
	if got["claimed"] != true {
		t.Fatalf("expected claimed true, got %v", got["claimed"])
	}
	if got["ratings_enabled"] != true {
		t.Fatalf("expected ratings_enabled true, got %v", got["ratings_enabled"])
	}
	reviews, ok := got["reviews"].([]interface{})
	if !ok {
		t.Fatalf("reviews must be array, got %T", got["reviews"])
	}
	if len(reviews) != 0 {
		t.Fatalf("reviews must be empty, got len %d", len(reviews))
	}
}

func TestReviewUnsignedPOSTFails(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ReviewUnsigned A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	payload := map[string]interface{}{
		"slug":  slug,
		"score": 5,
		"text":  "Great place",
	}
	rr := doAnonRequest(t, "POST", "/api/entry/reviews", payload)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned POST: expected 401, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestReviewUnclaimedPublishedListingFails(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ReviewUnclaimed A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	userCookie := mustLogin("reviewer@test.lamsza")
	payload := map[string]interface{}{
		"slug":  slug,
		"score": 5,
		"text":  "Nice",
	}
	rr := doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, userCookie)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("unclaimed POST: expected 403, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestReviewClaimedButRatingsDisabledFails(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ReviewRatingsOff A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	ownerCookie := mustLogin("owner2@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	userCookie := mustLogin("reviewer2@test.lamsza")
	payload := map[string]interface{}{
		"slug":  slug,
		"score": 5,
		"text":  "Awesome",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, userCookie)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("ratings off POST: expected 403, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestReviewPostAndUpdateWorks(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ReviewPostUpdate A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	ownerCookie := mustLogin("owner3@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	entryIDInt := int(entryID.(float64))
	if _, err := db.DB.Exec(`UPDATE entries SET ratings_enabled = true WHERE id = $1`, entryIDInt); err != nil {
		t.Fatalf("enable ratings: %v", err)
	}

	userCookie := mustLogin("reviewer3@test.lamsza")
	payload := map[string]interface{}{
		"slug":  slug,
		"score": 5,
		"text":  "<b>hi</b>",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, userCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST review: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["review_count"] != float64(1) {
		t.Fatalf("review_count: expected 1, got %v", resp["review_count"])
	}
	if resp["rating"] != float64(5) {
		t.Fatalf("rating: expected 5, got %v", resp["rating"])
	}

	reviews, ok := resp["reviews"].([]interface{})
	if !ok || len(reviews) != 1 {
		t.Fatalf("reviews: expected array with 1 item, got %T len %v", resp["reviews"], len(reviews))
	}
	review := reviews[0].(map[string]interface{})
	if review["score"] != float64(5) {
		t.Fatalf("review score: expected 5, got %v", review["score"])
	}
	if review["text"] != "hi" {
		t.Fatalf("review text: expected 'hi' (sanitized), got %v", review["text"])
	}

	payload = map[string]interface{}{
		"slug":  slug,
		"score": 3,
		"text":  "Changed my mind",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, userCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST update: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["review_count"] != float64(1) {
		t.Fatalf("review_count after update: expected 1, got %v", resp["review_count"])
	}
	if resp["rating"] != float64(3) {
		t.Fatalf("rating after update: expected 3, got %v", resp["rating"])
	}
}

func TestReviewTwoUsersAverage(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ReviewTwoUsers A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	ownerCookie := mustLogin("owner4@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	entryIDInt := int(entryID.(float64))
	if _, err := db.DB.Exec(`UPDATE entries SET ratings_enabled = true WHERE id = $1`, entryIDInt); err != nil {
		t.Fatalf("enable ratings: %v", err)
	}

	user1Cookie := mustLogin("reviewer4a@test.lamsza")
	payload := map[string]interface{}{
		"slug":  slug,
		"score": 3,
		"text":  "OK",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, user1Cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST user1: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	user2Cookie := mustLogin("reviewer4b@test.lamsza")
	payload = map[string]interface{}{
		"slug":  slug,
		"score": 5,
		"text":  "Great",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, user2Cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST user2: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["review_count"] != float64(2) {
		t.Fatalf("review_count: expected 2, got %v", resp["review_count"])
	}
	if resp["rating"] != float64(4.0) {
		t.Fatalf("rating: expected 4.0, got %v", resp["rating"])
	}
}

func TestReviewDeleteWorks(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ReviewDelete A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	ownerCookie := mustLogin("owner5@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	entryIDInt := int(entryID.(float64))
	if _, err := db.DB.Exec(`UPDATE entries SET ratings_enabled = true WHERE id = $1`, entryIDInt); err != nil {
		t.Fatalf("enable ratings: %v", err)
	}

	user1Cookie := mustLogin("reviewer5a@test.lamsza")
	payload := map[string]interface{}{
		"slug":  slug,
		"score": 3,
		"text":  "Meh",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, user1Cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST user1: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	user2Cookie := mustLogin("reviewer5b@test.lamsza")
	payload = map[string]interface{}{
		"slug":  slug,
		"score": 5,
		"text":  "Excellent",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, user2Cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST user2: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequestWithCookie(t, "DELETE", "/api/entry/reviews?slug="+slug, nil, user1Cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["review_count"] != float64(1) {
		t.Fatalf("review_count after delete: expected 1, got %v", resp["review_count"])
	}
}

func TestReviewValidationErrors(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ReviewValidation A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	ownerCookie := mustLogin("owner6@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	entryIDInt := int(entryID.(float64))
	if _, err := db.DB.Exec(`UPDATE entries SET ratings_enabled = true WHERE id = $1`, entryIDInt); err != nil {
		t.Fatalf("enable ratings: %v", err)
	}

	userCookie := mustLogin("reviewer6@test.lamsza")

	payload := map[string]interface{}{
		"slug":  slug,
		"score": 0,
		"text":  "Bad score",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, userCookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("score 0: expected 400, got %d", rr.Code)
	}

	payload = map[string]interface{}{
		"slug":  slug,
		"score": 6,
		"text":  "Bad score",
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, userCookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("score 6: expected 400, got %d", rr.Code)
	}

	longText := strings.Repeat("a", 2001)
	payload = map[string]interface{}{
		"slug":  slug,
		"score": 5,
		"text":  longText,
	}
	rr = doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, userCookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("text too long: expected 400, got %d", rr.Code)
	}
}

func TestReviewUnpublishedListingStrangerFails(t *testing.T) {
	locID := mustLocID(t)
	entryID, slug := createEntry(t, "ReviewUnpublished A", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	// Make the entry unpublished
	entryIDInt := int(entryID.(float64))
	if _, err := db.DB.Exec(`UPDATE entries SET published = false WHERE id = $1`, entryIDInt); err != nil {
		t.Fatalf("unpublish entry: %v", err)
	}

	userCookie := mustLogin("stranger@test.lamsza")
	payload := map[string]interface{}{
		"slug":  slug,
		"score": 5,
		"text":  "Should fail",
	}
	rr := doRequestWithCookie(t, "POST", "/api/entry/reviews", payload, userCookie)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unpublished stranger POST: expected 404, got %d; body: %s", rr.Code, rr.Body.String())
	}
}
