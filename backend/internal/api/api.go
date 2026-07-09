package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/maquino96/timeline/internal/models"
	"github.com/maquino96/timeline/internal/scheduler"
	"github.com/maquino96/timeline/internal/store"
)

type API struct {
	store     *store.Store
	scheduler *scheduler.Scheduler
}

func New(s *store.Store, sc *scheduler.Scheduler) *API {
	return &API{store: s, scheduler: sc}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/items", a.handleGetItems)
	mux.HandleFunc("GET /api/sources", a.handleGetSources)
	mux.HandleFunc("POST /api/sources", a.handleCreateSource)
	mux.HandleFunc("PUT /api/sources/{id}", a.handleUpdateSource)
	mux.HandleFunc("DELETE /api/sources/{id}", a.handleDeleteSource)
	mux.HandleFunc("POST /api/sources/{id}/poll", a.handlePollSource)
	mux.HandleFunc("GET /api/topics", a.handleGetTopics)
	mux.HandleFunc("POST /api/topics", a.handleCreateTopic)
	mux.HandleFunc("PUT /api/topics/{id}", a.handleUpdateTopic)
	mux.HandleFunc("DELETE /api/topics/{id}", a.handleDeleteTopic)
}

func (a *API) handleGetItems(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)
	since := r.URL.Query().Get("since")
	sourceIDRaw := r.URL.Query().Get("source_id")
	sourceID := int64(queryInt(r, "source_id", 0))
	topicID := int64(queryInt(r, "topic_id", 0))
	sourceType := r.URL.Query().Get("source_type")
	search := r.URL.Query().Get("q")

	var sourceIDs []int64
	if strings.Contains(sourceIDRaw, ",") {
		for _, s := range strings.Split(sourceIDRaw, ",") {
			if id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil && id > 0 {
				sourceIDs = append(sourceIDs, id)
			}
		}
	}

	var items []models.Item
	var err error
	var total int

	if since != "" {
		switch {
		case search != "":
			items, err = a.store.SearchItemsSince(search, since, limit)
		case topicID > 0:
			items, err = a.store.GetItemsSinceByTopic(topicID, since, limit)
		case len(sourceIDs) > 1:
			items, err = a.store.GetItemsSinceBySources(sourceIDs, since, limit)
		case sourceID > 0:
			items, err = a.store.GetItemsSinceBySource(sourceID, since, limit)
		case sourceType != "":
			items, err = a.store.GetItemsSinceByType(sourceType, since, limit)
		default:
			items, err = a.store.GetItemsSince(since, limit)
		}
	} else {
		switch {
		case search != "":
			items, err = a.store.SearchItems(search, limit, offset)
		case topicID > 0:
			items, err = a.store.GetItemsByTopic(topicID, limit, offset)
		case len(sourceIDs) > 1:
			items, err = a.store.GetItemsBySources(sourceIDs, limit, offset)
		case sourceID > 0:
			items, err = a.store.GetItemsBySource(sourceID, limit, offset)
		case sourceType != "":
			items, err = a.store.GetItemsByType(sourceType, limit, offset)
		default:
			items, err = a.store.GetItems(limit, offset)
		}
	}

	switch {
	case search != "":
		total, _ = a.store.CountSearchItems(search)
	case topicID > 0:
		total, _ = a.store.CountItemsByTopic(topicID)
	case len(sourceIDs) > 1:
		total, _ = a.store.CountItemsBySources(sourceIDs)
	case sourceID > 0:
		total, _ = a.store.CountItemsBySource(sourceID)
	case sourceType != "":
		total, _ = a.store.CountItemsByType(sourceType)
	default:
		total, _ = a.store.CountItems()
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("X-Total-Count", fmt.Sprintf("%d", total))
	writeJSON(w, http.StatusOK, items)
}

func (a *API) handleGetSources(w http.ResponseWriter, r *http.Request) {
	sources, err := a.store.GetSources()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sources)
}

func (a *API) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var source models.Source
	if err := json.NewDecoder(r.Body).Decode(&source); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if source.Interval == 0 {
		source.Interval = 300
	}
	source.Enabled = true
	if err := a.store.AddSource(&source); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, source)
}

func (a *API) handleUpdateSource(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var source models.Source
	if err := json.NewDecoder(r.Body).Decode(&source); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	source.ID = id
	if err := a.store.UpdateSource(&source); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, source)
}

func (a *API) handleDeleteSource(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.store.DeleteSource(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handlePollSource(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.scheduler.PollSourceNow(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleGetTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := a.store.GetTopics()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, topics)
}

func (a *API) handleCreateTopic(w http.ResponseWriter, r *http.Request) {
	var topic models.Topic
	if err := json.NewDecoder(r.Body).Decode(&topic); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	topic.Enabled = true
	if err := a.store.AddTopic(&topic); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, topic)
}

func (a *API) handleUpdateTopic(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var topic models.Topic
	if err := json.NewDecoder(r.Body).Decode(&topic); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	topic.ID = id
	if err := a.store.UpdateTopic(&topic); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, topic)
}

func (a *API) handleDeleteTopic(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.store.DeleteTopic(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(zeroSlice(v))
}

func zeroSlice(v interface{}) interface{} {
	switch t := v.(type) {
	case []models.Item:
		if t == nil {
			return []models.Item{}
		}
	case []models.Source:
		if t == nil {
			return []models.Source{}
		}
	case []models.Topic:
		if t == nil {
			return []models.Topic{}
		}
	}
	return v
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func pathID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}
